package server

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/jagac/excelify/internal/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Router struct {
	handler        *Handler
	logger         *slog.Logger
	logMiddleware  func(http.Handler) http.Handler
	corsMiddleware func(http.Handler) http.Handler
}

func NewRouter(handler *Handler, logger *slog.Logger) *Router {
	loggingConfig := middleware.LoggingConfig{Logger: logger}
	corsConfig := middleware.CORSConfig{}
	logMiddleware := loggingConfig.Middleware
	corsMiddleware := corsConfig.Middleware
	prometheus.MustRegister(requestsTotal, requestDuration, requestPayload, responsePayload)

	return &Router{
		handler:        handler,
		logger:         logger,
		logMiddleware:  logMiddleware,
		corsMiddleware: corsMiddleware,
	}
}

func (r *Router) RegisterRoutes(mux *http.ServeMux) {
	mux.Handle("/metrics", promhttp.Handler())
	mux.Handle("POST /api/v1/conversions/to-excel", MetricsMiddleware("POST /api/v1/conversions/to-excel", r.corsMiddleware(r.logMiddleware(http.HandlerFunc(r.handler.HandleJsonToExcel)))))
	mux.Handle("POST /api/v1/conversions/to-json", MetricsMiddleware("POST /api/v1/conversions/to-json", r.corsMiddleware(r.logMiddleware(http.HandlerFunc(r.handler.HandleExcelToJson)))))

}

var (
	requestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total HTTP requests processed, labeled by method and path",
		},
		[]string{"method", "path"},
	)

	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Histogram of request durations in seconds, labeled by method and path",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	requestPayload = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_request_payload_bytes_total",
			Help: "Total size of incoming request payloads in bytes, labeled by method and path",
		},
		[]string{"method", "path"},
	)

	responsePayload = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_response_payload_bytes_total",
			Help: "Total size of outgoing response payloads in bytes, labeled by method and path",
		},
		[]string{"method", "path"},
	)
)

func MetricsMiddleware(path string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		startTime := time.Now()

		reqSize := req.ContentLength
		if reqSize > 0 {
			requestPayload.WithLabelValues(req.Method, path).Add(float64(reqSize))
		}

		rec := &responseRecorder{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(rec, req)

		duration := time.Since(startTime).Seconds()
		requestDuration.WithLabelValues(req.Method, path).Observe(duration)
		respSize := rec.size
		responsePayload.WithLabelValues(req.Method, path).Add(float64(respSize))

		requestsTotal.WithLabelValues(req.Method, path).Inc()
	})
}

type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	size       int
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *responseRecorder) Write(data []byte) (int, error) {
	size, err := r.ResponseWriter.Write(data)
	r.size += size
	return size, err
}
