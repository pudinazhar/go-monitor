# APLIKASI MONITORING SEDERHANA
### Membuat Binary dengan kode Golang dan menjalakanya di server ubuntu
Kode main.go yang ada di root folder adalah khusus untuk saya membuat/create menjadi binary<br/>
Jika Kalian ingin menjalankan di golang, maka gunakan main.go yang ada di dalam folder cmd/main.go

## Tampilan Aplikasi
![Tampilan Aplikasi](public/img/thumbnail.jpg)

## Pembuatan Binary
Build Binary :
```
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o gomonitor
```
Build Binary di Windows Powershwll
```
$env:CGO_ENABLED="0"; $env:GOOS="linux"; $env:GOARCH="amd64"; go build -trimpath -ldflags="-s -w" -o gomonitor
```

## Runing di server
Struktur Server
```
/opt/gomonitor
├── gomonitor
├── .env
├── db/
│   └── go-monitor.sqlite
└── logs/
```
Setelah di Upload file binnary ke Server
Jangan lupa:
```
chmod +x /opt/gomonitor/gomonitor
```
Tes jalan:
```
./gomonitor
```
Jika tidak ada error, kita akan lanjut,

Cara yang benar:
system user, Buat user khusus untuk gomonitor
Di Ubuntu/Linux, gunakan:
```
sudo useradd --system --no-create-home --shell /usr/sbin/nologin gomonitor
```
Hasilnya
User: gomonitor

Set ownership ke user ini
```
sudo chown -R gomonitor:gomonitor /opt/gomonitor
```
systemd configuration
Ini bagian penting:
Buat Sytemd
```
sudo nano /etc/systemd/system/gomonitor.service
```
Paste kode dibawah ini :
```
[Unit]
Description=Go Monitor
After=network.target

[Service]
User=gomonitor
Group=gomonitor
WorkingDirectory=/opt/gomonitor
ExecStart=/opt/gomonitor/gomonitor

Restart=always
RestartSec=5

NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=full
ProtectHome=true
RestartSec=3

Environment=APP_ENV=production
Environment=GIN_MODE=release
Environment=PDN_PORT=8086

[Install]
WantedBy=multi-user.target
```

Untuk SQLite kamu
SQLite tetap aman, tapi pastikan:
```
chown -R gomonitor:gomonitor /opt/gomonitor/db
```
Jalankan Service
```
sudo systemctl daemon-reload
```
```
sudo systemctl enable gomonitor
```
```
sudo systemctl start gomonitor
```
Cek status:
```
sudo systemctl status gomonitor
```
Restart
```
sudo systemctl restart gomonitor
```
Lihat log realtime:
```
journalctl -u gomonitor -f
```

## Konfigurasi dengan Nginx
Buat file
```
sudo nano /etc/nginx/sites-available/gomonitor
```
Isi dengan :
```
server {
    listen 80;
    server_name monitor.italazhar.com;

    # redirect HTTP → HTTPS
    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl;
    server_name monitor.italazhar.com;
    # Saya menggunkan SSL dari CLoudflare
    ssl_certificate     /opt/cloudflare/italazhar.com.pem;
    ssl_certificate_key /opt/cloudflare/italazhar.com.key;

    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_prefer_server_ciphers on;

    location / {
        proxy_pass http://127.0.0.1:8086;

        proxy_http_version 1.1;

        # WebSocket support (WAJIB untuk app kamu)
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";

        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto https;
    }
}
```
Tes Konfigurasi Nginx
```
sudo nginx -t
```
Jangan sampai ada error

Kita Aaktifkan konfigurasi nginx yang tadi kita buat
```
ln -s /etc/nginx/sites-available/gomonitor /etc/nginx/sites-enabled/.
```

Tes Konfigurasi Nginx
```
sudo nginx -t
```
Jangan sampai ada error


Restart Nginx
```
sudo service nginx restart
```

### Pudin Saepudin
- https://italazhar.com
- https://github.com/pudintea
- https://t.me/pudin_ira
- https://hub.docker.com/r/pudinalazhar