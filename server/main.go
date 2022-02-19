// package main is a server used for local development. It is not used in the
// production setup.
package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path"
	"runtime"
	"syscall"
	"time"
)

func main() {
	if err := realMain(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func realMain() error {
	ctx, done := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM)
	defer done()

	_, filename, _, _ := runtime.Caller(0)
	dir := path.Join(path.Dir(filename), "..", "public")

	fileSrv := http.FileServer(http.Dir(dir))
	srv := &http.Server{
		Addr:    ":8080",
		Handler: fileSrv,
	}

	errCh := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			select {
			case errCh <- err:
			default:
			}
		}
	}()

	fmt.Fprintf(os.Stderr, "server listening at http://127.0.0.1:8080\n")

	select {
	case <-ctx.Done():
	case err := <-errCh:
		return fmt.Errorf("server error: %w", err)
	}

	fmt.Fprintf(os.Stderr, "\nshutting down...\n")
	shutdownCtx, shutdownDone := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownDone()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("failed to shutdown server: %w", err)
	}

	return nil
}
