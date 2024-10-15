package server

import (
	"log"
	"net/http"

	"github.com/jagac/excelify/internal/converter"
	"github.com/jagac/excelify/internal/logging"
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
	mux := http.NewServeMux()
	logger, err := logging.NewLogger()
	if err != nil {
		log.Fatalf("could not initialize logger: %v", err)
	}
	converter := converter.NewConverter()

	handler := NewHandler(converter)
	router := NewRouter(handler, logger)
	router.RegisterRoutes(mux)

	server := &http.Server{
		Addr:    s.addr,
		Handler: mux,
	}
	return server.ListenAndServe()
}
