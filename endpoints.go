package gotell

import (
	"os"
)

const (
	otelExporterOTLPEndpoint = "OTEL_EXPORTER_OTLP_ENDPOINT"
)

func getLogsEndpoint() string {
	endpoint, ok := os.LookupEnv("OTEL_EXPORTER_OTLP_LOGS_ENDPOINT")
	if ok && endpoint != "" {
		return endpoint
	}

	return os.Getenv(otelExporterOTLPEndpoint)
}

func getMetricEndpoint() string {
	endpoint, ok := os.LookupEnv("OTEL_EXPORTER_OTLP_METRICS_ENDPOINT")
	if ok && endpoint != "" {
		return endpoint
	}

	return os.Getenv(otelExporterOTLPEndpoint)
}

func getTracesEndpoint() string {
	endpoint, ok := os.LookupEnv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT")
	if ok && endpoint != "" {
		return endpoint
	}

	return os.Getenv(otelExporterOTLPEndpoint)
}
