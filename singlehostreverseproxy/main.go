package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/cdreier/dockerfiles/singlehostreverseproxy/logger"
	"github.com/cdreier/dockerfiles/singlehostreverseproxy/metrics"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	target := os.Getenv("TARGET")
	if strings.TrimSpace(target) == "" {
		slog.Error("TARGET env is required")
		os.Exit(1)
	}

	port := envInt("PORT", 8080)
	metricsPort := envInt("METRICS_PORT", 9102)

	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(metrics.Middleware)
	r.Use(logger.RequestMiddleware)
	r.Get("/*", getHandler(target))

	metrics.Start(metricsPort)

	addr := fmt.Sprintf(":%d", port)
	slog.Info("starting reverse proxy", "addr", addr, "target", target)
	if err := http.ListenAndServe(addr, r); err != nil {
		slog.Error("server stopped", "error", err)
		os.Exit(1)
	}
}

func envInt(key string, fallback int) int {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		slog.Error("invalid integer env", "key", key, "value", raw, "error", err)
		os.Exit(1)
	}
	return n
}

func getHandler(target string) http.HandlerFunc {
	upstream, err := url.Parse(target)
	if err != nil {
		slog.Error("unable to parse TARGET", "target", target, "error", err)
		os.Exit(1)
	}
	realServer := httputil.NewSingleHostReverseProxy(upstream)
	realServer.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		slog.Error("proxy error", "error", err, "path", r.URL.Path, "target", target)
		metrics.IncProxyError("upstream")
		w.WriteHeader(http.StatusBadGateway)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		// check for routing, append explicit file if none is set
		if filepath.Ext(r.URL.Path) == "" {
			if !strings.HasSuffix(r.URL.Path, "/") {
				r.URL.Path = r.URL.Path + "/"
			}
			r.URL.Path = r.URL.Path + "index.html"
		}

		// Update the headers to allow for SSL redirection
		r.URL.Host = upstream.Host
		r.URL.Scheme = upstream.Scheme
		r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
		r.Host = upstream.Host

		// realServer.ServeHTTP(&customResponseWriter{w}, r)
		realServer.ServeHTTP(w, r)
	}
}

type customResponseWriter struct {
	http.ResponseWriter
}

func (w *customResponseWriter) WriteHeader(code int) {
	w.ResponseWriter.WriteHeader(code)
	slog.Info("set code", "code", code)
	if code >= 400 {
		_, err := w.ResponseWriter.Write([]byte("oops"))
		if err != nil {
			slog.Error("write error body", "error", err)
		}
	}
}
