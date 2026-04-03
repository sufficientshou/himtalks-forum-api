package controllers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	"himtalks-backend/utils"

	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var (
	googleOAuthConfig *oauth2.Config
	randomState       = "random"
)

func init() {
	// Load .env (opsional - di Docker env dari container)
	_ = godotenv.Load()

	// Inisialisasi googleOAuthConfig
	googleOAuthConfig = &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email"},
		Endpoint:     google.Endpoint,
	}

	// Debug
	log.Println("GOOGLE_CLIENT_ID:", os.Getenv("GOOGLE_CLIENT_ID"))
	log.Println("GOOGLE_CLIENT_SECRET:", os.Getenv("GOOGLE_CLIENT_SECRET"))
	log.Println("GOOGLE_REDIRECT_URL:", os.Getenv("GOOGLE_REDIRECT_URL"))
}

type AdminController struct{}

func (ac *AdminController) Login(w http.ResponseWriter, r *http.Request) {
	url := googleOAuthConfig.AuthCodeURL(randomState, oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (ac *AdminController) Callback(w http.ResponseWriter, r *http.Request) {
	if r.FormValue("state") != randomState {
		log.Println("State is not valid")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	token, err := googleOAuthConfig.Exchange(context.Background(), r.FormValue("code"))
	if err != nil {
		log.Printf("Could not get token: %v\n", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		log.Printf("Could not get user info: %v\n", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	defer resp.Body.Close()

	var userInfo struct {
		Email string `json:"email"`
	}
	err = json.NewDecoder(resp.Body).Decode(&userInfo)
	if err != nil {
		log.Printf("Could not decode user info: %v\n", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	if !strings.HasSuffix(userInfo.Email, "@student.unsika.ac.id") {
		log.Println("Email not allowed")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// Generate token JWT
	tokenString, err := utils.GenerateToken(userInfo.Email)
	if err != nil {
		log.Printf("Could not generate token: %v\n", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Set token sebagai HTTP-Only cookie
	cookieDomain := os.Getenv("COOKIE_DOMAIN")
	if cookieDomain == "" {
		cookieDomain = "" // Kosong untuk mengikat ke domain saat ini
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "jwt",
		Value:    tokenString,
		Path:     "/",
		Domain:   cookieDomain,
		MaxAge:   24 * 60 * 60, // 24 jam
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode, // Untuk cross-origin
	})

	// Redirect ke frontend
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "https://admin-himtalks.vercel.app/"
	}

	http.Redirect(w, r, frontendURL+"/auth-success", http.StatusTemporaryRedirect)
}

// Tambahkan fungsi logout
func (ac *AdminController) Logout(w http.ResponseWriter, r *http.Request) {
	cookieDomain := os.Getenv("COOKIE_DOMAIN")

	// Hapus cookie dengan mengatur MaxAge: -1
	http.SetCookie(w, &http.Cookie{
		Name:     "jwt",
		Value:    "",
		Path:     "/",
		Domain:   cookieDomain,
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Logout successful",
	})
}
