package gotell

import (
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// Span extends the trace.Span with convenience methods.
type Span interface {
	trace.Span

	// Assert changes the span status based on the error. It returns the
	// unmodified error.
	//
	// A non-nil error will record it and set the span status to codes.Error.
	// Conversely, a nil error will set the span status to Ok.
	Assert(err error) error
}

type span struct {
	trace.Span
}

//nolint:revive // unexported struct, no docs needed
func (s *span) Assert(err error) error {
	if err == nil {
		s.SetStatus(codes.Ok, "")
	} else {
		s.SetStatus(codes.Error, err.Error())
		s.RecordError(err)
	}

	return err
}
