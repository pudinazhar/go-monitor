package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	_ "modernc.org/sqlite"
)

// 1. Definisikan Struktur Data di Luar
type Stats struct {
	CPU       float64 `json:"cpu"`
	CPUCores  int     `json:"cpu_cores"` // Jumlah Core
	RAM       float64 `json:"ram"`
	RAMTotal  float64 `json:"ram_total"` // Total RAM (GB)
	Disk      float64 `json:"disk"`
	DiskTotal float64 `json:"disk_total"` // Total Disk (GB)
	NetIn     uint64  `json:"net_in"`
	NetOut    uint64  `json:"net_out"`
	Time      string  `json:"time"`
}

// 2. Variabel Konfigurasi
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// 3. FUNGSI PENGAMBIL DATA (Taruh di sini agar rapi)
func getSystemStats() Stats {
	c, _ := cpu.Percent(time.Second, false)
	m, _ := mem.VirtualMemory()
	d, _ := disk.Usage("/")
	n, _ := net.IOCounters(false)

	// Pastikan ada pengecekan jika n kosong untuk menghindari crash
	var in, out uint64
	if len(n) > 0 {
		in = n[0].BytesRecv
		out = n[0].BytesSent
	}

	// Hitung total core
	cores, _ := cpu.Counts(true)

	return Stats{
		CPU:       c[0],
		CPUCores:  cores,
		RAM:       m.UsedPercent,
		RAMTotal:  float64(m.Total) / 1024 / 1024 / 1024, // Convert ke GB
		Disk:      d.UsedPercent,
		DiskTotal: float64(d.Total) / 1024 / 1024 / 1024, // Convert ke GB
		NetIn:     in,
		NetOut:    out,
		Time:      time.Now().Format("15:04:05"),
	}
}

// 4. FUNGSI HANDLER WEBSOCKET
func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer ws.Close()

	ticker := 0
	for {
		stats := getSystemStats()

		// Simpan ke DB setiap 4 detik (4 kali iterasi loop jika sleep 1s)
		ticker++
		if ticker >= 4 {
			saveToDB(stats.CPU, stats.RAM)
			ticker = 0
		}

		err := ws.WriteJSON(stats)
		if err != nil {
			break
		}
		time.Sleep(1 * time.Second)
	}
}

var db *sql.DB

func initDB() {
	// 1. Buat folder "db" jika belum ada
	if _, err := os.Stat("db"); os.IsNotExist(err) {
		err := os.Mkdir("db", 0755)
		if err != nil {
			panic("Gagal membuat folder db: " + err.Error())
		}
	}

	var err error
	// 2. Ubah path ke db/go-monitor.sqlite
	// Pastikan menggunakan driver "sqlite" (modernc) dan parameter WAL agar tidak locked
	path := "db/go-monitor.sqlite?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)"

	db, err = sql.Open("sqlite", path)
	if err != nil {
		panic(err)
	}

	db.SetMaxOpenConns(1)

	// Buat tabel jika belum ada
	sqlStmt := `
	CREATE TABLE IF NOT EXISTS stats (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		cpu REAL,
		ram REAL,
		time DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		panic(err)
	}
}

// Fungsi untuk menyimpan data
func saveToDB(cpu, ram float64) {
	// Gunakan format "2006-01-02 15:04:05" (ini adalah layout standar di Go)
	currentTime := time.Now().Format("2006-01-02 15:04:05")
	_, err := db.Exec("INSERT INTO stats (cpu, ram, time) VALUES (?, ?, ?)", cpu, ram, currentTime)
	if err != nil {
		fmt.Println("Gagal simpan ke DB:", err)
	}
}

// Fungsi pembersih data > 1 jam
func cleanupDB() {
	for {
		time.Sleep(10 * time.Minute)
		// Gunakan "time" sesuai nama kolom di tabel kamu
		_, err := db.Exec(`DELETE FROM stats WHERE "time" < datetime('now', '-1 hour', 'localtime')`)
		if err != nil {
			fmt.Println("Gagal cleanup database:", err)
		}
	}
}

// Endpoint baru untuk mengambil data 1 jam terakhir
func getHistoryHandler(w http.ResponseWriter, r *http.Request) {
	// Kita gunakan "time" (nama kolom kamu)
	// Kita ambil jam:menit:detik menggunakan substr karena format kamu string lengkap
	query := `
        SELECT cpu, ram, substr("time", 12, 8) 
        FROM stats 
        WHERE "time" > datetime('now', '-1 hour', 'localtime') 
        ORDER BY "time" ASC`

	rows, err := db.Query(query)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer rows.Close()

	var history []Stats
	for rows.Next() {
		var s Stats
		// Scan data: cpu -> s.CPU, ram -> s.RAM, substr -> s.Time
		err := rows.Scan(&s.CPU, &s.RAM, &s.Time)
		if err != nil {
			fmt.Println("Error scan:", err)
			continue
		}
		history = append(history, s)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}

// 5. FUNGSI UTAMA (ENTRY POINT)
func main() {
	initDB()
	go cleanupDB()

	http.Handle("/", http.FileServer(http.Dir("./public")))
	http.HandleFunc("/ws", handleConnections)
	http.HandleFunc("/history", getHistoryHandler)
	fmt.Println("Server berjalan di port :8086")
	http.ListenAndServe(":8086", nil)
}
