package server

import (
	"github.com/Jagac/excelify/internal/services"
	"log"
	"net/http"
	"time"
)

type Server struct {
	addr string
}

func NewServer(addr string) *Server {
	return &Server{
		addr: addr,
	}
}

func (s *Server) Run() error {
	router := http.NewServeMux()
	logger, err := services.NewLogger()
	if err != nil {
		log.Fatalf("could not initialize logger: %v", err)
	}
	converter := services.NewConverter()
	handler := NewHandler(converter, logger)
	handler.RegisterRoutes(router)

	server := &http.Server{
		Addr:         s.addr,
		Handler:      router,
		ReadTimeout:  120 * time.Second,
		WriteTimeout: 120 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	return server.ListenAndServe()
}
