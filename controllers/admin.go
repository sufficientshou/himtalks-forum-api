package controllers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"himtalks-backend/models"
)

var emailKey = contextKey("email")

type contextKey string

type AdminHandler struct {
	DB *sql.DB
}

// Menambahkan admin baru
func (ah *AdminHandler) AddAdmin(w http.ResponseWriter, r *http.Request) {
	// Pastikan sudah dicek IsAdmin di middleware
	var data struct{ Email string }
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}
	if data.Email == "" {
		http.Error(w, "Email required", http.StatusBadRequest)
		return
	}
	err := models.InsertAdmin(ah.DB, data.Email)
	if err != nil {
		http.Error(w, "Failed to add admin", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// Ubah batas hari untuk songfess
func (ah *AdminHandler) UpdateSongfessDays(w http.ResponseWriter, r *http.Request) {
	var data struct{ Days string }
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}
	if err := models.SetConfig(ah.DB, "songfess_days", data.Days); err != nil {
		http.Error(w, "Failed to update config", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// Tambah kata blacklist
func (ah *AdminHandler) AddBlacklistWord(w http.ResponseWriter, r *http.Request) {
	var data struct{ Word string }
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}
	if data.Word == "" {
		http.Error(w, "Word required", http.StatusBadRequest)
		return
	}
	err := models.InsertBlacklistWord(ah.DB, data.Word)
	if err != nil {
		http.Error(w, "Failed to add word", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// GetAdminList mengembalikan daftar semua admin
func (ah *AdminHandler) GetAdminList(w http.ResponseWriter, r *http.Request) {
	rows, err := ah.DB.Query("SELECT email FROM admins")
	if err != nil {
		http.Error(w, "Failed to fetch admins", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	adminList := []string{}
	for rows.Next() {
		var email string
		err := rows.Scan(&email)
		if err != nil {
			http.Error(w, "Failed to scan admin", http.StatusInternalServerError)
			return
		}
		adminList = append(adminList, email)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(adminList)
}

// RemoveAdmin menghapus admin dari sistem
func (ah *AdminHandler) RemoveAdmin(w http.ResponseWriter, r *http.Request) {
	var data struct{ Email string }
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	if data.Email == "" {
		http.Error(w, "Email required", http.StatusBadRequest)
		return
	}

	// Dapatkan email dari admin yang melakukan request
	requesterEmail, ok := r.Context().Value(emailKey).(string)
	if !ok {
		http.Error(w, "Unauthorized: missing requester email", http.StatusUnauthorized)
		return
	}

	// Jangan izinkan admin menghapus dirinya sendiri
	if data.Email == requesterEmail {
		http.Error(w, "Cannot remove yourself from admin", http.StatusBadRequest)
		return
	}

	// Verifikasi bahwa target adalah admin
	isAdmin, err := models.IsAdmin(ah.DB, data.Email)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if !isAdmin {
		http.Error(w, "Email is not an admin", http.StatusBadRequest)
		return
	}

	// Hapus admin
	err = models.DeleteAdmin(ah.DB, data.Email)
	if err != nil {
		http.Error(w, "Failed to remove admin", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Admin removed successfully",
	})
}

// GetBlacklistWords mengembalikan daftar kata-kata terlarang
func (ah *AdminHandler) GetBlacklistWords(w http.ResponseWriter, r *http.Request) {
	words, err := models.GetBlacklistedWords(ah.DB)
	if err != nil {
		http.Error(w, "Failed to fetch blacklisted words", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(words)
}

// RemoveBlacklistWord menghapus kata dari blacklist
func (ah *AdminHandler) RemoveBlacklistWord(w http.ResponseWriter, r *http.Request) {
	var data struct{ Word string }
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	if data.Word == "" {
		http.Error(w, "Word required", http.StatusBadRequest)
		return
	}

	_, err := ah.DB.Exec("DELETE FROM blacklist_words WHERE word=$1", data.Word)
	if err != nil {
		http.Error(w, "Failed to remove word", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Word removed from blacklist successfully",
	})
}

// GetConfigs mengembalikan semua konfigurasi sistem
func (ah *AdminHandler) GetConfigs(w http.ResponseWriter, r *http.Request) {
	configs, err := models.GetAllConfigs(ah.DB)
	if err != nil {
		http.Error(w, "Failed to fetch configs", http.StatusInternalServerError)
		return
	}

	// Jika songfess_days tidak ada, berikan nilai default
	if _, exists := configs["songfess_days"]; !exists {
		configs["songfess_days"] = "7" // Default 7 hari
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(configs)
}
