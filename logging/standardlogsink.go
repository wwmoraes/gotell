package logging

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/go-logr/logr"
)

// StandardLogSink is an adatpter to use the standard log as a logr sink.
//
// It creates two loggers internally, one for the standard output and one for
// the standard error. It writes Info messages on the former and error ones on
// the latter.
type StandardLogSink struct {
	stdout *log.Logger
	stderr *log.Logger
	values map[any]any
}

// NewStandardLogSink initializes a StandardLogSink for use.
//
// It creates two standard log.Logger instances: one for information messages,
// sent to the standard output descriptor; and another for error messages, sent
// to the standard error descriptor.
func NewStandardLogSink() *StandardLogSink {
	return &StandardLogSink{
		stdout: log.New(os.Stdout, "", log.LstdFlags),
		stderr: log.New(os.Stderr, "", log.LstdFlags),
		values: map[any]any{},
	}
}

// Enabled always returns true.
func (*StandardLogSink) Enabled(_ int) bool {
	return true
}

// Error prints a log message on the standard error descriptor.
func (sink *StandardLogSink) Error(err error, msg string, keysAndValues ...any) {
	values := mergeMaps(sink.values, kv2Map(keysAndValues...))

	sink.stderr.Printf("%s: %s%s", err.Error(), msg, mapString(values))
}

// Info prints a log message on the standard output descriptor.
func (sink *StandardLogSink) Info(_ int, msg string, keysAndValues ...any) {
	values := mergeMaps(sink.values, kv2Map(keysAndValues...))

	sink.stdout.Printf("%s%s", msg, mapString(values))
}

// Init does nothing. It exists to satisfy the logr.LogSink interface.
func (*StandardLogSink) Init(_ logr.RuntimeInfo) {}

// WithValues returns a new StandardLogSink with additional key/value pairs.
func (sink *StandardLogSink) WithValues(keysAndValues ...any) logr.LogSink {
	return &StandardLogSink{
		stdout: log.New(sink.stdout.Writer(), sink.stdout.Prefix(), sink.stdout.Flags()),
		stderr: log.New(sink.stderr.Writer(), sink.stderr.Prefix(), sink.stderr.Flags()),
		values: mergeMaps(sink.values, kv2Map(keysAndValues...)),
	}
}

// WithName returns a new StandardLogSink with the specified name appended.
func (sink *StandardLogSink) WithName(name string) logr.LogSink {
	return &StandardLogSink{
		stdout: log.New(sink.stdout.Writer(), sink.stdout.Prefix()+name, sink.stderr.Flags()),
		stderr: log.New(sink.stderr.Writer(), sink.stderr.Prefix()+name, sink.stderr.Flags()),
		values: sink.values,
	}
}

func mergeMaps(maps ...map[any]any) map[any]any {
	var length int

	for _, entry := range maps {
		length += len(entry)
	}

	if length == 0 {
		return nil
	}

	values := make(map[any]any, length)

	for _, entry := range maps {
		for k, v := range entry {
			values[k] = v
		}
	}

	return values
}

func kv2Map(keysAndValues ...any) map[any]any {
	values := make(map[any]any, len(keysAndValues)/2)

	for i := 0; i+1 < len(keysAndValues); i += 2 {
		values[keysAndValues[i]] = keysAndValues[i+1]
	}

	return values
}

func mapString(kvs map[any]any) string {
	if len(kvs) == 0 {
		return ""
	}

	var result strings.Builder

	for k, v := range kvs {
		if k == nil || v == nil {
			continue
		}

		result.WriteString(fmt.Sprintf(" %v=%v", k, v))
	}

	return result.String()
}
