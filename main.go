package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	// Pastikan nama module sesuai dengan yang ada di file go.mod kamu!
	"github.com/acil23/email-guard-api/internal/middleware"
	"github.com/acil23/email-guard-api/internal/validator"
)

func verifyHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	email := r.URL.Query().Get("email")
	if email == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Parameter 'email' wajib diisi. Contoh: /v1/verify?email=tes@gmail.com",
		})
		return
	}

	result := validator.Analyze(email)
	duration := time.Since(start).String()

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Response-Time", duration)
	json.NewEncoder(w).Encode(result)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"status":                "healthy",
		"loaded_temp_domains":   validator.GetDomainCount(),
		"timestamp":             time.Now().UTC().Format(time.RFC3339),
	})
}

func main() {
	// 1. Jalankan pemuat database temp-mail & auto-update setiap 24 jam
	validator.StartAutoUpdater(24 * time.Hour)

	// 2. Siapkan Router
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler)
	mux.HandleFunc("GET /v1/verify", verifyHandler)

	// 3. Bungkus router dengan Middleware Keamanan RapidAPI
	securedMux := middleware.RapidAPIAuth(mux)

	// 4. Gunakan PORT dari environment server cloud (default ke 8080 jika di lokal)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("🚀 Server Email Guard API menyala di http://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, securedMux))
}