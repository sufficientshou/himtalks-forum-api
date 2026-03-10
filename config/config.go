package config

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq" // Import driver PostgreSQL
)

// LoadEnv memuat variabel lingkungan dari file .env (opsional di Docker)
func LoadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file, using environment variables")
	}
}

// GetDBConfig mengembalikan string koneksi database PostgreSQL
func GetDBConfig() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("POSTGRES_DB"),
		os.Getenv("SSL_MODE"),
	)
}

// GetSecretKey mengembalikan secret key dari environment
func GetSecretKey() string {
	return os.Getenv("SECRET_KEY")
}

// ConnectDB membuat koneksi ke database PostgreSQL
func ConnectDB() (*sql.DB, error) {
	db, err := sql.Open("postgres", GetDBConfig())
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	log.Println("Connected to PostgreSQL database!")
	return db, nil
}
