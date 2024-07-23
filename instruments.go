package gotell

import (
	"sync"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

const (
	failedToInitializeMetricMessage = "failed to initialize metric"
	ucumBytes                       = "By"
	ucumSeconds                     = "s"
)

//nolint:gochecknoglobals // metrics are global, no way around it
var (
	httpClientRequestBodySize = sync.OnceValue(func() metric.Float64Histogram {
		instrument, err := Meter().Float64Histogram(
			"http.client.request.body.size",
			metric.WithDescription("Size of HTTP client request bodies."),
			metric.WithUnit(ucumBytes),
		)
		if err != nil {
			otel.Handle(err)
		}

		return instrument
	})

	httpClientResponseBodySize = sync.OnceValue(func() metric.Float64Histogram {
		instrument, err := Meter().Float64Histogram(
			"http.client.response.body.size",
			metric.WithDescription("Size of HTTP client response bodies."),
			metric.WithUnit(ucumBytes),
		)
		if err != nil {
			otel.Handle(err)
		}

		return instrument
	})

	httpClientRequestDuration = sync.OnceValue(func() metric.Float64Histogram {
		instrument, err := Meter().Float64Histogram(
			"http.client.request.duration",
			metric.WithDescription("Duration of HTTP client requests."),
			metric.WithUnit(ucumSeconds),
			//nolint:mnd // as per https://opentelemetry.io/docs/specs/semconv/http/http-metrics/
			metric.WithExplicitBucketBoundaries(0.005, 0.01, 0.025, 0.05, 0.075, 0.1, 0.25, 0.5, 0.75, 1, 2.5, 5, 7.5, 10),
		)
		if err != nil {
			otel.Handle(err)
		}

		return instrument
	})

	httpServerActiveRequestsInstrument = sync.OnceValue(func() metric.Int64Gauge {
		instrument, err := Meter().Int64Gauge(
			"http.server.active_requests",
			metric.WithDescription("Number of active HTTP server requests."),
			metric.WithUnit("{request}"),
		)
		if err != nil {
			otel.Handle(err)
		}

		return instrument
	})

	httpServerRequestBodySizeInstrument = sync.OnceValue(func() metric.Float64Histogram {
		instrument, err := Meter().Float64Histogram(
			"http.server.request.body.size",
			metric.WithDescription("Size of HTTP server request bodies."),
			metric.WithUnit(ucumBytes),
		)
		if err != nil {
			otel.Handle(err)
		}

		return instrument
	})

	httpServerRequestDurationInstrument = sync.OnceValue(func() metric.Float64Histogram {
		instrument, err := Meter().Float64Histogram(
			"http.server.request.duration",
			metric.WithDescription("Duration of HTTP server requests."),
			metric.WithUnit(ucumSeconds),
			//nolint:mnd // as per https://opentelemetry.io/docs/specs/semconv/http/http-metrics/
			metric.WithExplicitBucketBoundaries(0.005, 0.01, 0.025, 0.05, 0.075, 0.1, 0.25, 0.5, 0.75, 1, 2.5, 5, 7.5, 10),
		)
		if err != nil {
			otel.Handle(err)
		}

		return instrument
	})

	httpServerResponseBodySizeInstrument = sync.OnceValue(func() metric.Float64Histogram {
		instrument, err := Meter().Float64Histogram(
			"http.server.response.body.size",
			metric.WithDescription("Size of HTTP server response bodies."),
			metric.WithUnit(ucumBytes),
		)
		if err != nil {
			otel.Handle(err)
		}

		return instrument
	})
)
