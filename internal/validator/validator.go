package validator

import (
	"net/mail"
	"strings"
)

// Result adalah cetakan (Struct) untuk output JSON kita nanti.
// Huruf awal nama field WAJIB Kapital (Public/Exported) agar bisa diubah jadi JSON!
type Result struct {
	Email             string `json:"email"`
	IsValidSyntax     bool   `json:"is_valid_syntax"`
	Username          string `json:"username,omitempty"`
	Domain            string `json:"domain,omitempty"`
	IsRoleAccount     bool   `json:"is_role_account"`
	SuggestedDomain   string `json:"suggested_domain,omitempty"`
	RiskScore         int    `json:"risk_score"` // 0 (Aman) - 100 (Berbahaya)
	Recommendation    string `json:"recommendation"`
}

// Kamus typo domain populer
var domainTypos = map[string]string{
	"gmial.com":  "gmail.com",
	"gamil.com":  "gmail.com",
	"gmal.com":   "gmail.com",
	"yaho.com":   "yahoo.com",
	"outlok.com": "outlook.com",
}

// Daftar kata kunci email grup/kantor (bukan personal)
var rolePrefixes = map[string]bool{
	"admin": true, "support": true, "info": true,
	"sales": true, "billing": true, "contact": true,
	"help": true, "office": true, "hr": true,
}

// Analyze adalah fungsi utama yang akan dipanggil oleh HTTP Handler
func Analyze(rawEmail string) Result {
	email := strings.TrimSpace(strings.ToLower(rawEmail))

	res := Result{
		Email:          email,
		RiskScore:      0,
		Recommendation: "allow",
	}

	// 1. Cek format email menggunakan standard library net/mail
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

	// 2. Cek apakah email ini akun grup/kantor (Role Account)
	if rolePrefixes[res.Username] {
		res.IsRoleAccount = true
		res.RiskScore += 20
	}

	// 3. Cek apakah ada typo di domain
	if correctDomain, exists := domainTypos[res.Domain]; exists {
		res.SuggestedDomain = correctDomain
		res.RiskScore += 50
		res.Recommendation = "flag_typo"
	}

	return res
}