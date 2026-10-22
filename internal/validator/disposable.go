package validator

import (
	"bufio"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Menggunakan struct{} karena ukurannya 0 byte di memori (Set pattern di Go)
var disposableDomains = make(map[string]struct{})

// Sumber open-source daftar domain temp-mail yang selalu diupdate komunitas global
const blocklistURL = "https://raw.githubusercontent.com/disposable-email-domains/disposable-email-domains/master/disposable_email_blocklist.conf"

// LoadDisposableDomains mengunduh daftar domain ke dalam RAM saat server mulai menyala
func LoadDisposableDomains() {
	// 1. Masukkan daftar bawaan (fallback jika internet sedang gangguan)
	defaultDomains := []string{
		"10minutemail.com", "yopmail.com", "mailinator.com",
		"tempmail.com", "guerrillamail.com", "trashmail.com",
		"sharklasers.com", "getnada.com", "temp-mail.org",
	}
	for _, d := range defaultDomains {
		disposableDomains[d] = struct{}{}
	}

	// 2. Unduh daftar lengkap dari GitHub dengan batas waktu (timeout) 10 detik
	client := http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(blocklistURL)
	if err != nil {
		fmt.Printf("⚠️ Gagal mengunduh daftar temp-mail eksternal, memakai data lokal: %v\n", err)
		return
	}
	defer resp.Body.Close() // Wajib ditutup agar tidak terjadi memory leak!

	// 3. Baca baris demi baris secara efisien menggunakan bufio.Scanner
	scanner := bufio.NewScanner(resp.Body)
	count := 0
	for scanner.Scan() {
		line := strings.TrimSpace(strings.ToLower(scanner.Text()))
		if line != "" && !strings.HasPrefix(line, "#") {
			disposableDomains[line] = struct{}{}
			count++
		}
	}

	fmt.Printf("✅ Berhasil memuat %d domain temp-mail ke dalam memori RAM\n", len(disposableDomains))
}

// IsDisposable mengecek apakah domain ada di daftar hitam dalam waktu nanodetik (O(1))
func IsDisposable(domain string) bool {
	_, exists := disposableDomains[domain]
	return exists
}