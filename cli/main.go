package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"
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
		// NOTE:
		// maybe look for signal cancelation and maybe treat it
		// different: https://github.com/golang/go/issues/60756

		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	var (
		showVersion = false
		logLevel    = slog.LevelDebug

		// app settings
		wait      time.Duration
		serverURL = "http://localhost:8080"
		runClient bool
	)

	flag.TextVar(&logLevel, "log-level", logLevel, "log level (DEBUG, INFO, WARN, ERROR)")
	flag.BoolVar(&showVersion, "version", showVersion, "print version and exit")

	flag.DurationVar(&wait, "wait", wait, "wait setting")
	flag.StringVar(&serverURL, "url", serverURL, "server url")
	flag.BoolVar(&runClient, "client", runClient, "run the client")

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

	// we explicitly set the default logger that it is set for the old log package as well
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: logLevel})))

	if runClient {
		// will log with slog
		log.Print("run client")
		return runAppClient(ctx, serverURL, wait)
	}
	log.Print("run dummy")
	return runAppDummy(ctx, wait)
}

func runAppClient(ctx context.Context, serverURL string, wait time.Duration) error {
	u, err := url.Parse(serverURL)
	if err != nil {
		return err
	}

	queryParams := u.Query()
	queryParams.Set("duration", wait.String())
	u.RawQuery = queryParams.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return err
	}

	slog.Debug("run get", "url", u)
	start := time.Now()
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	slog.Debug("request returned", "code", resp.StatusCode, "duration", time.Since(start))
	_, err = io.Copy(os.Stdout, resp.Body)
	return err
}

// runAppDummy ilustrates how defer runs properly (for example to remove temp
// files) when the context gets canceled (e.g. with Ctrl-C).
func runAppDummy(ctx context.Context, wait time.Duration) error {
	f, err := os.CreateTemp("", "")
	if err != nil {
		return err
	}
	defer func() {
		err := os.Remove(f.Name())
		if err != nil {
			slog.Error("failed to remove temp file", "file", f.Name(), "err", err)
		}
		slog.Debug("removed temp file", "file", f.Name())
	}()
	slog.Debug("created temp file", "file", f.Name())

	t := time.NewTimer(wait)
	select {
	case <-t.C:
		slog.Debug("time elapsed", "time", wait)
		return nil
	case <-ctx.Done():
		slog.Debug("context canceled")
		t.Stop()
		return ctx.Err()
	}
}
