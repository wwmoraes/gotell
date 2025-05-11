package logging

import (
	"context"
	"log/slog"
	"path"
	"time"

	"go.opentelemetry.io/otel/log"
)

// OpenTelemetryHandler is a slog.Handler that emits OpenTelemetry log records.
//
// See https://pkg.go.dev/log/slog#Handler
type OpenTelemetryHandler struct {
	Logger log.Logger
	Group  string
	Attrs  []slog.Attr
}

// Enabled returns true if the OpenTelemetry logger emits records for this
// context and level.
func (handler *OpenTelemetryHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return handler.Logger.Enabled(ctx, log.EnabledParameters{
		Severity: slogLevel2otelSeverity(level),
	})
}

// Handle emits an OpenTelemetry log entry with the data from the record.
//
//nolint:gocritic // upstream slog.Handler interface uses pass-by-value ¯\_(ツ)_/¯
func (handler *OpenTelemetryHandler) Handle(ctx context.Context, rec slog.Record) error {
	otelRecord := log.Record{}

	otelRecord.SetTimestamp(time.Now())
	otelRecord.SetBody(log.StringValue(rec.Message))
	otelRecord.SetObservedTimestamp(rec.Time)
	otelRecord.SetSeverity(slogLevel2otelSeverity(rec.Level))

	keyValues := make([]log.KeyValue, 0, rec.NumAttrs())
	rec.Attrs(func(a slog.Attr) bool {
		keyValues = append(keyValues, log.KeyValue{
			Key:   a.Key,
			Value: log.StringValue(a.Value.String()),
		})

		return true
	})

	otelRecord.AddAttributes(keyValues...)

	handler.Logger.Emit(ctx, otelRecord)

	return nil
}

// WithAttrs returns a new instance with additional key/value attributes.
func (handler *OpenTelemetryHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := make([]slog.Attr, 0, len(handler.Attrs)+len(attrs))

	copy(newAttrs[0:], handler.Attrs)
	copy(newAttrs[len(handler.Attrs):], attrs)

	return &OpenTelemetryHandler{
		Logger: handler.Logger,
		Group:  handler.Group,
		Attrs:  newAttrs,
	}
}

// WithGroup returns a new instance with the specified group name appended.
func (handler *OpenTelemetryHandler) WithGroup(name string) slog.Handler {
	attrs := make([]slog.Attr, 0, len(handler.Attrs))

	copy(attrs, handler.Attrs)

	return &OpenTelemetryHandler{
		Logger: handler.Logger,
		Group:  path.Join(handler.Group, name),
		Attrs:  attrs,
	}
}
