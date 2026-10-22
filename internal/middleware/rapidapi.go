package middleware

import (
	"encoding/json"
	"net/http"
	"os"
)

// RapidAPIAuth memastikan request hanya datang dari proxy resmi RapidAPI
func RapidAPIAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Biarkan endpoint /health tetap terbuka agar mudah dipantau
		if r.URL.Path == "/health" {
			next.ServeHTTP(w, r)
			return
		}

		// Ambil kode rahasia dari Environment Variable server
		expectedSecret := os.Getenv("RAPIDAPI_PROXY_SECRET")

		// Jika RAPIDAPI_PROXY_SECRET belum diatur (seperti saat koding di laptop lokal),
		// izinkan lewat agar kamu mudah melakukan testing.
		if expectedSecret == "" {
			next.ServeHTTP(w, r)
			return
		}

		// Jika sudah di server production, cek kecocokan header dari RapidAPI
		incomingSecret := r.Header.Get("X-RapidAPI-Proxy-Secret")
		if incomingSecret != expectedSecret {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Forbidden: Akses langsung ditolak. Gunakan API ini melalui RapidAPI Hub.",
			})
			return
		}

		// Jika lolos, lanjutkan ke handler utama
		next.ServeHTTP(w, r)
	})
}