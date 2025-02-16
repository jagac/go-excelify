package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	RequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total HTTP requests processed, labeled by method and path",
		},
		[]string{"method", "path"},
	)

	RequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Histogram of request durations in seconds, labeled by method and path",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	RequestPayload = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_request_payload_bytes_total",
			Help: "Total size of incoming request payloads in bytes, labeled by method and path",
		},
		[]string{"method", "path"},
	)

	ResponsePayload = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_response_payload_bytes_total",
			Help: "Total size of outgoing response payloads in bytes, labeled by method and path",
		},
		[]string{"method", "path"},
	)
)
