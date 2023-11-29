package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
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
		handler http.Handler
		addr    = ":8080"
	)

	flag.StringVar(&addr, "addr", addr, "listen address")
	flag.Parse()

	handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok")
	})

	handler = logHandler(handler, slog.Default())

	// https://blog.gopheracademy.com/advent-2016/exposing-go-on-the-internet/
	// potential upcoming public HTTP Server mode https://words.filippo.io/dispatches/go-1-21-plan/
	srv := &http.Server{
		Addr:         addr,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
		Handler:      handler,
	}

	return srv.ListenAndServe()
}
