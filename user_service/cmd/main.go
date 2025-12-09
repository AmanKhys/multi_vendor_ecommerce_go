package main

import (
	"log"
	"net/http"
	"os"
	"user_service"

	"github.com/joho/godotenv"
)

func main() {
	// Try to load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Note: .env file not found, relying on environment variables")
	}

	// Initialize the service (database connection, etc.)
	user_service.Init()

	mux := http.NewServeMux()
	user_service.RegisterRoutes(mux)

	port := "7777"
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}

	log.Printf("Starting user_service on port %s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}
