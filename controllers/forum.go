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
	"himtalks-backend/utils"

	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
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

