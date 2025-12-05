package main

import (
	"book-api/database"
	"book-api/handlers"
	"book-api/middleware"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// Initialize database
	err := database.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.CloseDB()

	// Run migrations
	err = database.RunMigrations()
	if err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Start collecting database metrics every 15 seconds
	database.StartMetricsCollection(15 * time.Second)

	// Define routes for main API
	http.HandleFunc("/health",
		middleware.Logger(middleware.EnableCORS(handlers.HealthCheck)))

	http.HandleFunc("/books",
		middleware.Logger(middleware.EnableCORS(bookRouter)))

	http.HandleFunc("/books/",
		middleware.Logger(middleware.EnableCORS(bookRouter)))

	// Add metrics endpoint
	http.Handle("/metrics", promhttp.Handler())

	// Start main API server
	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}

	server := &http.Server{
		Addr:         ":" + port,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("API server starting on port %s...", port)
		log.Printf("Metrics available at https://localhost:%s/metrics", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("API server failed to start: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Server is shutting down...")
}

// bookRouter routes requests to appropriate handlers based on HTTP method
func bookRouter(w http.ResponseWriter, r *http.Request) {
	// Check if path has an ID
	path := strings.TrimPrefix(r.URL.Path, "/books")
	hasID := path != "" && path != "/"

	switch r.Method {
	case http.MethodGet:
		if hasID {
			handlers.GetBook(w, r)
		} else {
			handlers.GetBooks(w, r)
		}
	case http.MethodPost:
		if !hasID {
			handlers.CreateBook(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	case http.MethodPut:
		if hasID {
			handlers.UpdateBook(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	case http.MethodDelete:
		if hasID {
			handlers.DeleteBook(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}