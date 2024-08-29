package middleware

import (
	"bytes"
	"log/slog"
	"net/http"
	"time"
)

type CORSConfig struct{}

type LoggingConfig struct {
	Logger *slog.Logger
}

type MiddlewareConfig struct {
	CORSConfig    CORSConfig
	LoggingConfig LoggingConfig
}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
}

func (c *CORSConfig) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		enableCors(&w)

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (l *LoggingConfig) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		l.Logger.Info("Incoming request",
			slog.String("method", r.Method),
			slog.String("url", r.URL.String()),
			slog.String("remote_addr", r.RemoteAddr),
			slog.Time("start_time", start))

		recorder := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK, body: &bytes.Buffer{}}

		next.ServeHTTP(recorder, r)

		duration := time.Since(start)
		l.Logger.Info("Request processed",
			slog.String("method", r.Method),
			slog.String("url", r.URL.String()),
			slog.String("remote_addr", r.RemoteAddr),
			slog.Duration("duration", duration),
			slog.Int("status", recorder.statusCode))

		if recorder.statusCode >= 400 {
			l.Logger.Error("Request resulted in an error",
				slog.String("method", r.Method),
				slog.String("url", r.URL.String()),
				slog.String("remote_addr", r.RemoteAddr),
				slog.Int("status", recorder.statusCode))

			if recorder.err != nil {
				l.Logger.Error("Error details", slog.String("error", recorder.err.Error()))
			}
		}
	})
}

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
	err        error
	body       *bytes.Buffer
}

func (rec *statusRecorder) WriteHeader(code int) {
	rec.statusCode = code
	rec.ResponseWriter.WriteHeader(code)
}

func (rec *statusRecorder) Write(b []byte) (int, error) {
	n, err := rec.ResponseWriter.Write(b)
	if err != nil {
		rec.err = err
	}
	rec.body.Write(b)
	return n, err
}
