package logging

import (
	"log/slog"

	"go.opentelemetry.io/otel/log"
)

func slogLevel2otelSeverity(level slog.Level) log.Severity {
	switch level {
	case slog.LevelDebug:
		return log.SeverityDebug
	case slog.LevelInfo:
		return log.SeverityInfo
	case slog.LevelWarn:
		return log.SeverityWarn
	case slog.LevelError:
		return log.SeverityError
	}

	return log.SeverityUndefined
}
