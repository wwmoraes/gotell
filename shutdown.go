package gotell

import (
	"context"
	"errors"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/log/global"
	"golang.org/x/sync/errgroup"
)

// ErrShutdownFailed happens when a provider fails to shutdown
var ErrShutdownFailed = errors.New("failed to shutdown provider")

// Shutdowner represents values that support a graceful stop process, also known
// as shutdown.
type Shutdowner interface {
	Shutdown(ctx context.Context) error
}

// Shutdown shuts down the default logger, metric and tracer providers. Each
// will run on a separate goroutine. It'll return the first error if any happens
// and cancel the other routines.
//
// It works on any provider that implements the Shutdowner interface. The OTEL
// standard providers do, for instance. It'll ignore providers that don't.
func Shutdown(ctx context.Context) error {
	group, ctx := errgroup.WithContext(ctx)

	group.Go(func() error {
		err := tryShutdown(ctx, global.GetLoggerProvider())
		if !errors.Is(err, errors.ErrUnsupported) {
			return err
		}

		return nil
	})

	group.Go(func() error {
		err := tryShutdown(ctx, otel.GetMeterProvider())
		if !errors.Is(err, errors.ErrUnsupported) {
			return err
		}

		return nil
	})

	group.Go(func() error {
		err := tryShutdown(ctx, otel.GetTracerProvider())
		if !errors.Is(err, errors.ErrUnsupported) {
			return err
		}

		return nil
	})

	err := group.Wait()
	if err != nil {
		return errors.Join(ErrShutdownFailed, err)
	}

	return nil
}

// tryShutdown calls Shutdown with the provided context if target supports it.
// Otherwise it returns errors.ErrUnsupported.
func tryShutdown(ctx context.Context, target any) error {
	shutdowner, ok := target.(Shutdowner)
	if !ok {
		return errors.ErrUnsupported
	}

	//nolint:wrapcheck // internal function, no wrapping needed
	return shutdowner.Shutdown(ctx)
}
