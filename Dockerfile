# Build stage
FROM golang:1.26-alpine AS builder

# Instal git jika diperlukan untuk download modul
RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Tambahkan CGO_ENABLED=0 agar binary statis dan ringan
RUN CGO_ENABLED=0 GOOS=linux go build -o main cmd/main.go

# Run stage
FROM alpine:latest

# Instal tzdata agar waktu di database sesuai (WIB)
RUN apk add --no-cache tzdata
ENV TZ=Asia/Jakarta

WORKDIR /app

COPY --from=builder /app/main .
COPY --from=builder /app/public ./public

# Kita tidak COPY folder db, tapi kita buat foldernya saja
RUN mkdir ./db

EXPOSE 8086

CMD ["./main"]