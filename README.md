# 🛡️ Email & Domain Fraud Guard API (Go)

![Go Version](https://img.shields.io/badge/Go-1.22%2B-00ADD8?style=for-the-badge&logo=go)
![Architecture](https://img.shields.io/badge/Architecture-Modular%20Clean-success?style=for-the-badge)
![Dependencies](https://img.shields.io/badge/Dependencies-Zero%20(Stdlib%20Only)-blue?style=for-the-badge)
![Deployment](https://img.shields.io/badge/AWS%20EC2-Production%20Ready-FF9900?style=for-the-badge&logo=amazon-aws)

A high-performance, zero-dependency RESTful Micro-API built with **Go (Golang)** to detect disposable/temporary email addresses, verify real-time DNS `MX` records, catch domain typos, identify role-based accounts, and calculate user registration fraud risk scores in milliseconds.

🌐 **Live Production Status:** `[https://email-api.aklein.tech/health](https://email-api.aklein.tech/health)`  
🚀 **RapidAPI Marketplace:** Available on RapidAPI Hub (`Fast Email Verifier and Disposable Temp Mail Detector`)

---

## ✨ Key Engineering Highlights

This service is designed around idiomatic Go backend engineering principles, focusing on high throughput, concurrency safety, and an ultra-low memory footprint (~15 MB RAM in production):

* **100% Go Standard Library (Zero External Dependencies):** Built entirely using native packages (`net/http`, `net/mail`, `net`, `sync`, `context`, `bufio`), ensuring maximum stability, zero supply-chain vulnerabilities, and tiny binary sizes.
* **Concurrent Verification Pipeline (`Goroutines` & `sync.WaitGroup`):** Executes role-account detection, in-memory disposable domain lookup, and live DNS `MX` resolution concurrently for every incoming request.
* **Zero-Byte Memory Set Optimization (`map[string]struct{}`):** Loads **9,900+ active disposable email domains** into RAM using empty structs (`0 bytes` value allocation) for $O(1)$ nanosecond lookups.
* **Thread-Safe Background Auto-Updater (`sync.RWMutex` & Pointer Swap):** A background worker refreshes the disposable domain blocklist every 24 hours from open-source intelligence feeds without blocking ongoing read requests or risking concurrent map read/write panics.
* **Strict Network Timeouts (`context.WithTimeout`):** Enforces a strict `1500ms` deadline on external DNS resolvers to prevent goroutine leaks and cascading latency.
* **API Gateway Authentication Middleware:** Protects production endpoints by validating cryptographic `X-RapidAPI-Proxy-Secret` headers.

---

## 📂 Project Architecture

```text
email-guard-api/
├── Dockerfile                        # Multi-stage Docker build (<20MB Alpine production image)
├── .gitignore                        # Git ignore rules for compiled binaries and env files
├── go.mod                            # Go module definition
├── main.go                           # Application entrypoint, HTTP routing & server lifecycle
└── internal/
    ├── middleware/
    │   └── rapidapi.go               # RapidAPI Proxy Secret authentication middleware
    └── validator/
        ├── disposable.go             # Thread-safe in-memory blocklist & 24h background ticker
        └── validator.go              # Concurrent email analysis, DNS MX lookup & risk scoring
```

---

## 🔌 API Documentation

### 1. Verify Email & Fraud Risk Score
Analyzes an email address across multiple fraud vectors simultaneously.

* **Endpoint:** `GET /v1/verify`
* **Query Parameter:** `email` *(required, string)*

#### Example Request (via RapidAPI)
```bash
curl --request GET \
  --url 'https://fast-email-verifier-and-disposable-temp-mail-detector.p.rapidapi.com/v1/verify?email=admin%40gmial.com' \
  --header 'x-rapidapi-host: fast-email-verifier-and-disposable-temp-mail-detector.p.rapidapi.com' \
  --header 'x-rapidapi-key: YOUR_RAPIDAPI_KEY'
```

#### Example Response (`200 OK`)
```json
{
  "email": "admin@gmial.com",
  "is_valid_syntax": true,
  "username": "admin",
  "domain": "gmial.com",
  "is_role_account": true,
  "is_disposable": false,
  "has_mx_record": false,
  "suggested_domain": "gmail.com",
  "risk_score": 70,
  "recommendation": "flag_typo"
}
```

#### Decision Matrix (`recommendation` field)
| Recommendation | Risk Score Range | Meaning & Suggested Client Action |
| :--- | :--- | :--- |
| **`allow`** | `0 – 29` | Valid syntax, active mail server (`MX`), personal domain. Safe to allow registration. |
| **`flag_typo`** | `40 – 69` | Detected a common domain misspelling (e.g., `gmial.com`). Prompt user with `suggested_domain`. |
| **`review`** | `30 – 69` | Role-based account (`admin@`, `support@`) or moderate risk. Require secondary verification. |
| **`block`** | `70 – 100` | Disposable/temp-mail provider, invalid syntax, or missing `MX` records. Block registration. |

---

### 2. Service Health Check
Public monitoring endpoint returning server health and real-time in-memory database metrics.

* **Endpoint:** `GET /health`

#### Example Response (`200 OK`)
```json
{
  "loaded_temp_domains": 9914,
  "status": "healthy",
  "timestamp": "2026-10-22T03:38:56Z"
}
```

---

## 🛠️ Local Development & Production Deployment

### 1. Run Locally
```bash
git clone https://github.com/YOUR_GITHUB_USERNAME/email-guard-api.git
cd email-guard-api
go run main.go
```
*(Note: When `RAPIDAPI_PROXY_SECRET` is not set in local environment variables, the authentication middleware automatically allows local testing requests).*

Test in browser or terminal:
```bash
curl "http://localhost:8080/v1/verify?email=hacker@yopmail.com"
```

### 2. Cross-Compile Binary for Linux Production (AWS EC2)
Compile a static, stripped binary (`~6 MB`) directly from Windows/macOS/Linux:

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o email-guard main.go
```

### 3. Build & Run with Docker
```bash
docker build -t email-guard-api .
docker run -d -p 8080:8080 -e RAPIDAPI_PROXY_SECRET=your_secret_key email-guard-api
```

---

## 📜 License
MIT License. Built with Go.