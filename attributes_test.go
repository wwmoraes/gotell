package gotell_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"

	"github.com/wwmoraes/gotell"
)

func TestRequestAttributes(t *testing.T) {
	t.Parallel()

	type args struct {
		req *http.Request
	}

	tests := []struct {
		name string
		args args
		want []attribute.KeyValue
	}{
		// TODO: Add test cases.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, gotell.RequestAttributes(tt.args.req))
		})
	}
}

func TestResponseWriterAttributes(t *testing.T) {
	t.Parallel()

	type args struct {
		headers    http.Header
		body       []byte
		statusCode int
	}

	tests := []struct {
		name string
		args args
		want []attribute.KeyValue
	}{
		{
			name: "empty",
			args: args{
				headers:    nil,
				body:       nil,
				statusCode: http.StatusOK,
			},
			want: []attribute.KeyValue{
				attribute.Int("http.response.status_code", http.StatusOK),
				attribute.Int("http.response.body.size", 0),
			},
		},
		{
			name: "body",
			args: args{
				headers:    nil,
				body:       []byte("foo"),
				statusCode: http.StatusOK,
			},
			want: []attribute.KeyValue{
				attribute.Int("http.response.status_code", http.StatusOK),
				attribute.Int("http.response.body.size", 3),
			},
		},
		{
			name: "error",
			args: args{
				headers:    nil,
				body:       []byte("foo"),
				statusCode: http.StatusInternalServerError,
			},
			want: []attribute.KeyValue{
				attribute.Int("http.response.status_code", http.StatusInternalServerError),
				attribute.Int("http.response.body.size", 3),
			},
		},
		{
			name: "custom headers",
			args: args{
				headers: http.Header{
					"X-Foo": []string{"bar"},
				},
				body:       nil,
				statusCode: http.StatusNoContent,
			},
			want: []attribute.KeyValue{
				attribute.StringSlice("http.response.x-foo", []string{"bar"}),
				attribute.Int("http.response.status_code", http.StatusNoContent),
				attribute.Int("http.response.body.size", 0),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			resWriter := gotell.NewResponseWriter(httptest.NewRecorder())

			for key, values := range tt.args.headers {
				for _, value := range values {
					resWriter.Header().Add(key, value)
				}
			}

			resWriter.WriteHeader(tt.args.statusCode)

			writtenBytes, err := resWriter.Write(tt.args.body)
			require.NoError(t, err)
			assert.Equal(t, len(tt.args.body), writtenBytes)

			assert.Equal(t, tt.want, gotell.ResponseWriterAttributes(resWriter))
		})
	}
}
