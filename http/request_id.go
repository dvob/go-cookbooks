package main

// Adapted from chi RequestID middleware and inspired from opentelemetry trace IDs

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"log/slog"
	"net/http"
)

func requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := requestID{}
		_, err := rand.Read(id[:])
		if err != nil {
			slog.LogAttrs(r.Context(), slog.LevelError, "failed to generate request id", slog.Any("err", err))
		}
		ctx := context.WithValue(r.Context(), requestIDKey, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type requestID [16]byte

var zeroRequestID requestID

func (r requestID) IsZero() bool {
	return bytes.Equal(r[:], zeroRequestID[:])
}

func (r requestID) String() string {
	if r.IsZero() {
		return ""
	}
	return base64.RawURLEncoding.EncodeToString(r[:])
}

// Key to use when setting the request ID.
type ctxKeyRequestID int

// requestIDKey is the key that holds the unique request ID in a request context.
const requestIDKey ctxKeyRequestID = 0

// GetReqID returns a request ID from the given context if one is present.
// Returns the zero request id if no request id is set.
func getRequestID(ctx context.Context) requestID {
	if ctx == nil {
		return requestID{}
	}
	if reqID, ok := ctx.Value(requestIDKey).(requestID); ok {
		return reqID
	}
	return requestID{}
}
