package main

import (
	"log"
	"net/http"
	"os"

	"himtalks-backend/config"
	"himtalks-backend/models"
	"himtalks-backend/routes"

	"himtalks-backend/ws"

	"github.com/gorilla/handlers"
	"github.com/joho/godotenv"
)

func main() {

	// Load .env file (opsional - di Docker env vars dari container)
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Akses environment variables
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	redirectURL := os.Getenv("GOOGLE_REDIRECT_URL")
	secretKey := os.Getenv("SECRET_KEY")
	cookieDomain := os.Getenv("COOKIE_DOMAIN")
	frontendURL := os.Getenv("FRONTEND_URL")

	log.Println("Cookie Domain:", cookieDomain)
	log.Println("Frontend URL:", frontendURL)

	log.Println("Client ID:", clientID)
	log.Println("Client Secret:", clientSecret)
	log.Println("Redirect URL:", redirectURL)
	log.Println("Secret Key:", secretKey)

	// Load environment variables
	config.LoadEnv()

	// Connect to PostgreSQL
	db, err := config.ConnectDB()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Create tables if they don't exist
	err = models.CreateTableMessages(db)
	if err != nil {
		log.Fatal("Failed to create messages table:", err)
	}
	err = models.CreateTableSongfess(db)
	if err != nil {
		log.Fatal("Failed to create songfess table:", err)
	}
	err = models.CreateTableAdmins(db)
	if err != nil {
		log.Fatal("Failed to create admins table:", err)
	}
	err = models.CreateTableConfigs(db)
	if err != nil {
		log.Fatal("Failed to create configs table:", err)
	}
	err = models.CreateTableBlacklist(db)
	if err != nil {
		log.Fatal("Failed to create blacklist table:", err)
	}
	err = models.CreateTableForums(db)
	if err != nil {
		log.Fatal("Failed to create forums table:", err)
	}
	err = models.CreateTableComments(db)
	if err != nil {
		log.Fatal("Failed to create comments table:", err)
	}

	// Hardcode admin pertama (ganti dengan email yang lu mau)
	hardcodedAdmin := "2510631170035@student.unsika.ac.id"
	
	// Cek apakah admin sudah ada
	isAdmin, err := models.IsAdmin(db, hardcodedAdmin)
	if err != nil {
		log.Printf("Error checking admin: %v", err)
	} else if !isAdmin {
		// Insert admin pertama
		err = models.InsertAdmin(db, hardcodedAdmin)
		if err != nil {
			log.Printf("Error inserting hardcoded admin: %v", err)
		} else {
			log.Printf("Hardcoded admin inserted: %s", hardcodedAdmin)
		}
	} else {
		log.Printf("Admin already exists: %s", hardcodedAdmin)
	}

	// Setup routes
	r := routes.SetupRoutes(db)

	// Start WebSocket handler
	go ws.HandleMessages()

	// Middleware untuk menonaktifkan CORS
	corsHandler := handlers.CORS(
		handlers.AllowedOrigins([]string{"http://localhost:3000", "http://localhost:5173", "http://himtalks.japaneast.cloudapp.azure.com", "https://himtalks.vercel.app", "https://himtalks-admin.vercel.app"}), // Ganti * dengan domain FE
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
		handlers.AllowCredentials(), // Mengaktifkan penggunaan credentials (cookies/session)
	)

	// Start server dengan middleware CORS
	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", corsHandler(r)))
}
