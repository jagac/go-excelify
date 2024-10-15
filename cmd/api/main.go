package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/jagac/excelify/internal/server"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	portStr := os.Getenv("PORT")
	if portStr == "" {
		log.Fatal("PORT environment variable is not set")
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Fatalf("Invalid PORT value: %s", portStr)
	}

	serverAddress := fmt.Sprintf(":%d", port)
	srv := server.NewServer(serverAddress)

	if err := srv.Run(); err != nil {
		log.Fatal(err)
	}
}
