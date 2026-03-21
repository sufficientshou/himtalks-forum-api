package utils

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4" // Ubah import path
)

var jwtKey []byte

func getJWTKey() []byte {
	if len(jwtKey) == 0 {
		jwtKey = []byte(os.Getenv("SECRET_KEY"))
	}
	return jwtKey
}

type Claims struct {
	Email                string `json:"email"`
	jwt.RegisteredClaims        // Ubah StandardClaims menjadi RegisteredClaims
}

// GenerateToken membuat token JWT baru
func GenerateToken(email string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour) // Token berlaku 24 jam
	claims := &Claims{
		Email: email,
		RegisteredClaims: jwt.RegisteredClaims{ // Ubah StandardClaims menjadi RegisteredClaims
			ExpiresAt: jwt.NewNumericDate(expirationTime), // Ubah format ExpiresAt
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(getJWTKey())
}

// ValidateToken memvalidasi token JWT
func ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Validasi metode signing
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return getJWTKey(), nil
	})

	if err != nil {
		log.Printf("JWT validation error: %v", err)
		return nil, err
	}

	if !token.Valid {
		log.Printf("JWT token is invalid")
		return nil, fmt.Errorf("invalid token")
	}

	log.Printf("JWT validated successfully for email: %s", claims.Email)
	return claims, nil
}
