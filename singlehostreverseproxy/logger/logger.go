package logger

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

type ctxLoggerKey int

const ctxRequestLogger ctxLoggerKey = iota

func init() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))
}

// Get returns the process-wide slog logger.
func Get() *slog.Logger {
	return slog.Default()
}

// GetRequestLogger extracts the request scoped logger from the request.
// If no logger is found in the context, the default logger is returned.
func GetRequestLogger(r *http.Request) *slog.Logger {
	return GetLoggerFromContext(r.Context())
}

// GetLoggerFromContext extracts the request scoped logger from the context.
// If no logger is found in the context, the default logger is returned.
func GetLoggerFromContext(ctx context.Context) *slog.Logger {
	l, ok := ctx.Value(ctxRequestLogger).(*slog.Logger)
	if !ok {
		return slog.Default()
	}
	return l
}

func AddLoggerToContext(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, ctxRequestLogger, logger)
}

func RequestMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		path := r.URL.Path
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r)
		slog.Info("request finished",
			"duration", time.Since(start).Milliseconds(),
			"path", path,
			"method", r.Method,
			"status", ww.Status(),
			"bytes", ww.BytesWritten(),
			"referer", r.Header.Get("Referer"),
			"userAgent", r.Header.Get("User-Agent"),
		)
	})
}
