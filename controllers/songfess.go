package controllers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv" // Untuk konversi string ke int
	"time"

	"himtalks-backend/models"
	"himtalks-backend/ws"

	"github.com/gorilla/mux" // Untuk akses URL params
)

type SongfessController struct {
	DB *sql.DB
}

// SendSongfess menangani pengiriman songfess
func (sc *SongfessController) SendSongfess(w http.ResponseWriter, r *http.Request) {
	log.Println("Received POST request to /songfess")
	var songfess models.Songfess
	err := json.NewDecoder(r.Body).Decode(&songfess)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Validasi panjang pesan berdasarkan konfigurasi
	charLimit, _ := models.GetMessageCharLimit(sc.DB)
	if len(songfess.Content) > charLimit {
		http.Error(w, fmt.Sprintf("Message content exceeds character limit of %d", charLimit), http.StatusBadRequest)
		return
	}

	// Simpan songfess ke database (perhatikan penambahan preview_url)
	query := `INSERT INTO songfess (content, song_id, song_title, artist, album_art, preview_url, start_time, end_time, sender_name, recipient_name) 
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) 
              RETURNING id, created_at`

	err = sc.DB.QueryRow(
		query,
		songfess.Content,
		songfess.SongID,
		songfess.SongTitle,
		songfess.Artist,
		songfess.AlbumArt,
		songfess.PreviewURL, // Field baru
		songfess.StartTime,
		songfess.EndTime,
		songfess.SenderName,
		songfess.RecipientName,
	).Scan(&songfess.ID, &songfess.CreatedAt)

	if err != nil {
		http.Error(w, "Failed to save songfess: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Broadcast songfess to WebSocket clients
	ws.BroadcastMessage(ws.Message{
		Type: "songfess",
		Data: songfess,
	})

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(songfess)
	log.Printf("Received songfess: %+v", songfess)
}

// GetSongfessList mengembalikan daftar songfess
func (sc *SongfessController) GetSongfessList(w http.ResponseWriter, r *http.Request) {
	rows, err := sc.DB.Query(`
		SELECT id, content, song_id, song_title, artist, album_art, 
		       COALESCE(preview_url, '') as preview_url, 
		       COALESCE(start_time, 0) as start_time, 
		       COALESCE(end_time, 0) as end_time, 
		       COALESCE(sender_name, '') as sender_name, 
		       COALESCE(recipient_name, '') as recipient_name, 
		       created_at 
		FROM songfess 
		ORDER BY created_at DESC`)
	if err != nil {
		log.Printf("Error querying songfess: %v", err)
		http.Error(w, "Failed to fetch songfess", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	songfessList := []models.Songfess{}
	for rows.Next() {
		var songfess models.Songfess
		err := rows.Scan(
			&songfess.ID,
			&songfess.Content,
			&songfess.SongID,
			&songfess.SongTitle,
			&songfess.Artist,
			&songfess.AlbumArt,
			&songfess.PreviewURL,
			&songfess.StartTime,
			&songfess.EndTime,
			&songfess.SenderName,
			&songfess.RecipientName,
			&songfess.CreatedAt)
		if err != nil {
			log.Printf("Error scanning songfess row: %v", err)
			http.Error(w, "Failed to scan songfess", http.StatusInternalServerError)
			return
		}
		songfessList = append(songfessList, songfess)
	}

	// Check for errors from iterating over rows
	if err = rows.Err(); err != nil {
		log.Printf("Error iterating over songfess rows: %v", err)
		http.Error(w, "Failed to process songfess data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(songfessList)
}

// Hanya menampilkan data >= cutoff
func (sc *SongfessController) GetSongfessListWithCutoff(w http.ResponseWriter, r *http.Request, cutoff time.Time) {
	rows, err := sc.DB.Query(`
        SELECT id, content, song_id, song_title, artist, album_art, 
               COALESCE(preview_url, '') as preview_url, 
               COALESCE(start_time, 0) as start_time, 
               COALESCE(end_time, 0) as end_time, 
               COALESCE(sender_name, '') as sender_name, 
               COALESCE(recipient_name, '') as recipient_name, 
               created_at 
        FROM songfess 
        WHERE created_at >= $1 
        ORDER BY created_at DESC`, cutoff)
	if err != nil {
		log.Printf("Error querying songfess with cutoff: %v", err)
		http.Error(w, "Failed to fetch songfess", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	songfessList := []models.Songfess{}
	for rows.Next() {
		var songfess models.Songfess
		err := rows.Scan(
			&songfess.ID,
			&songfess.Content,
			&songfess.SongID,
			&songfess.SongTitle,
			&songfess.Artist,
			&songfess.AlbumArt,
			&songfess.PreviewURL, // Field baru
			&songfess.StartTime,
			&songfess.EndTime,
			&songfess.SenderName,
			&songfess.RecipientName,
			&songfess.CreatedAt)
		if err != nil {
			log.Printf("Error scanning songfess row with cutoff: %v", err)
			http.Error(w, "Failed to scan songfess", http.StatusInternalServerError)
			return
		}
		songfessList = append(songfessList, songfess)
	}

	// Check for errors from iterating over rows
	if err = rows.Err(); err != nil {
		log.Printf("Error iterating over songfess rows with cutoff: %v", err)
		http.Error(w, "Failed to process songfess data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(songfessList)
}

// GetSongfessById mengembalikan songfess berdasarkan ID
func (sc *SongfessController) GetSongfessById(w http.ResponseWriter, r *http.Request) {
	// Ambil ID dari parameter URL
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid songfess ID", http.StatusBadRequest)
		return
	}

	// Query database untuk mencari songfess dengan ID tertentu
	query := `
        SELECT id, content, song_id, song_title, artist, album_art, 
               COALESCE(preview_url, '') as preview_url,
               COALESCE(start_time, 0) as start_time, 
               COALESCE(end_time, 0) as end_time, 
               COALESCE(sender_name, '') as sender_name, 
               COALESCE(recipient_name, '') as recipient_name, 
               created_at 
        FROM songfess 
        WHERE id = $1`

	var songfess models.Songfess
	err = sc.DB.QueryRow(query, id).Scan(
		&songfess.ID,
		&songfess.Content,
		&songfess.SongID,
		&songfess.SongTitle,
		&songfess.Artist,
		&songfess.AlbumArt,
		&songfess.PreviewURL, // Field baru
		&songfess.StartTime,
		&songfess.EndTime,
		&songfess.SenderName,
		&songfess.RecipientName,
		&songfess.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Songfess not found", http.StatusNotFound)
		} else {
			log.Printf("Error fetching songfess by ID: %v", err)
			http.Error(w, "Failed to fetch songfess", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(songfess)
}

// DeleteSongfess menghapus songfess berdasarkan ID
func (sc *SongfessController) DeleteSongfess(w http.ResponseWriter, r *http.Request) {
	var data struct{ ID int }
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	if data.ID <= 0 {
		http.Error(w, "Invalid songfess ID", http.StatusBadRequest)
		return
	}

	// Check if songfess exists before deleting
	var exists bool
	err := sc.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM songfess WHERE id=$1)", data.ID).Scan(&exists)
	if err != nil {
		log.Printf("Error checking songfess existence: %v", err)
		http.Error(w, "Failed to verify songfess", http.StatusInternalServerError)
		return
	}

	if !exists {
		http.Error(w, "Songfess not found", http.StatusNotFound)
		return
	}

	// Delete the songfess
	result, err := sc.DB.Exec("DELETE FROM songfess WHERE id=$1", data.ID)
	if err != nil {
		log.Printf("Error deleting songfess: %v", err)
		http.Error(w, "Failed to delete songfess", http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Songfess not found", http.StatusNotFound)
		return
	}

	log.Printf("Songfess with ID %d deleted successfully", data.ID)

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Songfess deleted successfully",
		"id":      data.ID,
	})

	// Kirim pesan ke WebSocket
	msg := ws.Message{
		Type: "delete_songfess",
		Data: data.ID,
	}
	ws.BroadcastMessage(msg)
}
