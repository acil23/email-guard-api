package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	// Ganti "github.com/usernamekamu/email-guard-api" sesuai nama di go.mod kamu!
	"github.com/acil23/email-guard-api/internal/validator"
)

func verifyHandler(w http.ResponseWriter, r *http.Request) {
	// Catat waktu mulai untuk mengukur kecepatan eksekusi Go
	start := time.Now()

	// Ambil parameter ?email=... dari URL
	email := r.URL.Query().Get("email")
	if email == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Parameter 'email' wajib diisi. Contoh: /v1/verify?email=tes@gmail.com",
		})
		return
	}

	// Panggil fungsi Analyze dari package validator kita
	result := validator.Analyze(email)

	// Hitung durasi proses dalam mikrodetik/milidetik
	duration := time.Since(start).String()

	// Set Header sebagai JSON dan tambahkan info kecepatan di Header
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Response-Time", duration)

	// Kirim hasil dalam bentuk JSON
	json.NewEncoder(w).Encode(result)
}

func main() {
	mux := http.NewServeMux() // Jika Go versi < 1.22, gunakan http.NewServeMux()

	// Daftarkan rute endpoint
	mux.HandleFunc("GET /v1/verify", verifyHandler)

	port := ":8080"
	fmt.Printf("🚀 Server Email Guard API menyala di http://localhost%s\n", port)
	log.Fatal(http.ListenAndServe(port, mux))
}
