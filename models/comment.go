package models

import (
	"database/sql"
	"time"
)

type Comment struct {
	ID        int       `json:"id"`
	ForumID   int       `json:"forum_id"`
	Name      string    `json:"name"`
	AvatarID  string    `json:"avatar_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateTableComments membuat tabel comments di PostgreSQL
func CreateTableComments(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS comments (
		id SERIAL PRIMARY KEY,
		forum_id INTEGER NOT NULL REFERENCES forums(id) ON DELETE CASCADE,
		name VARCHAR(50),
		avatar_id VARCHAR(50),
		content TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_comments_forum_id ON comments(forum_id);`
	_, err := db.Exec(query)
	return err
}

