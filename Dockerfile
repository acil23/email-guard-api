# ==========================================
# STAGE 1: Builder (Kompilasi Kode Go)
# ==========================================
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Salin file modul dan kode sumber
COPY go.mod ./
COPY . .

# Kompilasi menjadi satu file binary statis bernama "server"
# Flag -ldflags="-w -s" membuang simbol debug agar ukuran file menyusut drastis
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o server main.go

# ==========================================
# STAGE 2: Production Image (Super Ringan)
# ==========================================
FROM alpine:latest

# Wajib instal ca-certificates agar server Go bisa mengakses HTTPS (saat download daftar temp-mail)
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Ambil HANYA file binary "server" dari Stage 1 (kode sumber tidak ikut terbawa)
COPY --from=builder /app/server .

# Ekspos port default
EXPOSE 8080

# Jalankan binary
CMD ["./server"]