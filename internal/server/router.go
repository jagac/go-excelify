package server

import (
	"log/slog"
	"net/http"

	"github.com/jagac/excelify/internal/middleware"
)

type Router struct {
	handler       *Handler
	logger        *slog.Logger
	logMiddleware func(http.Handler) http.Handler
}

func NewRouter(handler *Handler, logger *slog.Logger) *Router {
	loggingConfig := middleware.LoggingConfig{Logger: logger}

	logMiddleware := loggingConfig.Middleware

	return &Router{
		handler:       handler,
		logger:        logger,
		logMiddleware: logMiddleware,
	}
}

func (r *Router) RegisterRoutes(mux *http.ServeMux) {
	mux.Handle("POST /api/v1/conversions/to-excel", r.logMiddleware(http.HandlerFunc(r.handler.HandleJsonToExcel)))
	mux.Handle("POST /api/v1/conversions/to-json", r.logMiddleware(http.HandlerFunc(r.handler.HandleExcelToJson)))

}
