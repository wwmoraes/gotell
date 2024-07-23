package gotell

import (
	"context"
	"errors"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/log/global"
	"golang.org/x/sync/errgroup"
)

// ErrForceFlush happens when a provider fails to force flush
var ErrForceFlush = errors.New("failed to force flush provider")

// ForceFlusher represents values that supports forcing a flush process.
type ForceFlusher interface {
	ForceFlush(ctx context.Context) error
}

// ForceFlush flushes the default logger, metric and tracer providers. Each will
// run on a separate goroutine. It'll return the first error if any happens and
// cancel the other routines.
//
// It works on any provider that implements the ForceFlusher interface. The OTEL
// standard providers do, for instance. It'll ignore providers that don't.
func ForceFlush(ctx context.Context) error {
	group := errgroup.Group{}

	group.Go(func() error {
		err := tryForceFlush(ctx, global.GetLoggerProvider())
		if !errors.Is(err, errors.ErrUnsupported) {
			return err
		}

		return nil
	})

	group.Go(func() error {
		err := tryForceFlush(ctx, otel.GetMeterProvider())
		if !errors.Is(err, errors.ErrUnsupported) {
			return err
		}

		return nil
	})

	group.Go(func() error {
		err := tryForceFlush(ctx, otel.GetTracerProvider())
		if !errors.Is(err, errors.ErrUnsupported) {
			return err
		}

		return nil
	})

	err := group.Wait()
	if err != nil {
		return errors.Join(ErrForceFlush, err)
	}

	return nil
}

// tryForceFlush calls ForceFlush with the provided context if target supports
// it. Otherwise it returns errors.ErrUnsupported.
func tryForceFlush(ctx context.Context, target any) error {
	forceFlusher, ok := target.(ForceFlusher)
	if !ok {
		return errors.ErrUnsupported
	}

	//nolint:wrapcheck // passthrough, its up to the caller wrapping it
	return forceFlusher.ForceFlush(ctx)
}
