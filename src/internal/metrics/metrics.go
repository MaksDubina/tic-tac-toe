package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Metrics struct {
	HTTPRequestsTotal   *prometheus.CounterVec
	HTTPRequestDuration *prometheus.HistogramVec
	GamesCreatedTotal   prometheus.Counter
	GamesActive         prometheus.Gauge
	MovesTotal          prometheus.Counter
	AuthAttemptsTotal   *prometheus.CounterVec
}

func New() *Metrics {
	return &Metrics{
		HTTPRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "path", "status"},
		),
		HTTPRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "Duration of HTTP requests in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "path"},
		),
		GamesCreatedTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "games_created_total",
				Help: "Total number of created games",
			},
		),
		GamesActive: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "games_active",
				Help: "Number of currently active games",
			},
		),
		MovesTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "moves_total",
				Help: "Total number of moves made",
			},
		),
		AuthAttemptsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "auth_attempts_total",
				Help: "Total number of auth attempts",
			},
			[]string{"type", "result"},
		),
	}
}
