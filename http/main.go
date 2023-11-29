package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"net/http"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	var (
		handler             http.Handler
		tlsCert             string
		tlsKey              string
		shutdownGracePeriod = time.Minute
		server              = newDefaultServer()
	)

	logger := slog.New(newRequestIDLogger(slog.NewTextHandler(os.Stderr, nil)))
	slog.SetDefault(logger)

	flag.StringVar(&server.Addr, "addr", server.Addr, "server listen address")
	flag.StringVar(&tlsCert, "tls-cert", tlsCert, "tls certificate file")
	flag.StringVar(&tlsKey, "tls-key", tlsKey, "tls key file")
	flag.DurationVar(&shutdownGracePeriod, "shutdown-grace-period", shutdownGracePeriod, "shutdown grace period")
	flag.Parse()

	handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// duration to simulate a long request
		if r.URL.Query().Has("duration") {
			duration, err := time.ParseDuration(r.URL.Query().Get("duration"))
			if err != nil {
				errorHandler(w, r, http.StatusBadRequest, err)
				return
			}
			slog.InfoContext(r.Context(), "process request", "duration", duration)
			time.Sleep(duration)
		}

		// error parameter to show error behaviour
		if r.URL.Query().Has("error") {
			errorHandler(w, r, 500, "this is a test error")
			return
		}

		fmt.Fprintln(w, "ok")
	})

	handler = logHandler(handler, logger)
	handler = requestIDMiddleware(handler)

	server.Handler = handler

	runServer := server.ListenAndServe
	if tlsCert != "" && tlsKey != "" {
		runServer = func() error {
			return server.ListenAndServeTLS(tlsCert, tlsKey)
		}
	}

	ctx, cancelFn := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancelFn()

	errChan := make(chan error, 1)
	go func() {
		slog.InfoContext(ctx, "start server", "addr", server.Addr)
		errChan <- runServer()
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		cancelFn()
	}

	shutdownCtx, cancelFn := context.WithTimeout(context.Background(), shutdownGracePeriod)
	defer cancelFn()
	slog.Info("shutdown server", "grace period", shutdownGracePeriod)
	return server.Shutdown(shutdownCtx)
}

func newDefaultServer() *http.Server {
	// https://blog.gopheracademy.com/advent-2016/exposing-go-on-the-internet/
	// potential upcoming public HTTP Server mode https://words.filippo.io/dispatches/go-1-21-plan/
	return &http.Server{
		Addr:         ":8080",
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
}
