package validator

import (
	"context"
	"net"
	"net/mail"
	"strings"
	"sync"
	"time"
)

type Result struct {
	Email           string `json:"email"`
	IsValidSyntax   bool   `json:"is_valid_syntax"`
	Username        string `json:"username,omitempty"`
	Domain          string `json:"domain,omitempty"`
	IsRoleAccount   bool   `json:"is_role_account"`
	IsDisposable    bool   `json:"is_disposable"`
	HasMXRecord     bool   `json:"has_mx_record"`
	SuggestedDomain string `json:"suggested_domain,omitempty"`
	RiskScore       int    `json:"risk_score"`
	Recommendation  string `json:"recommendation"`
}

var domainTypos = map[string]string{
	"gmial.com":  "gmail.com",
	"gamil.com":  "gmail.com",
	"gmal.com":   "gmail.com",
	"yaho.com":   "yahoo.com",
	"outlok.com": "outlook.com",
}

var rolePrefixes = map[string]bool{
	"admin": true, "support": true, "info": true,
	"sales": true, "billing": true, "contact": true,
	"help": true, "office": true, "hr": true,
}

// checkMX melakukan kueri DNS dengan batas waktu maksimal 1.5 detik
func checkMX(domain string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 1500*time.Millisecond)
	defer cancel()

	mxRecords, err := net.DefaultResolver.LookupMX(ctx, domain)
	return err == nil && len(mxRecords) != 0
}

func Analyze(rawEmail string) Result {
	email := strings.TrimSpace(strings.ToLower(rawEmail))

	res := Result{
		Email:          email,
		RiskScore:      0,
		Recommendation: "allow",
	}

	// 1. Validasi Sintaks Dasar
	_, err := mail.ParseAddress(email)
	parts := strings.Split(email, "@")
	if err != nil || len(parts) != 2 || !strings.Contains(parts[1], ".") {
		res.IsValidSyntax = false
		res.RiskScore = 100
		res.Recommendation = "block"
		return res
	}

	res.IsValidSyntax = true
	res.Username = parts[0]
	res.Domain = parts[1]

	// 2. Cek Typo Domain terlebih dahulu
	if correctDomain, exists := domainTypos[res.Domain]; exists {
		res.SuggestedDomain = correctDomain
	}

	// 3. JALANKAN PENGECEKAN SECARA PARALEL DENGAN GOROUTINES
	var wg sync.WaitGroup
	var isRole, isDisp, hasMX bool

	// Kita akan menjalankan 3 tugas secara bersamaan
	wg.Add(3)

	// Goroutine 1: Cek Role Account
	go func() {
		defer wg.Done()
		isRole = rolePrefixes[res.Username]
	}()

	// Goroutine 2: Cek Temp-Mail / Disposable Domain
	go func() {
		defer wg.Done()
		isDisp = IsDisposable(res.Domain)
	}()

	// Goroutine 3: Cek DNS MX Record ke jaringan internet
	go func() {
		defer wg.Done()
		hasMX = checkMX(res.Domain)
	}()

	// Tunggu sampai ketiga Goroutine di atas selesai!
	wg.Wait()

	// 4. Gabungkan hasil & hitung skor risiko (Risk Score)
	res.IsRoleAccount = isRole
	res.IsDisposable = isDisp
	res.HasMXRecord = hasMX

	if res.IsDisposable {
		res.RiskScore += 85
	}
	if !res.HasMXRecord {
		res.RiskScore += 60
	}
	if res.SuggestedDomain != "" {
		res.RiskScore += 40
	}
	if res.IsRoleAccount {
		res.RiskScore += 15
	}

	// Batasi skor maksimal di angka 100
	if res.RiskScore >= 100 {
		res.RiskScore = 100
	}

	// Tentukan keputusan akhir (Recommendation)
	if res.IsDisposable || !res.HasMXRecord || res.RiskScore >= 70 {
		res.Recommendation = "block"
	} else if res.SuggestedDomain != "" || res.RiskScore >= 30 {
		res.Recommendation = "review"
	}

	return res
}