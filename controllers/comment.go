package controllers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"himtalks-backend/models"
	"himtalks-backend/utils"

	"github.com/gorilla/mux"
)

type CommentController struct {
	DB *sql.DB
}

// CreateComment membuat komentar untuk forum (publik, user/anonym)
func (cc *CommentController) CreateComment(w http.ResponseWriter, r *http.Request) {
	if !utils.IsMiniForumOpen() {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Mini forum hanya menerima komentar antara pukul 19:00–21:00 WIB",
		})
		return
	}

	vars := mux.Vars(r)
	forumIDStr := vars["id"]
	forumID, err := strconv.Atoi(forumIDStr)
	if err != nil || forumID <= 0 {
		http.Error(w, "Invalid forum id", http.StatusBadRequest)
		return
	}

	var comment models.Comment
	if err := json.NewDecoder(r.Body).Decode(&comment); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	comment.Name = strings.TrimSpace(comment.Name)
	comment.Content = strings.TrimSpace(comment.Content)
	if comment.Name == "" {
		comment.Name = "Anonim"
	}
	if comment.Content == "" {
		http.Error(w, "Content is required", http.StatusBadRequest)
		return
	}

	// Pastikan forum ada dan cek apakah masih bisa dikomentari (max 7 hari)
	var forumCreatedAt time.Time
	if err := cc.DB.QueryRow("SELECT created_at FROM forums WHERE id=$1", forumID).Scan(&forumCreatedAt); err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Forum not found", http.StatusNotFound)
			return
		}
		log.Printf("Error checking forum: %v", err)
		http.Error(w, "Failed to verify forum", http.StatusInternalServerError)
		return
	}

	// Cek apakah forum sudah lewat 7 hari
	if time.Since(forumCreatedAt) > 7*24*time.Hour {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Forum ini sudah tidak menerima komentar baru (lewat 7 hari)",
		})
		return
	}

	comment.ForumID = forumID
	query := `INSERT INTO comments (forum_id, name, content, avatar_id)
	          VALUES ($1, $2, $3, NULLIF($4, ''))
	          RETURNING id, created_at`
	err = cc.DB.QueryRow(query, comment.ForumID, comment.Name, comment.Content, comment.AvatarID).Scan(&comment.ID, &comment.CreatedAt)
	if err != nil {
		log.Printf("Error inserting comment: %v", err)
		http.Error(w, "Failed to create comment", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(comment)
}

// GetCommentsByForum mengembalikan komentar untuk forum tertentu (publik)
func (cc *CommentController) GetCommentsByForum(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	forumIDStr := vars["id"]
	forumID, err := strconv.Atoi(forumIDStr)
	if err != nil || forumID <= 0 {
		http.Error(w, "Invalid forum id", http.StatusBadRequest)
		return
	}

	rows, err := cc.DB.Query(`
		SELECT id, forum_id,
		       COALESCE(name, '') as name,
		       COALESCE(avatar_id, '') as avatar_id,
		       content,
		       created_at
		FROM comments
		WHERE forum_id = $1
		ORDER BY created_at ASC`, forumID)
	if err != nil {
		log.Printf("Error querying comments: %v", err)
		http.Error(w, "Failed to fetch comments", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	comments := []models.Comment{}
	for rows.Next() {
		var c models.Comment
		if err := rows.Scan(&c.ID, &c.ForumID, &c.Name, &c.AvatarID, &c.Content, &c.CreatedAt); err != nil {
			log.Printf("Error scanning comment row: %v", err)
			http.Error(w, "Failed to scan comment", http.StatusInternalServerError)
			return
		}
		comments = append(comments, c)
	}
	if err := rows.Err(); err != nil {
		log.Printf("Error iterating comment rows: %v", err)
		http.Error(w, "Failed to process comment data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comments)
}

