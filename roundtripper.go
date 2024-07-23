package gotell

import "net/http"

// RoundTripperFn is a functor that implements http.RoundTripper
type RoundTripperFn func(req *http.Request) (*http.Response, error)

// RoundTrip executes an HTTP transaction.
//
// See https://pkg.go.dev/net/http#RoundTripper
func (fn RoundTripperFn) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}
