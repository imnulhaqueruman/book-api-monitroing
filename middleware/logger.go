package middleware

import (
	"book-api/metrics"
	"log"
	"net/http"
	"strconv"
	"time"
)

// responseWriter wrapper to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Logger middleware logs HTTP requests and collects metrics
func Logger(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Track in-flight requests
		metrics.HttpRequestsInFlight.Inc()
		defer metrics.HttpRequestsInFlight.Dec()
		
		// Wrap response writer to capture status code
		wrapped := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK, // Default status
		}
		
		// Call the next handler
		next.ServeHTTP(wrapped, r)
		
		// Calculate duration
		duration := time.Since(start).Seconds()
		
		// Record metrics
		status := strconv.Itoa(wrapped.statusCode)
		metrics.HttpRequestsTotal.WithLabelValues(
			r.Method,
			r.URL.Path,
			status,
		).Inc()
		
		metrics.HttpRequestDuration.WithLabelValues(
			r.Method,
			r.URL.Path,
		).Observe(duration)
		
		// Log after the request is handled
		log.Printf(
			"%s %s %d %s",
			r.Method,
			r.RequestURI,
			wrapped.statusCode,
			time.Since(start),
		)
	}
}

// EnableCORS adds CORS headers to responses
func EnableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		
		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		next.ServeHTTP(w, r)
	}
}