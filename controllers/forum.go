package controllers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"himtalks-backend/config"
	"himtalks-backend/models"

	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/gorilla/mux"
)

type ForumController struct {
	DB *sql.DB
}

// CreateForum membuat postingan forum (admin only via route middleware)
// Admin bisa membuat forum kapan saja (tanpa batasan waktu)
func (fc *ForumController) CreateForum(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form dengan batas max 2MB di memory
	err := r.ParseMultipartForm(2 << 20) // 2MB
	if err != nil {
		http.Error(w, "File size is too large or invalid multipart form", http.StatusBadRequest)
		return
	}

	var forum models.Forum
	forum.Title = strings.TrimSpace(r.FormValue("title"))
	forum.Content = strings.TrimSpace(r.FormValue("content"))

	if forum.Title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}
	if forum.Content == "" {
		http.Error(w, "Content is required", http.StatusBadRequest)
		return
	}

	// Proses upload gambar (opsional)
	file, header, err := r.FormFile("image")
	if err == nil {
		defer file.Close()

		// Validasi ukuran file (maks 2MB)
		if header.Size > 2<<20 {
			http.Error(w, "Ukuran gambar maksimal 2MB", http.StatusBadRequest)
			return
		}

		// Validasi ekstensi/tipe konten
		ext := strings.ToLower(header.Filename)
		if !strings.HasSuffix(ext, ".jpg") && !strings.HasSuffix(ext, ".jpeg") && !strings.HasSuffix(ext, ".png") {
			http.Error(w, "Gambar harus berformat jpg/png", http.StatusBadRequest)
			return
		}

		// Upload ke Cloudinary jika client sudah diinisialisasi
		if config.CloudinaryClient != nil {
			uploadParam, errUpload := config.CloudinaryClient.Upload.Upload(r.Context(), file, uploader.UploadParams{
				Folder: "forums",
			})
			if errUpload != nil {
				log.Printf("Error upload ke Cloudinary: %v", errUpload)
				http.Error(w, "Failed to upload image", http.StatusInternalServerError)
				return
			}
			forum.ImageURL = uploadParam.SecureURL
		} else {
			log.Println("Cloudinary Client is not initialized, skipping upload")
		}
	} else if err != http.ErrMissingFile {
		http.Error(w, "Error saat membaca file", http.StatusBadRequest)
		return
	}

	query := `INSERT INTO forums (title, content, image_url)
	          VALUES ($1, $2, NULLIF($3, ''))
	          RETURNING id, created_at`
	err = fc.DB.QueryRow(query, forum.Title, forum.Content, forum.ImageURL).Scan(&forum.ID, &forum.CreatedAt)
	if err != nil {
		log.Printf("Error inserting forum: %v", err)
		http.Error(w, "Failed to create forum", http.StatusInternalServerError)
		return
	}

	forum.IsCommentable = true

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
		forum.IsCommentable = time.Since(forum.CreatedAt) <= 7*24*time.Hour
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

	forum.IsCommentable = time.Since(forum.CreatedAt) <= 7*24*time.Hour

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(forum)
}

// UpdateForum memperbarui data forum (mendukung partial update)
func (fc *ForumController) UpdateForum(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	err := r.ParseMultipartForm(2 << 20) // 2MB
	if err != nil {
		http.Error(w, "File size is too large or invalid multipart form", http.StatusBadRequest)
		return
	}

	// Fetch existing forum data first for partial update
	var existingTitle, existingContent, imageURL string
	err = fc.DB.QueryRow("SELECT title, content, COALESCE(image_url, '') FROM forums WHERE id = $1", id).Scan(&existingTitle, &existingContent, &imageURL)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Forum not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to fetch forum", http.StatusInternalServerError)
		return
	}

	// Use existing values as fallback if not provided
	title := strings.TrimSpace(r.FormValue("title"))
	content := strings.TrimSpace(r.FormValue("content"))

	if title == "" {
		title = existingTitle
	}
	if content == "" {
		content = existingContent
	}

	// Proses upload gambar (opsional)
	file, header, err := r.FormFile("image")
	if err == nil {
		defer file.Close()

		if header.Size > 2<<20 {
			http.Error(w, "Ukuran gambar maksimal 2MB", http.StatusBadRequest)
			return
		}

		ext := strings.ToLower(header.Filename)
		if !strings.HasSuffix(ext, ".jpg") && !strings.HasSuffix(ext, ".jpeg") && !strings.HasSuffix(ext, ".png") {
			http.Error(w, "Gambar harus berformat jpg/png", http.StatusBadRequest)
			return
		}

		if config.CloudinaryClient != nil {
			uploadParam, errUpload := config.CloudinaryClient.Upload.Upload(r.Context(), file, uploader.UploadParams{
				Folder: "forums",
			})
			if errUpload != nil {
				log.Printf("Error upload ke Cloudinary: %v", errUpload)
				http.Error(w, "Failed to upload image", http.StatusInternalServerError)
				return
			}
			imageURL = uploadParam.SecureURL
		} else {
			log.Println("Cloudinary Client is not initialized, skipping upload")
		}
	} else if err != http.ErrMissingFile {
		http.Error(w, "Error saat membaca file", http.StatusBadRequest)
		return
	}

	_, err = fc.DB.Exec("UPDATE forums SET title = $1, content = $2, image_url = NULLIF($3, '') WHERE id = $4", title, content, imageURL, id)
	if err != nil {
		log.Printf("Error updating forum: %v", err)
		http.Error(w, "Failed to update forum", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Forum updated successfully"})
}

// DeleteForum menghapus data forum dan komentarnya
func (fc *ForumController) DeleteForum(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var exists bool
	err := fc.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM forums WHERE id=$1)", id).Scan(&exists)
	if err != nil {
		http.Error(w, "Failed to verify forum", http.StatusInternalServerError)
		return
	}
	if !exists {
		http.Error(w, "Forum not found", http.StatusNotFound)
		return
	}

	_, _ = fc.DB.Exec("DELETE FROM comments WHERE forum_id = $1", id)

	_, err = fc.DB.Exec("DELETE FROM forums WHERE id = $1", id)
	if err != nil {
		log.Printf("Error deleting forum: %v", err)
		http.Error(w, "Failed to delete forum", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Forum deleted successfully"})
}
