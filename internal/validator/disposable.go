package validator

import (
	"bufio"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

var (
	disposableDomains = make(map[string]struct{})
	// mu menjaga agar map aman dibaca ribuan Goroutine sekaligus saat sedang diupdate
	mu sync.RWMutex
)

const blocklistURL = "https://raw.githubusercontent.com/disposable-email-domains/disposable-email-domains/master/disposable_email_blocklist.conf"

// LoadDisposableDomains mengunduh daftar domain dan menukarnya ke memori secara aman
func LoadDisposableDomains() {
	// 1. Buat map baru sementara (agar proses download tidak menghambat user yang sedang memakai API)
	newDomains := make(map[string]struct{})

	defaultDomains := []string{
		"10minutemail.com", "yopmail.com", "mailinator.com",
		"tempmail.com", "guerrillamail.com", "trashmail.com",
		"sharklasers.com", "getnada.com", "temp-mail.org",
	}
	for _, d := range defaultDomains {
		newDomains[d] = struct{}{}
	}

	client := http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(blocklistURL)
	if err != nil {
		fmt.Printf("⚠️ Gagal mengunduh daftar temp-mail eksternal, memakai data fallback: %v\n", err)
	} else {
		defer resp.Body.Close()
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := strings.TrimSpace(strings.ToLower(scanner.Text()))
			if line != "" && !strings.HasPrefix(line, "#") {
				newDomains[line] = struct{}{}
			}
		}
	}

	// 2. Kunci memori (Write Lock) HANYA saat menukar map lama dengan map baru (hitungan nanodetik)
	mu.Lock()
	disposableDomains = newDomains
	total := len(disposableDomains)
	mu.Unlock()

	fmt.Printf("✅ Database Temp-Mail diperbarui: %d domain aktif di RAM\n", total)
}

// StartAutoUpdater menjalankan pekerja latar belakang untuk update data setiap interval tertentu
func StartAutoUpdater(interval time.Duration) {
	// Muat data pertama kali saat server menyala
	LoadDisposableDomains()

	// Jalankan Goroutine abadi di latar belakang menggunakan Ticker
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			fmt.Println("🔄 Memulai pembaruan otomatis database temp-mail...")
			LoadDisposableDomains()
		}
	}()
}

// IsDisposable membaca map dengan RLock (Read Lock) sehingga ribuan user bisa baca bersamaan
func IsDisposable(domain string) bool {
	mu.RLock()
	_, exists := disposableDomains[domain]
	mu.RUnlock()
	return exists
}

// GetDomainCount mengembalikan jumlah domain yang sedang dimuat di RAM (untuk endpoint /health)
func GetDomainCount() int {
	mu.RLock()
	defer mu.RUnlock()
	return len(disposableDomains)
}