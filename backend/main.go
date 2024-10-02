package main

import (
	"log"
	"net/http"

	"github.com/jpfraneto/vibra/handlers"
	"github.com/jpfraneto/vibra/server"
	"github.com/rs/cors"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	log.Println("Starting Vibra application...")

	srv := server.NewServer()
	log.Println("Created new server instance")

	// Register API routes
	srv.Router.HandleFunc("/api/hello", handlers.HelloHandler).Methods("GET")
	srv.Router.HandleFunc("/api/present", handlers.PresentHandler).Methods("GET")
	srv.Router.HandleFunc("/api/cast", handlers.WriteCastHandler).Methods("POST")
	srv.Router.HandleFunc("/api/auth/login", handlers.LoginHandler).Methods("POST")
	srv.Router.HandleFunc("/api/auth/signer-status", handlers.SignerStatusHandler).Methods("GET")

	log.Println("Registered all routes")

	// Create a new CORS handler
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"http://localhost:3001"}, // Allow requests from Next.js dev server
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"*"},
	})

	// Wrap the router with the CORS middleware
	handler := c.Handler(srv.Router)

	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}