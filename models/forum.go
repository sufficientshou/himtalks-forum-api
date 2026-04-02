package models

import (
	"database/sql"
	"time"
)

type Forum struct {
	ID            int       `json:"id"`
	Title         string    `json:"title"`
	Content       string    `json:"content"`
	ImageURL      string    `json:"image_url"`
	CreatedAt     time.Time `json:"created_at"`
	IsCommentable bool      `json:"is_commentable"`
}

// CreateTableForums membuat tabel forums di PostgreSQL
func CreateTableForums(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS forums (
		id SERIAL PRIMARY KEY,
		title TEXT NOT NULL,
		content TEXT NOT NULL,
		image_url TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`
	_, err := db.Exec(query)
	return err
}

