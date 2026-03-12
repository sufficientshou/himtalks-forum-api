package controllers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"himtalks-backend/models"
	"himtalks-backend/utils"

	"github.com/gorilla/mux"
)

type ForumController struct {
	DB *sql.DB
}

// CreateForum membuat postingan forum (admin only via route middleware)
func (fc *ForumController) CreateForum(w http.ResponseWriter, r *http.Request) {
	if !utils.IsMiniForumOpen() {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Mini forum hanya bisa dibuat antara pukul 19:00–21:00 WIB",
		})
		return
	}

	var forum models.Forum
	if err := json.NewDecoder(r.Body).Decode(&forum); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	forum.Title = strings.TrimSpace(forum.Title)
	forum.Content = strings.TrimSpace(forum.Content)
	forum.ImageURL = strings.TrimSpace(forum.ImageURL)

	if forum.Title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}
	if forum.Content == "" {
		http.Error(w, "Content is required", http.StatusBadRequest)
		return
	}

	query := `INSERT INTO forums (title, content, image_url)
	          VALUES ($1, $2, NULLIF($3, ''))
	          RETURNING id, created_at`
	err := fc.DB.QueryRow(query, forum.Title, forum.Content, forum.ImageURL).Scan(&forum.ID, &forum.CreatedAt)
	if err != nil {
		log.Printf("Error inserting forum: %v", err)
		http.Error(w, "Failed to create forum", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(forum)
}

// GetForumList mengembalikan daftar forum (publik)
func (fc *ForumController) GetForumList(w http.ResponseWriter, r *http.Request) {
	rows, err := fc.DB.Query(`
		SELECT id, title, content,
		       COALESCE(image_url, '') as image_url,
		       created_at
		FROM forums
		ORDER BY created_at DESC`)
	if err != nil {
		log.Printf("Error querying forums: %v", err)
		http.Error(w, "Failed to fetch forums", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	forums := []models.Forum{}
	for rows.Next() {
		var forum models.Forum
		if err := rows.Scan(&forum.ID, &forum.Title, &forum.Content, &forum.ImageURL, &forum.CreatedAt); err != nil {
			log.Printf("Error scanning forum row: %v", err)
			http.Error(w, "Failed to scan forum", http.StatusInternalServerError)
			return
		}
		forums = append(forums, forum)
	}
	if err := rows.Err(); err != nil {
		log.Printf("Error iterating forum rows: %v", err)
		http.Error(w, "Failed to process forum data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(forums)
}

// GetForumByID mengembalikan forum berdasarkan id (publik)
func (fc *ForumController) GetForumByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var forum models.Forum
	err := fc.DB.QueryRow(`
		SELECT id, title, content,
		       COALESCE(image_url, '') as image_url,
		       created_at
		FROM forums
		WHERE id = $1`, id).Scan(&forum.ID, &forum.Title, &forum.Content, &forum.ImageURL, &forum.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Forum not found", http.StatusNotFound)
			return
		}
		log.Printf("Error fetching forum by id: %v", err)
		http.Error(w, "Failed to fetch forum", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(forum)
}

