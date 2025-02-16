package middleware

import (
	"net/http"
	"time"

	"github.com/jagac/excelify/internal/metrics"
)



func MetricsMiddleware(path string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		startTime := time.Now()

		reqSize := req.ContentLength
		if reqSize > 0 {
			metrics.RequestPayload.WithLabelValues(req.Method, path).Add(float64(reqSize))
		}

		rec := &responseRecorder{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(rec, req)

		duration := time.Since(startTime).Seconds()
		metrics.RequestDuration.WithLabelValues(req.Method, path).Observe(duration)
		respSize := rec.size
		metrics.ResponsePayload.WithLabelValues(req.Method, path).Add(float64(respSize))

		metrics.RequestsTotal.WithLabelValues(req.Method, path).Inc()
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