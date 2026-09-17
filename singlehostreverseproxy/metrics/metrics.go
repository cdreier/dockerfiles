package metrics

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const unmatchedRoute = "other"

var (
	enabled bool
	routeRe *regexp.Regexp

	requestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "reverse_proxy",
		Name:      "http_requests_total",
		Help:      "Count of HTTP requests handled by the reverse proxy.",
	}, []string{"code", "method", "route"})

	requestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "reverse_proxy",
		Name:      "http_request_duration_seconds",
		Help:      "Duration of HTTP requests handled by the reverse proxy.",
		Buckets:   prometheus.DefBuckets,
	}, []string{"code", "method", "route"})

	responseSize = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "reverse_proxy",
		Name:      "http_response_size_bytes",
		Help:      "Size of HTTP responses from the reverse proxy.",
		Buckets:   prometheus.ExponentialBuckets(100, 10, 7),
	}, []string{"code", "method", "route"})

	inFlight = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "reverse_proxy",
		Name:      "http_requests_in_flight",
		Help:      "Number of HTTP requests currently being handled.",
	}, []string{"route"})

	proxyErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "reverse_proxy",
		Name:      "errors_total",
		Help:      "Reverse proxy errors by reason.",
	}, []string{"reason", "route"})
)

// SetRouteRegex compiles METRICS_ROUTE_REGEX.
// Unset or empty disables per-route labels: every request is "other".
// A capturing group becomes the label value; otherwise the full matching path is used.
//
// Examples:
//
//	# only exact paths
//	METRICS_ROUTE_REGEX='^/$|^/health$|^/favicon.ico$'
//	# / → "/", /health → "/health", /about → "other"
//
//	# SPA entry + static files, everything else collapsed
//	METRICS_ROUTE_REGEX='^/$|^/assets/.+'
//	# / → "/", /assets/app.js → "/assets/app.js", /users/42 → "other"
//
//	# capture group keeps cardinality bounded
//	METRICS_ROUTE_REGEX='^(/api/[^/]+)'
//	# /api/users/42/profile → "/api/users", /login → "other"
//
//	# several prefixes
//	METRICS_ROUTE_REGEX='^(/assets/|/api/|/health$)'
func SetRouteRegex(pattern string) error {
	pattern = strings.TrimSpace(pattern)
	if pattern == "" {
		routeRe = nil
		return nil
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}
	routeRe = re
	return nil
}

func routeLabel(path string) string {
	if routeRe == nil {
		return unmatchedRoute
	}
	m := routeRe.FindStringSubmatch(path)
	if len(m) == 0 {
		return unmatchedRoute
	}
	if len(m) > 1 && m[1] != "" {
		return m[1]
	}
	return path
}

// Handler returns a mux that only serves /metrics.
func Handler() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	return mux
}

// Start launches the scrape server on its own port.
// METRICS_PORT=0 disables scrape endpoint, request instrumentation and error counters.
func Start(port int) {
	if port <= 0 {
		enabled = false
		slog.Info("metrics disabled", "port", port)
		return
	}

	enabled = true
	addr := fmt.Sprintf(":%d", port)
	srv := &http.Server{
		Addr:              addr,
		Handler:           Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		slog.Info("prometheus metrics listening", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("metrics server stopped", "error", err)
			os.Exit(1)
		}
	}()
}

// Middleware records request count, duration, response size and in-flight gauge.
// The route label is taken from the original request path before index rewrite.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !enabled {
			next.ServeHTTP(w, r)
			return
		}
		route := routeLabel(r.URL.Path)
		labels := prometheus.Labels{"route": route}
		promhttp.InstrumentHandlerInFlight(inFlight.With(labels),
			promhttp.InstrumentHandlerDuration(requestDuration.MustCurryWith(labels),
				promhttp.InstrumentHandlerCounter(requestsTotal.MustCurryWith(labels),
					promhttp.InstrumentHandlerResponseSize(responseSize.MustCurryWith(labels), next),
				),
			),
		).ServeHTTP(w, r)
	})
}

// IncProxyError increments the error counter for the given reason and path.
func IncProxyError(reason, path string) {
	if !enabled {
		return
	}
	proxyErrors.WithLabelValues(reason, routeLabel(path)).Inc()
}
