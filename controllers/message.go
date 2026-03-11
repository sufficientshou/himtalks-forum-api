package controllers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"himtalks-backend/models"
	"himtalks-backend/ws"
)

type MessageController struct {
	DB *sql.DB
}

// SendMessage menangani pengiriman pesan anonim
func (mc *MessageController) SendMessage(w http.ResponseWriter, r *http.Request) {
	var message models.Message
	err := json.NewDecoder(r.Body).Decode(&message)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Validasi kategori
	if message.Category != "kritik" && message.Category != "saran" {
		http.Error(w, "Category must be 'kritik' or 'saran'", http.StatusBadRequest)
		return
	}

	// Cek blacklist
	blacklisted, err := models.IsBlacklisted(mc.DB, message.Content)
	if err != nil {
		http.Error(w, "Error checking blacklist", http.StatusInternalServerError)
		return
	}
	if blacklisted {
		http.Error(w, "Message contains blacklisted word", http.StatusForbidden)
		return
	}

	// Simpan pesan ke database
	query := `INSERT INTO messages (content, sender_name, recipient_name, category) 
              VALUES ($1, $2, $3, $4) 
              RETURNING id, created_at`

	err = mc.DB.QueryRow(
		query,
		message.Content,
		message.SenderName,
		message.RecipientName,
		message.Category,
	).Scan(&message.ID, &message.CreatedAt)

	if err != nil {
		http.Error(w, "Failed to save message: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Kirim pesan ke WebSocket
	ws.BroadcastMessage(ws.Message{
		Type: "message",
		Data: message,
	})

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(message)
}

// GetMessageList mengembalikan daftar pesan
func (mc *MessageController) GetMessageList(w http.ResponseWriter, r *http.Request) {
	rows, err := mc.DB.Query(`
		SELECT id, content, 
		       COALESCE(sender_name, '') as sender_name, 
		       COALESCE(recipient_name, '') as recipient_name, 
		       category, created_at 
		FROM messages 
		ORDER BY created_at DESC`)
	if err != nil {
		log.Printf("Error querying messages: %v", err)
		http.Error(w, "Failed to fetch messages", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	messageList := []models.Message{}
	for rows.Next() {
		var message models.Message
		err := rows.Scan(
			&message.ID,
			&message.Content,
			&message.SenderName,
			&message.RecipientName,
			&message.Category,
			&message.CreatedAt,
		)
		if err != nil {
			log.Printf("Error scanning message row: %v", err)
			http.Error(w, "Failed to scan message", http.StatusInternalServerError)
			return
		}
		messageList = append(messageList, message)
	}

	// Check for errors from iterating over rows
	if err = rows.Err(); err != nil {
		log.Printf("Error iterating over message rows: %v", err)
		http.Error(w, "Failed to process message data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messageList)
}

// DeleteMessage menghapus pesan berdasarkan ID
func (mc *MessageController) DeleteMessage(w http.ResponseWriter, r *http.Request) {
	var data struct{ ID int }
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	if data.ID <= 0 {
		http.Error(w, "Invalid message ID", http.StatusBadRequest)
		return
	}

	// Check if message exists before deleting
	var exists bool
	err := mc.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM messages WHERE id=$1)", data.ID).Scan(&exists)
	if err != nil {
		log.Printf("Error checking message existence: %v", err)
		http.Error(w, "Failed to verify message", http.StatusInternalServerError)
		return
	}

	if !exists {
		http.Error(w, "Message not found", http.StatusNotFound)
		return
	}

	// Delete the message
	result, err := mc.DB.Exec("DELETE FROM messages WHERE id=$1", data.ID)
	if err != nil {
		log.Printf("Error deleting message: %v", err)
		http.Error(w, "Failed to delete message", http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Message not found", http.StatusNotFound)
		return
	}

	log.Printf("Message with ID %d deleted successfully", data.ID)

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Message deleted successfully",
		"id":      data.ID,
	})

	// Kirim pesan ke WebSocket
	msg := ws.Message{
		Type: "delete_message", // Make type more specific
		Data: data.ID,
	}
	ws.BroadcastMessage(msg)
}
