package server

import (
	"log/slog"
	"net/http"

	"github.com/jagac/excelify/internal/metrics"
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
	prometheus.MustRegister(metrics.RequestsTotal, metrics.RequestDuration, metrics.RequestPayload, metrics.ResponsePayload)

	return &Router{
		handler:        handler,
		logger:         logger,
		logMiddleware:  logMiddleware,
		corsMiddleware: corsMiddleware,
	}
}

func (r *Router) RegisterRoutes(mux *http.ServeMux) {
	mux.Handle("/metrics", promhttp.Handler())
	mux.Handle("POST /api/v1/conversions/to-excel", middleware.MetricsMiddleware("POST /api/v1/conversions/to-excel", r.corsMiddleware(r.logMiddleware(http.HandlerFunc(r.handler.HandleJsonToExcel)))))
	mux.Handle("POST /api/v1/conversions/to-json", middleware.MetricsMiddleware("POST /api/v1/conversions/to-json", r.corsMiddleware(r.logMiddleware(http.HandlerFunc(r.handler.HandleExcelToJson)))))

}
