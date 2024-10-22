package main

import (
	"log"
	"net/http"
	"os"

	"github.com/jagac/excelify/internal/converter"
	"github.com/jagac/excelify/internal/logging"
	"github.com/jagac/excelify/internal/server"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	portStr := os.Getenv("PORT")
	address := ":" + portStr
	if portStr == "" {
		log.Fatal("PORT environment variable is not set")
	}

	mux := http.NewServeMux()
	logger, err := logging.NewLogger()
	if err != nil {
		log.Fatalf("could not initialize logger: %v", err)
	}
	converter := converter.NewConverter()

	handler := server.NewHandler(converter)
	router := server.NewRouter(handler, logger)
	router.RegisterRoutes(mux)

	if err := http.ListenAndServe(address, mux); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
