package main

import (
	"context"
	"log/slog"
)

var _ slog.Handler = (*requestIDLogger)(nil)

type requestIDLogger struct {
	slog.Handler
}

func (h *requestIDLogger) Handle(ctx context.Context, r slog.Record) error {
	id := getRequestID(ctx)
	if !id.IsZero() {
		r.AddAttrs(slog.String("request_id", id.String()))
	}
	return h.Handler.Handle(ctx, r)
}

func newRequestIDLogger(h slog.Handler) slog.Handler {
	return &requestIDLogger{
		Handler: h,
	}
}
