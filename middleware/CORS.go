package middleware

import (
	"net/http"
	"os"
	"strings"
)

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip CORS check untuk WebSocket dan OAuth callback
		// karena mereka tidak mengirim Origin header standar
		path := r.URL.Path
		if path == "/ws" || strings.HasPrefix(path, "/auth/google/") {
			next.ServeHTTP(w, r)
			return
		}

		// Ambil allowed origins dari environment + hardcoded production domains
		allowedOrigins := []string{
			"http://localhost:3000",
			"http://localhost:5173",
			"http://himtalks.japaneast.cloudapp.azure.com",
			"https://himtalks.vercel.app",
			"https://himtalks-admin.vercel.app",
			"https://api.teknohive.me",
		}

		// Tambahkan FRONTEND_URL dari env jika ada dan belum termasuk
		if frontendURL := os.Getenv("FRONTEND_URL"); frontendURL != "" {
			frontendURL = strings.TrimRight(frontendURL, "/")
			found := false
			for _, o := range allowedOrigins {
				if o == frontendURL {
					found = true
					break
				}
			}
			if !found {
				allowedOrigins = append(allowedOrigins, frontendURL)
			}
		}

		origin := r.Header.Get("Origin")

		// Cek apakah origin ada di daftar yang diizinkan
		originAllowed := false
		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin {
				originAllowed = true
				break
			}
		}

		// Untuk preflight (OPTIONS), harus tetap diproses meskipun tanpa Origin
		if r.Method == "OPTIONS" {
			if originAllowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Max-Age", "86400")
			w.WriteHeader(http.StatusOK)
			return
		}

		// Untuk request biasa: jika ada Origin header dan tidak diizinkan, tolak
		if origin != "" && !originAllowed {
			http.Error(w, "CORS: origin not allowed", http.StatusForbidden)
			return
		}

		// Jika tidak ada Origin header (direct access dari Postman/curl/browser langsung), tolak
		if origin == "" {
			http.Error(w, "CORS: origin header required", http.StatusForbidden)
			return
		}

		// Origin valid — set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		next.ServeHTTP(w, r)
	})
}
