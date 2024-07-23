package gotell

import (
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
)

//nolint:gochecknoglobals // re-exports
var (
	// NewGRPCServerHandler creates a stats.Handler for a gRPC server.
	NewGRPCServerHandler = otelgrpc.NewServerHandler

	// NewGRPCClientHandler creates a stats.Handler for a gRPC client.
	NewGRPCClientHandler = otelgrpc.NewClientHandler
)
