package logging

import (
	"context"
	"log/slog"

	"golang.org/x/sync/errgroup"
)

// TeeHandler composes multiple slog.Handler instances, forwarding all function
// calls to each.
//
// See https://pkg.go.dev/log/slog#Handler
type TeeHandler []slog.Handler

// Enabled tests whether all underlying sinks enables the specified level.
// It returns false if any does not.
//
// See https://pkg.go.dev/log/slog#Handler
func (tee TeeHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, handler := range tee {
		if !handler.Enabled(ctx, level) {
			return false
		}
	}

	return true
}

// Handle handles the given record by forwarding it to all underlying handlers.
//
// See https://pkg.go.dev/log/slog#Handler
//
//nolint:gocritic // upstream slog.Handler interface uses pass-by-value ¯\_(ツ)_/¯
func (tee TeeHandler) Handle(ctx context.Context, record slog.Record) error {
	group := errgroup.Group{}

	for _, handler := range tee {
		group.Go(func() error {
			return handler.Handle(ctx, record)
		})
	}

	//nolint:wrapcheck // its the caller responsibility to handle this
	return group.Wait()
}

// WithAttrs returns a new TeeHandler containing new copies of the underlying
// handlers with additional key/value attributes.
//
// See https://pkg.go.dev/log/slog#Handler
func (tee TeeHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newTee := make(TeeHandler, 0, len(tee))

	for _, handler := range tee {
		newTee = append(newTee, handler.WithAttrs(attrs))
	}

	return newTee
}

// WithGroup returns a new TeeHandler containing new copies of the underlying
// handlers with the specified group name appended.
//
// See https://pkg.go.dev/log/slog#Handler
func (tee TeeHandler) WithGroup(name string) slog.Handler {
	newTee := make(TeeHandler, 0, len(tee))

	for _, handler := range tee {
		newTee = append(newTee, handler.WithGroup(name))
	}

	return newTee
}
