package server

import (
	"log"
	"net/http"
	"time"

	"github.com/jagac/excelify/internal/services/converter"
	"github.com/jagac/excelify/internal/services/logging"
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
	logger, err := logging.NewLogger()
	if err != nil {
		log.Fatalf("could not initialize logger: %v", err)
	}
	converter := converter.NewConverter()
	handler := NewHandler(converter, logger)
	handler.RegisterRoutes(router)

	server := &http.Server{
		Addr:         s.addr,
		Handler:      router,
		ReadTimeout:  20 * time.Second,
		WriteTimeout: 20 * time.Second,
		IdleTimeout:  20 * time.Second,
	}
	return server.ListenAndServe()
}
