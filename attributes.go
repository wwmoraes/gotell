package gotell

import (
	"net/http"
	"strconv"
	"strings"

	"go.opentelemetry.io/otel/attribute"
)

// RequestAttributes populates
func RequestAttributes(req *http.Request) []attribute.KeyValue {
	commonAttributes := commonRequestAttributes(req)
	serverAttributes := serverRequestAttributes(req)
	headerAttributes := requestHeaderAttributes(req)

	attrs := make([]attribute.KeyValue, 0, len(commonAttributes)+len(serverAttributes)+len(headerAttributes))

	attrs = append(attrs, commonAttributes[:]...)
	attrs = append(attrs, serverAttributes...)
	attrs = append(attrs, headerAttributes...)

	return attrs
}

// ResponseWriterAttributes derives observability attributes from a response
// writer.
//
// It checks if the writer implements optional interfaces to further enrich the
// attributes.
func ResponseWriterAttributes(w http.ResponseWriter) []attribute.KeyValue {
	attrs := make([]attribute.KeyValue, 0, 2+len(w.Header()))

	for key, values := range w.Header() {
		attrs = append(attrs, attribute.StringSlice("http.response."+strings.ToLower(key), values))
	}

	if ww, ok := w.(ResponseStatusReporter); ok {
		attrs = append(attrs, attribute.Int("http.response.status_code", ww.Status()))
	}

	if ww, ok := w.(ResponseContentLengthReporter); ok {
		attrs = append(attrs, attribute.Int("http.response.body.size", ww.ContentLength()))
	}

	return attrs
}

func commonRequestAttributes(req *http.Request) [15]attribute.KeyValue {
	protoName, protoVersion, _ := strings.Cut(strings.ToLower(req.Proto), "/")
	serverAddress, serverPort, _ := strings.Cut(req.Host, ":")

	return [...]attribute.KeyValue{
		// HTTP
		attribute.Int("http.request.body.size", int(req.ContentLength)),
		attribute.String("http.request.method", parseMethod(req.Method)),
		attribute.String("http.route", req.URL.Path),
		// Network
		attribute.String("network.protocol.name", protoName),
		attribute.String("network.protocol.version", protoVersion),
		// Server
		attribute.Int("server.port", parsePort(serverPort)),
		attribute.String("server.address", serverAddress),
		// URL
		attribute.String("url.domain", req.URL.Host),
		attribute.String("url.fragment", req.URL.Fragment),
		attribute.String("url.full", req.URL.String()),
		attribute.String("url.path", req.URL.Path),
		attribute.Int("url.port", parsePort(req.URL.Port())),
		attribute.String("url.query", req.URL.RawQuery),
		attribute.String("url.scheme", req.URL.Scheme),
		// Others
		attribute.String("user_agent.original", req.UserAgent()),
	}
}

func requestHeaderAttributes(req *http.Request) []attribute.KeyValue {
	attrs := make([]attribute.KeyValue, 0, len(req.Header))

	for key, values := range req.Header {
		attrs = append(attrs, attribute.StringSlice("http.request."+strings.ToLower(key), values))
	}

	return attrs
}

func serverRequestAttributes(req *http.Request) []attribute.KeyValue {
	if req.RemoteAddr == "" {
		return []attribute.KeyValue{}
	}

	clientAddress, clientPort, _ := strings.Cut(req.RemoteAddr, ":")

	return []attribute.KeyValue{
		attribute.String("client.address", clientAddress),
		attribute.Int("client.port", parsePort(clientPort)),
	}
}

func parsePort(value string) int {
	port, err := strconv.ParseInt("0"+value, 10, 32)
	if err != nil {
		return 0
	}

	return int(port)
}

func parseMethod(value string) string {
	if value == "" {
		return http.MethodGet
	}

	return strings.ToUpper(value)
}
