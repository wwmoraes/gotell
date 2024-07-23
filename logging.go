package gotell

import (
	"context"
	"log/slog"

	"github.com/go-logr/logr"
	"go.opentelemetry.io/otel/log"

	"github.com/wwmoraes/gotell/logging"
)

// Logr creates an instrumented logr.Logger for the target context.
//
// It uses the existing logr from the context by combining the original sink
// and an OpenTelemetry sink. This allows the caller to configure defaults
// through the standard logr.NewContext.
//
// It'll return a standard logr.Logger with an OpenTelemetry sink if there's no
// default set.
func Logr(ctx context.Context) logr.Logger {
	sink := &logging.OpenTelemetryLogSink{
		Context:    ctx,
		Logger:     Logger(),
		Name:       "",
		Attributes: []log.KeyValue{},
	}

	logger, err := logr.FromContext(ctx)
	if err != nil {
		return logr.New(sink)
	}

	originalSink := logger.GetSink()

	if teeSink, ok := originalSink.(logging.TeeLogSink); ok {
		return logger.WithSink(append(teeSink, sink))
	}

	return logger.WithSink(logging.TeeLogSink{originalSink, sink})
}

// Slog creates an structured logger with an OpenTelemetry handler.
//
// It uses the handler of the default logger from the package, and concatenates
// with an OpenTelemetry one. This allows the caller to configure a base logger
// through the standard slog.SetDefault.
//
// Its the caller's responsibility to use the *Context methods to ensure logs
// retain traceability metadata.
func Slog() *slog.Logger {
	return slog.New(logging.TeeHandler{
		slog.Default().Handler(),
		&logging.OpenTelemetryHandler{
			Logger: Logger(),
			Group:  "",
			Attrs:  []slog.Attr{},
		},
	})
}
