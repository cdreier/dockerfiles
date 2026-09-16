package metrics

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	requestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "reverse_proxy",
		Name:      "http_requests_total",
		Help:      "Count of HTTP requests handled by the reverse proxy.",
	}, []string{"code", "method"})

	requestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "reverse_proxy",
		Name:      "http_request_duration_seconds",
		Help:      "Duration of HTTP requests handled by the reverse proxy.",
		Buckets:   prometheus.DefBuckets,
	}, []string{"code", "method"})

	responseSize = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "reverse_proxy",
		Name:      "http_response_size_bytes",
		Help:      "Size of HTTP responses from the reverse proxy.",
		Buckets:   prometheus.ExponentialBuckets(100, 10, 7),
	}, []string{"code", "method"})

	inFlight = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "reverse_proxy",
		Name:      "http_requests_in_flight",
		Help:      "Number of HTTP requests currently being handled.",
	})

	proxyErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "reverse_proxy",
		Name:      "errors_total",
		Help:      "Reverse proxy errors by reason.",
	}, []string{"reason"})
)

// Handler returns a mux that only serves /metrics.
func Handler() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	return mux
}

// Start launches the scrape server on its own port. Port 0 disables it.
func Start(port int) {
	if port <= 0 {
		slog.Info("metrics server disabled", "port", port)
		return
	}

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
func Middleware(next http.Handler) http.Handler {
	return promhttp.InstrumentHandlerInFlight(inFlight,
		promhttp.InstrumentHandlerDuration(requestDuration,
			promhttp.InstrumentHandlerCounter(requestsTotal,
				promhttp.InstrumentHandlerResponseSize(responseSize, next),
			),
		),
	)
}

// IncProxyError increments the error counter for the given reason.
func IncProxyError(reason string) {
	proxyErrors.WithLabelValues(reason).Inc()
}
