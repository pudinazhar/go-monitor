# APLIKASI MONITORING SEDERHANA

## Tampilan Aplikasi

![Tampilan Aplikasi](public/img/thumbnail.jpg)

## MENJALANKAN DENGAN GOLANG
Pertama
```
go mod tidy
```
Selanjutnya
```
go run cmd/main.go
```
Buka Browser
```
http://localhost:8086
```

## LANGSUNG MENGGUNAKAN DOCKER
Buat file:
docker-compose.yml

Masukan Script ini
```
services:
  app:
    image: pudinalazhar/go-monitoring:latest
    container_name: go-monitoring
    ports:
      - "8086:8086"
    restart: unless-stopped
    environment:
      - GIN_MODE=release
```

Aktifkan
```
docker compose up -d
```

Build No Cache
```
sudo docker compose build --no-cache && docker compose up -d
```

Log Monitoring
```
docker logs -f go-monitoring
```

Cara Update
```
docker-compose pull  && docker-compose up -d
```

### Pudin Saepudin
- https://italazhar.com
- https://github.com/pudintea
- https://t.me/pudin_ira