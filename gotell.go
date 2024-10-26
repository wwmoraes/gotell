// Package gotell extends the OpenTelemetry SDK with helpful utilities and sane
// defaults.
package gotell

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/XSAM/otelsql"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	otelruntime "go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

const (
	// NAME contains the package name. It's used to separate all OpenTelemetry
	// instances from the upstream.
	NAME = "github.com/wwmoraes/gotell"
)

// Options contains initialization properties that users can customize
type Options struct {
	LogsExporter    sdklog.Exporter
	MetricsExporter sdkmetric.Exporter
	Propagator      propagation.TextMapPropagator
	TracesExporter  sdktrace.SpanExporter
}

// Option represents an object that can modify a set of initialization options
type Option interface {
	Apply(opts *Options) error
}

// OptionFn is a function that implements Option
type OptionFn func(opts *Options) error

// Apply executes the function to apply its changes
func (fn OptionFn) Apply(opts *Options) error {
	return fn(opts)
}

// NewOptions creates a new Options object, applies any Option to it, and
// ensures defaults apply before returning a valid object
func NewOptions(ctx context.Context, options ...Option) (*Options, error) {
	opts := &Options{} //nolint:exhaustruct // we set defaults below

	options = append(options,
		withDefaultPropagator(),
		withDefaultLogsExporter(ctx),
		withDefaultMetricsExporter(ctx),
		withDefaultTracesExporter(ctx),
	)

	var err error
	for _, option := range options {
		err = option.Apply(opts)
		if err != nil {
			return nil, fmt.Errorf("failed to apply option: %w", err)
		}
	}

	return opts, nil
}

// WithLogsExporter sets a custom logs exporter
func WithLogsExporter(exporter sdklog.Exporter) Option {
	return OptionFn(func(opts *Options) error {
		opts.LogsExporter = exporter

		return nil
	})
}

// WithMetricsExporter sets a custom metrics exporter
func WithMetricsExporter(exporter sdkmetric.Exporter) Option {
	return OptionFn(func(opts *Options) error {
		opts.MetricsExporter = exporter

		return nil
	})
}

// WithTracesExporter sets a custom span exporter
func WithTracesExporter(exporter sdktrace.SpanExporter) Option {
	return OptionFn(func(opts *Options) error {
		opts.TracesExporter = exporter

		return nil
	})
}

// WithPropagator sets a custom propagator
func WithPropagator(propagator propagation.TextMapPropagator) Option {
	return OptionFn(func(opts *Options) error {
		opts.Propagator = propagator

		return nil
	})
}

// Initialize sets up OpenTelemetry log, metric and trace providers. It provides
// sane and configurable defaults. Those use upstream SDK environment variables
// as much as possible.
//
// This function is re-entrant but not idempotent: multiple calls yields new
// pointers for defaults. This is rarely what you want. It covers the use-case
// of an application initiatizing instrumentation in the main body as early
// as possible.
//
// As for the SDK environment variables, there's a few caveats:
//
//   - use both W3C Trace Context and W3C Baggage (OTEL_PROPAGATORS isn't supported by the Golang SDK)
//   - uses OTLP exporters over gRPC (ignores OTEL_LOGS_EXPORTER, OTEL_METRICS_EXPORTER and OTEL_TRACES_EXPORTER)
//   - uses batch processors for spans and logs
//   - uses a periodic reader processor for metrics
//
// See https://opentelemetry.io/docs/specs/otel/configuration/sdk-environment-variables/
func Initialize(ctx context.Context, res *resource.Resource, options ...Option) error {
	opts, err := NewOptions(ctx, options...)
	if err != nil {
		return fmt.Errorf("failed to build options: %w", err)
	}

	otel.SetTextMapPropagator(opts.Propagator)

	res, err = mergeResources(res)
	if err != nil {
		return fmt.Errorf("failed to merge resources: %w", err)
	}

	otel.SetTracerProvider(sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithBatcher(opts.TracesExporter),
	))

	otel.SetMeterProvider(sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(opts.MetricsExporter)),
	))

	global.SetLoggerProvider(sdklog.NewLoggerProvider(
		sdklog.WithResource(res),
		sdklog.WithProcessor(sdklog.NewBatchProcessor(opts.LogsExporter)),
	))

	//nolint:wrapcheck // no need to bloat this one
	return otelruntime.Start(
		otelruntime.WithMinimumReadMemStatsInterval(time.Second),
	)
}

// Start creates a new span and a context containing its reference.
//
// It augments the upstream span with attributes from the caller function and
// a wrapper that offers utility methods.
//
//nolint:ireturn // same practice as upstream to protect internal data
func Start(ctx context.Context, opts ...trace.SpanStartOption) (context.Context, Span) {
	return tracerStart(
		ctx,
		Tracer(ctx),
		"",
		opts...,
	)
}

// StartNamed creates a new named span and a context containing its reference.
//
// It augments the upstream span with attributes from the caller function and
// a wrapper that offers utility methods.
//
//nolint:ireturn // same practice as upstream to protect internal data
func StartNamed(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, Span) {
	return tracerStart(
		ctx,
		Tracer(ctx),
		spanName,
		opts...,
	)
}

// SpanFromContext returns the current Span from ctx.
//
// If no Span is currently set in ctx an implementation of a Span that
// performs no operations is returned.
//
//nolint:ireturn // same practice as upstream to protect internal data
func SpanFromContext(ctx context.Context) Span {
	return &span{trace.SpanFromContext(ctx)}
}

// ContextLabeler is an idempotent way to retrieve a labeler from a context.
//
// It'll return the existing labeler or set an empty labeler if none is present.
func ContextLabeler(ctx context.Context) (context.Context, *otelhttp.Labeler) {
	labeler, found := otelhttp.LabelerFromContext(ctx)
	if !found {
		ctx = otelhttp.ContextWithLabeler(ctx, labeler)
	}

	return ctx, labeler
}

// OpenSQL wraps database/sql.Open to add metadata and instrumentation.
func OpenSQL(driverName, dataSourceName string) (*sql.DB, error) {
	attributes := otelsql.WithAttributes(
		attribute.String("db.system", driverName),
	)

	dbHandler, err := otelsql.Open(driverName, dataSourceName, attributes)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	err = otelsql.RegisterDBStatsMetrics(dbHandler, attributes)
	if err != nil {
		return nil, fmt.Errorf("failed to register database metrics: %w", err)
	}

	return dbHandler, nil
}

// Logger returns a new logger configured by gotell.
func Logger() log.Logger {
	return global.GetLoggerProvider().Logger(NAME)
}

// Meter returns a new meter configured by gotell.
func Meter() metric.Meter {
	return otel.GetMeterProvider().Meter(NAME)
}

// Tracer returns a new tracer configured by gotell.
func Tracer(ctx context.Context) trace.Tracer {
	if span := trace.SpanFromContext(ctx); span.SpanContext().IsValid() {
		return span.TracerProvider().Tracer(NAME)
	}

	return otel.GetTracerProvider().Tracer(NAME)
}

func tracerStart(
	ctx context.Context,
	tracer trace.Tracer,
	spanName string,
	opts ...trace.SpanStartOption,
) (context.Context, *span) {
	info := GetFunctionInfo(2)

	opts = append(opts, trace.WithAttributes(FunctionInfoAttributes(info)...))

	if s := trace.SpanContextFromContext(ctx); s.IsValid() && s.IsRemote() {
		opts = append(opts, trace.WithLinks(trace.Link{
			SpanContext: s,
			Attributes:  nil,
		}))
	}

	if spanName == "" {
		spanName = info.FunctionName
	}

	ctx, upstreamSpan := tracer.Start(ctx, spanName, opts...)

	return ctx, &span{upstreamSpan}
}

func mergeResources(res *resource.Resource) (*resource.Resource, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("failed to get hostname: %w", err)
	}

	base, err := resource.Merge(resource.Default(), resource.NewSchemaless(
		attribute.Int("process.parent_pid", os.Getppid()),
		attribute.Int("process.pid", os.Getpid()),
		attribute.String("host.arch", runtime.GOARCH),
		attribute.String("host.name", hostname),
		attribute.String("os.type", runtime.GOOS),
		attribute.String("process.command", os.Args[0]),
		attribute.StringSlice("process.command_args", os.Args),
	))
	if err != nil {
		return nil, fmt.Errorf("failed to merge base resources: %w", err)
	}

	res, err = resource.Merge(base, res)
	if err != nil {
		return nil, fmt.Errorf("failed to merge user resources: %w", err)
	}

	return res, nil
}

func withDefaultPropagator() Option {
	return OptionFn(func(opts *Options) error {
		if opts.Propagator != nil {
			return nil
		}

		opts.Propagator = propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		)

		return nil
	})
}

func withDefaultLogsExporter(ctx context.Context) Option {
	return OptionFn(func(opts *Options) error {
		if opts.LogsExporter != nil {
			return nil
		}

		logsExporter, err := otlploggrpc.New(ctx)
		if err != nil {
			return fmt.Errorf("failed to create an OTLP log exporter: %w", err)
		}

		opts.LogsExporter = logsExporter

		return nil
	})
}

func withDefaultMetricsExporter(ctx context.Context) Option {
	return OptionFn(func(opts *Options) error {
		if opts.MetricsExporter != nil {
			return nil
		}

		metricsExporter, err := otlpmetricgrpc.New(ctx)
		if err != nil {
			return fmt.Errorf("failed to create an OTLP metric exporter: %w", err)
		}

		opts.MetricsExporter = metricsExporter

		return nil
	})
}

func withDefaultTracesExporter(ctx context.Context) Option {
	return OptionFn(func(opts *Options) error {
		if opts.TracesExporter != nil {
			return nil
		}

		tracesExporter, err := otlptracegrpc.New(ctx)
		if err != nil {
			return fmt.Errorf("failed to create an OTLP trace exporter: %w", err)
		}

		opts.TracesExporter = tracesExporter

		return nil
	})
}
