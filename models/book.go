package models

import "time"

// Book represents a book in our system
type Book struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Author    string    `json:"author"`
	ISBN      string    `json:"isbn"`
	Price     float64   `json:"price"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}