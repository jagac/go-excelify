package server

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/Jagac/excelify/internal/middleware"
	"github.com/Jagac/excelify/types"
	"github.com/xuri/excelize/v2"
)

type Handler struct {
	converter     types.Converter
	logger        *slog.Logger
	logMiddleware func(http.Handler) http.Handler
}

func NewHandler(converter types.Converter, logger *slog.Logger) *Handler {
	loggingConfig := middleware.LoggingConfig{Logger: logger}

	logMiddleware := loggingConfig.Middleware

	return &Handler{
		converter:     converter,
		logger:        logger,
		logMiddleware: logMiddleware,
	}
}

func (h *Handler) RegisterRoutes(router *http.ServeMux) {
	router.Handle("POST /api/v1/conversions/to-excel", h.logMiddleware(http.HandlerFunc(h.HandleJsonToExcel)))
	router.Handle("POST /api/v1/conversions/to-json", h.logMiddleware(http.HandlerFunc(h.HandleExcelToJson)))
}

func (h *Handler) HandleJsonToExcel(w http.ResponseWriter, r *http.Request) {

	var jsonData types.RequestJson
	if err := json.NewDecoder(r.Body).Decode(&jsonData); err != nil {
		http.Error(w, "Cannot decode JSON", http.StatusBadRequest)
		return
	}

	if len(jsonData.Data) == 0 {
		http.Error(w, "No data provided", http.StatusBadRequest)
		return
	}

	excelBuffer, err := h.converter.ConvertToExcel(jsonData.Data, jsonData.Meta.Columns)
	if err != nil {

		http.Error(w, "Failed to convert to Excel", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename="+jsonData.Filename)
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	if _, err := w.Write(excelBuffer.Bytes()); err != nil {

		http.Error(w, "Failed to write Excel response", http.StatusInternalServerError)
		return
	}

}

func (h *Handler) HandleExcelToJson(w http.ResponseWriter, r *http.Request) {
	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to read file from request", http.StatusBadRequest)
		return
	}
	defer file.Close()

	f, err := excelize.OpenReader(file)
	if err != nil {
		http.Error(w, "Failed to parse Excel file", http.StatusBadRequest)
		return
	}

	jsonData, err := h.converter.ConvertToJson(f)
	if err != nil {
		http.Error(w, "Failed to convert Excel to JSON", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(jsonData); err != nil {
		http.Error(w, "Failed to write json response", http.StatusInternalServerError)
	}

}
