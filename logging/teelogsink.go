package logging

import "github.com/go-logr/logr"

// TeeLogSink composes multiple logr.LogSink instances, forwarding all function
// calls to each.
//
// See https://pkg.go.dev/github.com/go-logr/logr#LogSink
type TeeLogSink []logr.LogSink

// Enabled tests whether all underlying sinks enables the specified level.
// It returns false if any does not.
//
// See https://pkg.go.dev/github.com/go-logr/logr#LogSink
func (sinks TeeLogSink) Enabled(level int) bool {
	for _, sink := range sinks {
		if !sink.Enabled(level) {
			return false
		}
	}

	return true
}

// Error logs an error on all underlying sinks.
//
// See https://pkg.go.dev/github.com/go-logr/logr#LogSink
func (sinks TeeLogSink) Error(err error, msg string, keysAndValues ...any) {
	for _, sink := range sinks {
		sink.Error(err, msg, keysAndValues...)
	}
}

// Info logs a non-error message on all underlying sinks.
//
// See https://pkg.go.dev/github.com/go-logr/logr#LogSink
func (sinks TeeLogSink) Info(level int, msg string, keysAndValues ...any) {
	for _, sink := range sinks {
		sink.Info(level, msg, keysAndValues...)
	}
}

// Init receives and forwards optional information about the logr library to
// all underlying sinks.
//
// See https://pkg.go.dev/github.com/go-logr/logr#LogSink
func (sinks TeeLogSink) Init(info logr.RuntimeInfo) {
	info.CallDepth++

	for _, sink := range sinks {
		sink.Init(info)
	}
}

// WithValues returns a new TeeSink containing new copies of the underlying
// sinks with additional key/value pairs.
//
// See https://pkg.go.dev/github.com/go-logr/logr#LogSink
func (sinks TeeLogSink) WithValues(keysAndValues ...any) logr.LogSink {
	newSinks := make(TeeLogSink, 0, len(sinks))

	for _, sink := range sinks {
		newSinks = append(newSinks, sink.WithValues(keysAndValues...))
	}

	return newSinks
}

// WithName returns a new TeeSink containing new copies of the underlying sinks
// with the specified name appended.
//
// See https://pkg.go.dev/github.com/go-logr/logr#LogSink
func (sinks TeeLogSink) WithName(name string) logr.LogSink {
	newSinks := make(TeeLogSink, 0, len(sinks))

	for _, sink := range sinks {
		newSinks = append(newSinks, sink.WithName(name))
	}

	return newSinks
}
