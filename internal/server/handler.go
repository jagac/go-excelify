package server

import (
	"encoding/json"
	"github.com/Jagac/excelify/internal/utils"
	"github.com/Jagac/excelify/types"
	"log/slog"
	"net/http"
)

type Handler struct {
	converter types.Converter
	logger    *slog.Logger
}

func NewHandler(converter types.Converter, logger *slog.Logger) *Handler {
	return &Handler{converter: converter, logger: logger}
}

func (h *Handler) RegisterRoutes(router *http.ServeMux) {
	router.HandleFunc("POST /api/v1/conversions", h.HandleJsonConversion)
}

func (h *Handler) HandleJsonConversion(w http.ResponseWriter, r *http.Request) {

	var jsonData types.RequestJson
	if err := json.NewDecoder(r.Body).Decode(&jsonData); err != nil {
		h.logger.Error("Failed to decode JSON",
			"error", err.Error(),
			"remote_addr", r.RemoteAddr,
			"url", r.URL.String())
		http.Error(w, "Cannot decode JSON", http.StatusBadRequest)
		return
	}

	if len(jsonData.Data) == 0 {
		h.logger.Error("No data provided",
			"remote_addr", r.RemoteAddr,
			"url", r.URL.String())
		http.Error(w, "No data provided", http.StatusBadRequest)
		return
	}

	excelBuffer, err := h.converter.ConvertToExcel(jsonData.Data, jsonData.Meta.Columns)
	if err != nil {
		h.logger.Error("Failed to convert to Excel",
			"error", err.Error(),
			"remote_addr", r.RemoteAddr,
			"url", r.URL.String())
		http.Error(w, "Failed to convert to Excel", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename="+jsonData.Filename)
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	if _, err := w.Write(excelBuffer.Bytes()); err != nil {
		h.logger.Error("Failed to write Excel response",
			"error", err.Error(),
			"remote_addr", r.RemoteAddr,
			"url", r.URL.String())
		http.Error(w, "Failed to write Excel response", http.StatusInternalServerError)
		return
	}

	h.logger.Info("success",
		slog.Group(
			"request",
			slog.String("ip", r.RemoteAddr),
			slog.String("method", r.Method),
			slog.String("url", r.RequestURI),
			slog.String("headers", utils.GetHeaders(r)),
			slog.Int("jsonSize", len(jsonData.Data)),
		))

}
