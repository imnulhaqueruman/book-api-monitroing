package database

import "log"

// RunMigrations creates the necessary database tables
func RunMigrations() error {
	createTableQuery := `
	CREATE TABLE IF NOT EXISTS books (
		id SERIAL PRIMARY KEY,
		title VARCHAR(255) NOT NULL,
		author VARCHAR(255) NOT NULL,
		isbn VARCHAR(20),
		price DECIMAL(10, 2),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_books_title ON books(title);
	CREATE INDEX IF NOT EXISTS idx_books_author ON books(author);
	`

	_, err := DB.Exec(createTableQuery)
	if err != nil {
		return err
	}

	log.Println("Database migrations completed successfully")
	return nil
}