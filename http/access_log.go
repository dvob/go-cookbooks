package main

import (
	"log/slog"
	"net/http"
	"time"
)

func logHandler(next http.Handler, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// TODO: hmmm?
		logHandler := logger.Handler().WithAttrs([]slog.Attr{
			slog.String("client", r.RemoteAddr),
			slog.String("method", r.Method),
			slog.String("uri", r.RequestURI),
			slog.Int64("content_length", r.ContentLength),
			slog.String("host", r.Host),
			slog.String("proto", r.Proto),
		})

		sw := newStatusResponseWriter(w)
		start := time.Now()
		next.ServeHTTP(sw, r)

		// TODO: hmmm?
		logger = slog.New(logHandler)

		if r.Context().Err() != nil {
			logger.LogAttrs(r.Context(), slog.LevelInfo, "access_log",
				slog.String("err", r.Context().Err().Error()),
				slog.Int("code", 499),
				slog.Duration("duration", time.Since(start)),
				slog.Int("bytes", sw.bytesWritten),
			)
		} else {
			logger.LogAttrs(r.Context(), slog.LevelInfo, "access_log",
				slog.Int("code", sw.statusCode),
				slog.Duration("duration", time.Since(start)),
				slog.Int("bytes", sw.bytesWritten),
			)
		}
	})
}

type requestLogValue struct {
	*http.Request
}

func (r requestLogValue) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("src", r.RemoteAddr),
		slog.String("proto", r.Proto),
		slog.String("method", r.Method),
		slog.String("host", r.Host),
		slog.String("uri", r.RequestURI),
	)
}

// statusResponseWriter to get bytes written and status code.
// inspired by https://www.alexedwards.net/blog/how-to-use-the-http-responsecontroller-type
type statusResponseWriter struct {
	http.ResponseWriter
	headerWritten bool
	statusCode    int
	bytesWritten  int
}

var _ http.ResponseWriter = (*statusResponseWriter)(nil)

func newStatusResponseWriter(w http.ResponseWriter) *statusResponseWriter {
	return &statusResponseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}
}

func (s *statusResponseWriter) Write(p []byte) (int, error) {
	n, err := s.ResponseWriter.Write(p)
	s.bytesWritten += n
	s.headerWritten = true
	return n, err
}

func (s *statusResponseWriter) Unwrap() http.ResponseWriter {
	return s.ResponseWriter
}

func (s *statusResponseWriter) WriteHeader(statusCode int) {
	s.ResponseWriter.WriteHeader(statusCode)
	if !s.headerWritten {
		s.statusCode = statusCode
		s.headerWritten = true
	}
}
