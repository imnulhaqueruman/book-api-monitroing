package handlers

import (
	"book-api/database"
	"book-api/metrics"
	"book-api/models"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// HealthCheck handles the health check endpoint
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	// Check database connection
	db := database.GetDB()
	err := db.Ping()
	
	if err != nil {
		metrics.ApiErrorsTotal.WithLabelValues("database", "/health").Inc()
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{
			"status":   "unhealthy",
			"service":  "book-api",
			"database": "disconnected",
			"error":    err.Error(),
		})
		return
	}
	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":   "healthy",
		"service":  "book-api",
		"database": "connected",
	})
}

// GetBooks returns all books
func GetBooks(w http.ResponseWriter, r *http.Request) {
	db := database.GetDB()
	
	start := time.Now()
	rows, err := db.Query(`
		SELECT id, title, author, isbn, price, created_at, updated_at 
		FROM books 
		ORDER BY id DESC
	`)
	metrics.DbQueryDuration.WithLabelValues("select_all_books").Observe(time.Since(start).Seconds())
	
	if err != nil {
		log.Printf("Error querying books: %v", err)
		metrics.ApiErrorsTotal.WithLabelValues("database", "/books").Inc()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Error fetching books",
		})
		return
	}
	defer rows.Close()
	
	books := []models.Book{}
	
	for rows.Next() {
		var book models.Book
		err := rows.Scan(
			&book.ID,
			&book.Title,
			&book.Author,
			&book.ISBN,
			&book.Price,
			&book.CreatedAt,
			&book.UpdatedAt,
		)
		if err != nil {
			log.Printf("Error scanning book: %v", err)
			continue
		}
		books = append(books, book)
	}
	
	// Update total books gauge
	metrics.BooksTotal.Set(float64(len(books)))
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(books)
}

// GetBook returns a single book by ID
func GetBook(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	idStr := strings.TrimPrefix(r.URL.Path, "/books/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		metrics.ValidationErrorsTotal.Inc()
		metrics.ApiErrorsTotal.WithLabelValues("validation", "/books/{id}").Inc()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid book ID",
		})
		return
	}
	
	db := database.GetDB()
	
	var book models.Book
	start := time.Now()
	err = db.QueryRow(`
		SELECT id, title, author, isbn, price, created_at, updated_at 
		FROM books 
		WHERE id = $1
	`, id).Scan(
		&book.ID,
		&book.Title,
		&book.Author,
		&book.ISBN,
		&book.Price,
		&book.CreatedAt,
		&book.UpdatedAt,
	)
	metrics.DbQueryDuration.WithLabelValues("select_book_by_id").Observe(time.Since(start).Seconds())
	
	if err != nil {
		metrics.ApiErrorsTotal.WithLabelValues("not_found", "/books/{id}").Inc()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Book not found",
		})
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(book)
}

// CreateBook creates a new book
func CreateBook(w http.ResponseWriter, r *http.Request) {
	var book models.Book
	
	// Decode JSON request body
	err := json.NewDecoder(r.Body).Decode(&book)
	if err != nil {
		metrics.ValidationErrorsTotal.Inc()
		metrics.ApiErrorsTotal.WithLabelValues("validation", "/books").Inc()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid request body",
		})
		return
	}
	
	// Validate required fields
	if book.Title == "" || book.Author == "" {
		metrics.ValidationErrorsTotal.Inc()
		metrics.ApiErrorsTotal.WithLabelValues("validation", "/books").Inc()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Title and Author are required",
		})
		return
	}
	
	db := database.GetDB()
	
	// Insert book and return generated ID
	start := time.Now()
	err = db.QueryRow(`
		INSERT INTO books (title, author, isbn, price, created_at, updated_at)
		VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id, created_at, updated_at
	`, book.Title, book.Author, book.ISBN, book.Price).Scan(
		&book.ID,
		&book.CreatedAt,
		&book.UpdatedAt,
	)
	metrics.DbQueryDuration.WithLabelValues("insert_book").Observe(time.Since(start).Seconds())
	
	if err != nil {
		log.Printf("Error creating book: %v", err)
		metrics.ApiErrorsTotal.WithLabelValues("database", "/books").Inc()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Error creating book",
		})
		return
	}
	
	// Increment business metrics
	metrics.BooksCreatedTotal.Inc()
	metrics.BooksTotal.Inc()
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(book)
}

// UpdateBook updates an existing book
func UpdateBook(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	idStr := strings.TrimPrefix(r.URL.Path, "/books/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		metrics.ValidationErrorsTotal.Inc()
		metrics.ApiErrorsTotal.WithLabelValues("validation", "/books/{id}").Inc()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid book ID",
		})
		return
	}
	
	var book models.Book
	
	// Decode JSON request body
	err = json.NewDecoder(r.Body).Decode(&book)
	if err != nil {
		metrics.ValidationErrorsTotal.Inc()
		metrics.ApiErrorsTotal.WithLabelValues("validation", "/books/{id}").Inc()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid request body",
		})
		return
	}
	
	db := database.GetDB()
	
	// Update book
	start := time.Now()
	err = db.QueryRow(`
		UPDATE books 
		SET title = $1, author = $2, isbn = $3, price = $4, updated_at = CURRENT_TIMESTAMP
		WHERE id = $5
		RETURNING id, title, author, isbn, price, created_at, updated_at
	`, book.Title, book.Author, book.ISBN, book.Price, id).Scan(
		&book.ID,
		&book.Title,
		&book.Author,
		&book.ISBN,
		&book.Price,
		&book.CreatedAt,
		&book.UpdatedAt,
	)
	metrics.DbQueryDuration.WithLabelValues("update_book").Observe(time.Since(start).Seconds())
	
	if err != nil {
		metrics.ApiErrorsTotal.WithLabelValues("not_found", "/books/{id}").Inc()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Book not found",
		})
		return
	}
	
	// Increment business metrics
	metrics.BooksUpdatedTotal.Inc()
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(book)
}

// DeleteBook deletes a book
func DeleteBook(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	idStr := strings.TrimPrefix(r.URL.Path, "/books/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		metrics.ValidationErrorsTotal.Inc()
		metrics.ApiErrorsTotal.WithLabelValues("validation", "/books/{id}").Inc()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid book ID",
		})
		return
	}
	
	db := database.GetDB()
	
	// Delete book
	start := time.Now()
	result, err := db.Exec("DELETE FROM books WHERE id = $1", id)
	metrics.DbQueryDuration.WithLabelValues("delete_book").Observe(time.Since(start).Seconds())
	
	if err != nil {
		log.Printf("Error deleting book: %v", err)
		metrics.ApiErrorsTotal.WithLabelValues("database", "/books/{id}").Inc()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Error deleting book",
		})
		return
	}
	
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		metrics.ApiErrorsTotal.WithLabelValues("not_found", "/books/{id}").Inc()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Book not found",
		})
		return
	}
	
	// Increment business metrics
	metrics.BooksDeletedTotal.Inc()
	metrics.BooksTotal.Dec()
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Book deleted successfully",
	})
}