package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"

	"net/http"
)

const envPrefix = "MY_APP_"

var version string

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	go func() {
		<-ctx.Done()
		// deregister signal handling to force exit with a second
		// interrupt (Ctrl-C)
		stop()
	}()
	if err := run(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	var (
		showVersion = false
		logLevel    = slog.LevelDebug

		tlsCert             string
		tlsKey              string
		shutdownGracePeriod = time.Minute
		server              = newDefaultServer()
	)

	flag.TextVar(&logLevel, "log-level", logLevel, "log level (DEBUG, INFO, WARN, ERROR)")
	flag.BoolVar(&showVersion, "version", showVersion, "print version and exit")

	flag.StringVar(&server.Addr, "addr", server.Addr, "server listen address")
	flag.StringVar(&tlsCert, "tls-cert", tlsCert, "tls certificate file")
	flag.StringVar(&tlsKey, "tls-key", tlsKey, "tls key file")
	flag.DurationVar(&shutdownGracePeriod, "shutdown-grace-period", shutdownGracePeriod, "shutdown grace period")
	flag.DurationVar(&server.WriteTimeout, "write-timeout", server.WriteTimeout, "server write timeout")
	flag.DurationVar(&server.ReadTimeout, "read-timeout", server.ReadTimeout, "server read timeout")
	flag.DurationVar(&server.IdleTimeout, "idle-timeout", server.IdleTimeout, "server idle timeout")

	err := readFlagsFromEnv(flag.CommandLine, envPrefix)
	if err != nil {
		return err
	}

	flag.Parse()

	if showVersion {
		if version != "" {
			fmt.Println(version)
			return nil
		}
		if buildInfo, ok := debug.ReadBuildInfo(); ok {
			fmt.Println(buildInfo.Main.Version)
			return nil
		}
		fmt.Println("(unknown)")
		return nil
	}

	logger := slog.New(newRequestIDLogger(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: logLevel})))
	slog.SetDefault(logger)

	// setup main handler
	var handler http.Handler
	handler = http.HandlerFunc(exampleAppHandler)

	// wrap main handler to add logging and request id
	handler = logHandler(handler, logger)
	handler = requestIDMiddleware(handler)

	server.Handler = handler

	runServer := server.ListenAndServe
	if tlsCert != "" && tlsKey != "" {
		runServer = func() error {
			return server.ListenAndServeTLS(tlsCert, tlsKey)
		}
	}

	errChan := make(chan error, 1)
	go func() {
		slog.InfoContext(ctx, "start server", "addr", server.Addr)
		errChan <- runServer()
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
	}

	shutdownCtx, cancelFn := context.WithTimeout(context.Background(), shutdownGracePeriod)
	defer cancelFn()
	slog.Info("shutdown server", "grace period", shutdownGracePeriod)
	return server.Shutdown(shutdownCtx)
}

func newDefaultServer() *http.Server {
	// https://blog.gopheracademy.com/advent-2016/exposing-go-on-the-internet/
	// potential upcoming public HTTP Server mode https://words.filippo.io/dispatches/go-1-21-plan/

	tlsConfig := &tls.Config{
		// Causes servers to use Go's default ciphersuite preferences,
		// which are tuned to avoid attacks. Does nothing on clients.
		PreferServerCipherSuites: true,
		// Only use curves which have assembly implementations
		CurvePreferences: []tls.CurveID{
			tls.CurveP256,
			tls.X25519, // Go 1.8 only
		},
		MinVersion: tls.VersionTLS12,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305, // Go 1.8 only
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,   // Go 1.8 only
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,

			// Best disabled, as they don't provide Forward Secrecy,
			// but might be necessary for some clients
			// tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
			// tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
		},
	}
	return &http.Server{
		Addr:         ":8080",
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
		TLSConfig:    tlsConfig,
	}
}

func exampleAppHandler(w http.ResponseWriter, r *http.Request) {
	// duration to simulate a long request
	if r.URL.Query().Has("duration") {
		duration, err := time.ParseDuration(r.URL.Query().Get("duration"))
		if err != nil {
			errorHandler(w, r, http.StatusBadRequest, err)
			return
		}
		slog.InfoContext(r.Context(), "process request", "duration", duration)

		delay := time.NewTimer(duration)
		select {
		case <-delay.C:
			// do something after one second.
		case <-r.Context().Done():
			// do something when context is finished and stop the timer.
			if !delay.Stop() {
				// if the timer has been stopped then read from the channel.
				<-delay.C
			}
		}
	}

	// error parameter to show error behaviour
	if r.URL.Query().Has("error") {
		errorHandler(w, r, 500, "this is a test error")
		return
	}

	n, err := fmt.Fprintln(w, "ok")
	slog.InfoContext(r.Context(), "outcome of write ok", "bytes", n, "err", err)
}
