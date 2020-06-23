package logger

import (
	"context"
	"net/http"
	"time"

	"go.uber.org/zap"
)

var _logger *zap.SugaredLogger

type ctxLoggerKey int

const ctxRequestLogger ctxLoggerKey = iota

func init() {
	loggerConfig := zap.NewProductionConfig()
	logger, _ := loggerConfig.Build()
	_logger = logger.Sugar()
}

// Get returns the current default SugaredLogger
func Get() *zap.SugaredLogger {
	return _logger
}

// GetRequestLogger extracts the request scoped logger from the request
// if no logger is found in the context, the default logger is returned
func GetRequestLogger(r *http.Request) *zap.SugaredLogger {
	return GetLoggerFromContext(r.Context())
}

// GetLoggerFromContext extracts the request scoped logger from the context
// if no logger is found in the context, the default logger is returned
func GetLoggerFromContext(ctx context.Context) *zap.SugaredLogger {
	l, ok := ctx.Value(ctxRequestLogger).(*zap.SugaredLogger)
	if !ok {
		return _logger
	}
	return l
}

func AddLoggerToContext(ctx context.Context, logger *zap.SugaredLogger) context.Context {
	return context.WithValue(ctx, ctxRequestLogger, logger)
}

func RequestMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		defer func() {
			diff := time.Since(start)
			_logger.Infow("request finished",
				"duration", diff.Milliseconds(),
				"path", r.URL.Path,
				"method", r.Method,
				"referer", r.Header.Get("referer"),
				"userAgent", r.Header.Get("user-agent"),
			)
		}()
		next.ServeHTTP(w, r)
	})
}
