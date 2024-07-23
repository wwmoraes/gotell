package gotell

import (
	"net/http"
)

var (
	_ http.ResponseWriter = (*responseWriter)(nil)
	_ ResponseWriter      = (*responseWriter)(nil)
)

// ResponseWriter extends a http.ResponseWriter to record the status code and
// content length.
type ResponseWriter interface {
	http.ResponseWriter
	ResponseStatusReporter
	ResponseContentLengthReporter

	// Unwrap returns the original proxied target.
	Unwrap() http.ResponseWriter
}

// ResponseStatusReporter is implemented by http.ResponseWriter values that
// allow retrieving the status code written to the client.
type ResponseStatusReporter interface {
	// Status returns the HTTP status of the request.
	Status() int
}

// ResponseContentLengthReporter is implemented by http.ResponseWriter values
// that allow retrieving the status code written to the client.
type ResponseContentLengthReporter interface {
	// ContentLength returns the total number of bytes sent to the client.
	ContentLength() int
}

type responseWriter struct {
	http.ResponseWriter

	headerWritten bool
	statusCode    int
	contentLength int
}

// NewResponseWriter wraps a http.ResponseWriter to record the status code and
// content length.
//
//nolint:ireturn // the underlying struct is private for safety purposes
func NewResponseWriter(w http.ResponseWriter) ResponseWriter {
	//nolint:exhaustruct // primitive zero values are safe
	return &responseWriter{ResponseWriter: w}
}

// Status returns the HTTP status of the request.
func (w *responseWriter) Status() int {
	return w.statusCode
}

// ContentLength returns the total number of bytes sent to the client.
func (w *responseWriter) ContentLength() int {
	return w.contentLength
}

// Unwrap returns the original proxied target.
func (w *responseWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

// WriteHeader stores the status code before calling the original method.
//
// It does nothing if the headers are written already.
func (w *responseWriter) WriteHeader(code int) {
	if w.headerWritten {
		return
	}

	w.statusCode = code
	w.headerWritten = true
	w.ResponseWriter.WriteHeader(code)
}

// Write writes data by calling the original method, then stores the content
// length.
//
// It respects the laws of the http.ResponseWriter by writing the headers first.
func (w *responseWriter) Write(data []byte) (int, error) {
	w.WriteHeader(http.StatusOK)

	n, err := w.ResponseWriter.Write(data)

	w.contentLength += n

	//nolint:wrapcheck // passthrough, no need to wrap
	return n, err
}
