package app

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Service is an abstraction for an application with a clear lifecycle:
// - Run: main execution loop (blocking)
// - Shutdown: cleanup and graceful shutdown
type Service interface {
	Run(context.Context) error
	Shutdown(context.Context) error
}

// Runner coordinates the application lifecycle: receives context from main,
// listens for signals, and calls Run and Shutdown.
type Runner struct {
	Name string // optional, can be used for logging later
}

// Start runs the service with:
// - ctx from main (does not create Background context)
// - signal.Notify to handle SIGINT/SIGTERM
// - graceful shutdown with 5s timeout
func (r Runner) Start(ctx context.Context, svc Service) error {
	// Wrap the original context from main with signal handling
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)

	// Run svc.Run in a goroutine, so the main goroutine can wait for signal / ctx.Done
	go func() {
		if err := svc.Run(ctx); err != nil {
			errCh <- err
			return
		}
		close(errCh)
	}()

	var runErr error

	// Wait for one of two conditions:
	// 1) ctx is cancelled (signal / parent cancel)
	// 2) svc.Run returns an error
	select {
	case <-ctx.Done():
		// signal or parent cancel
	case err := <-errCh:
		// service terminated on its own
		runErr = err
	}

	// Whether it's a signal or service error, always attempt graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	shutdownErr := svc.Shutdown(shutdownCtx)

	// Prioritize returning runErr if it exists
	if runErr != nil {
		return runErr
	}

	return shutdownErr
}
