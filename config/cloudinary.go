package config

import (
	"log"
	"os"

	"github.com/cloudinary/cloudinary-go/v2"
)

var CloudinaryClient *cloudinary.Cloudinary

// InitCloudinary inisialisasi client Cloudinary
func InitCloudinary() {
	cloudinaryURL := os.Getenv("CLOUDINARY_URL")
	if cloudinaryURL == "" {
		log.Println("Warning: CLOUDINARY_URL is not set")
		return
	}

	cld, err := cloudinary.NewFromURL(cloudinaryURL)
	if err != nil {
		log.Fatalf("Failed to initialize Cloudinary: %v", err)
	}

	cld.Config.URL.Secure = true
	CloudinaryClient = cld
	log.Println("Cloudinary initialized successfully")
}
