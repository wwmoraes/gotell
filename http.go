package gotell

import (
	"context"
	"net/http"
	"slices"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/exp/constraints"
)

// StatusText represents an HTTP status code text as per IANA definitions and
// related RFCs.
//
// See https://www.iana.org/assignments/http-status-codes/http-status-codes.xhtml
type StatusText = string

// Number is a constraint that permits any numeric type.
type Number interface {
	constraints.Integer | constraints.Float
}

// Counter represents an additive metric counter.
type Counter[T Number] interface {
	Add(ctx context.Context, incr T, options ...metric.AddOption)
}

// Histogram represents a histogram metric counter.
type Histogram[T Number] interface {
	Record(ctx context.Context, incr T, options ...metric.RecordOption)
}

// MiddlewareFn is an HTTP handler identity functor, also known as a decorator.
type MiddlewareFn func(http.Handler) http.Handler

// MetricHandlerFn allows custom logic on how to calculate a metric value during
// a request. It provides the next handler, the request and the response writer.
// This handler must call the handler exactly once or else no response will be
// sent.
type MetricHandlerFn[T any] func(http.Handler, http.ResponseWriter, *http.Request) T

// MiddlewareWrapperFn is a post-response middleware that receives a wrapped
// response writer capable of reporting status and written bytes length.
type MiddlewareWrapperFn func(ctx context.Context, w ResponseWriter, handle func())

// SpanNameFormatter builds a HTTP span name using the method as prefix and
// operation. It fallsback to the request line (as per RFC 7230 section 3.1.1.)
// if operation is empty.
func SpanNameFormatter(operation string, req *http.Request) string {
	if req == nil {
		return operation
	}

	if operation != "" {
		return req.Method + " " + operation
	}

	// server-side requests
	if req.RequestURI != "" {
		return req.RequestURI
	}

	// client-side requests
	return req.Method + " " + req.URL.Path + " " + req.Proto
}

// WithInstrumentationMiddleware creates and enriches a span with HTTP request
// and response attributes.
//
// Some frameworks provide their own middleware implementation that does roughly
// the same. You should use either the framework-specific or this one instead of
// both to reduce the overhead per handler.
func WithInstrumentationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))

		ctx, labeler := ContextLabeler(ctx)

		reqAttributes := RequestAttributes(r)

		ctx, span := StartNamed(
			ctx,
			SpanNameFormatter("", r),
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(labeler.Get()...),
			trace.WithAttributes(reqAttributes...),
		)
		defer span.End()

		httpServerActiveRequestsInstrument().Record(ctx, 1)

		resWriter := NewResponseWriter(w)

		start := time.Now()

		next.ServeHTTP(resWriter, r.WithContext(ctx))

		end := time.Since(start)

		resAttributes := ResponseWriterAttributes(resWriter)
		span.SetAttributes(resAttributes...)
		span.SetStatus(SpanStatusForStatusCode(resWriter.Status()))

		instrumentAttributeSet := attribute.NewSet(slices.Concat(
			reqAttributes,
			resAttributes,
		)...)

		httpServerRequestBodySizeInstrument().Record(
			ctx,
			float64(r.ContentLength),
			metric.WithAttributeSet(instrumentAttributeSet),
		)
		httpServerRequestDurationInstrument().Record(
			ctx,
			end.Seconds(),
			metric.WithAttributeSet(instrumentAttributeSet),
		)
		httpServerResponseBodySizeInstrument().Record(
			ctx,
			float64(resWriter.ContentLength()),
			metric.WithAttributeSet(instrumentAttributeSet),
		)
		httpServerActiveRequestsInstrument().Record(ctx, -1, metric.WithAttributeSet(instrumentAttributeSet))
	})
}

// WithCounterMiddleware wraps another handler, incrementing the counter on each
// call.
func WithCounterMiddleware[T Number](counter Counter[T], incr T) MiddlewareFn {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)

			counter.Add(r.Context(), incr)
		})
	}
}

// WithHistogramMiddleware wraps another handler, recording a histogram
// increment on each call.
func WithHistogramMiddleware[T Number](
	histogram Histogram[T],
	handler MetricHandlerFn[T],
) MiddlewareFn {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			increment := handler(next, w, r)

			histogram.Record(r.Context(), increment)
		})
	}
}

// WithMetrics instruments another HTTP round tripper with request and response
// metrics.
func WithMetrics(next http.RoundTripper) http.RoundTripper {
	return RoundTripperFn(func(req *http.Request) (*http.Response, error) {
		// we just pretend we didn't see anything...
		if req == nil {
			return next.RoundTrip(req)
		}

		start := time.Now()
		res, err := next.RoundTrip(req)
		end := time.Since(start)

		ctx, labeler := ContextLabeler(req.Context())

		attrs := metric.WithAttributes(labeler.Get()...)
		httpClientRequestDuration().Record(ctx, end.Seconds(), attrs)
		httpClientRequestBodySize().Record(ctx, float64(req.ContentLength), attrs)

		var contentLength int64

		// response may be nil, such as on context cancellation
		if res != nil {
			contentLength = res.ContentLength
		}

		httpClientResponseBodySize().Record(ctx, float64(contentLength), attrs)

		//nolint:wrapcheck // passthrough
		return res, err
	})
}

// FilterHeaders returns a subset of headers for the given keys.
// The values are not copies.
func FilterHeaders(headers http.Header, includedKeys ...string) http.Header {
	target := make(http.Header, len(includedKeys))

	var values []string
	for _, key := range includedKeys {
		values = headers.Values(key)
		if values == nil {
			continue
		}

		target[key] = values
	}

	return target
}

// SpanStatusForStatusCode assigns a span code and message for a HTTP status
// code.
func SpanStatusForStatusCode(statusCode int) (codes.Code, StatusText) {
	code := codes.Unset

	if statusCode >= http.StatusBadRequest {
		code = codes.Error
	} else if statusCode >= http.StatusOK {
		code = codes.Ok
	}

	return code, http.StatusText(statusCode)
}
