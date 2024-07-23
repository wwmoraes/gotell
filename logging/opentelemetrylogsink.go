package logging

import (
	"context"
	"fmt"
	"path"
	"time"

	"github.com/go-logr/logr"
	"go.opentelemetry.io/otel/log"
)

// OpenTelemetryLogSink is a logr.LogSink that emits OpenTelemetry log records.
//
// See https://pkg.go.dev/github.com/go-logr/logr#LogSink
type OpenTelemetryLogSink struct {
	//nolint:containedctx // logr does not receive a context value on its methods
	Context    context.Context
	Logger     log.Logger
	Name       string
	Attributes []log.KeyValue
}

// Enabled always returns true.
func (*OpenTelemetryLogSink) Enabled(_ int) bool {
	return true
}

// Error emits an OpenTelemetry log entry with an error severity.
func (sink *OpenTelemetryLogSink) Error(err error, msg string, keysAndValues ...any) {
	record := log.Record{}

	record.SetTimestamp(time.Now())
	record.SetBody(log.StringValue(fmt.Sprintf("%s: %s", msg, err.Error())))
	record.SetSeverity(log.SeverityError)
	record.AddAttributes(kv2akv(keysAndValues...)...)
	record.AddAttributes(log.String("log.name", sink.Name))

	sink.Logger.Emit(sink.Context, record)
}

// Info emits an OpenTelemetry log entry with an info severity.
func (sink *OpenTelemetryLogSink) Info(_ int, msg string, keysAndValues ...any) {
	record := log.Record{}

	record.SetTimestamp(time.Now())
	record.SetBody(log.StringValue(msg))
	record.SetSeverity(log.SeverityInfo)
	record.AddAttributes(kv2akv(keysAndValues...)...)
	record.AddAttributes(log.String("log.name", sink.Name))

	sink.Logger.Emit(sink.Context, record)
}

// Init does nothing. It exists to satisfy the logr.LogSink interface.
func (*OpenTelemetryLogSink) Init(_ logr.RuntimeInfo) {}

// WithName returns a new OpenTelemetrySink with the specified name appended.
func (sink *OpenTelemetryLogSink) WithName(name string) logr.LogSink {
	values := make([]log.KeyValue, 0, len(sink.Attributes))

	copy(values, sink.Attributes)

	return &OpenTelemetryLogSink{
		Context:    sink.Context,
		Logger:     sink.Logger,
		Name:       path.Join(sink.Name, name),
		Attributes: values,
	}
}

// WithValues returns a new OpenTelemetrySink with additional key/value pairs.
func (sink *OpenTelemetryLogSink) WithValues(keysAndValues ...any) logr.LogSink {
	newValues := kv2akv(keysAndValues...)

	values := make([]log.KeyValue, 0, len(sink.Attributes)+len(newValues))

	copy(values, sink.Attributes)
	copy(values[len(sink.Attributes):], newValues)

	return &OpenTelemetryLogSink{
		Context:    sink.Context,
		Logger:     sink.Logger,
		Name:       sink.Name,
		Attributes: values,
	}
}
