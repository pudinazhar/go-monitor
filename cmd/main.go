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
	CPU    float64 `json:"cpu"`
	RAM    float64 `json:"ram"`
	Disk   float64 `json:"disk"`
	NetIn  uint64  `json:"net_in"`
	NetOut uint64  `json:"net_out"`
	Time   string  `json:"time"`
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

	return Stats{
		CPU:    c[0],
		RAM:    m.UsedPercent,
		Disk:   d.UsedPercent,
		NetIn:  in,
		NetOut: out,
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
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		panic(err)
	}
}

// Fungsi untuk menyimpan data
func saveToDB(cpu, ram float64) {
	_, err := db.Exec("INSERT INTO stats (cpu, ram, created_at) VALUES (?, ?, ?)", cpu, ram, time.Now())
	if err != nil {
		fmt.Println("Gagal simpan ke DB:", err)
	}
}

// Fungsi pembersih data > 1 jam
func cleanupDB() {
	for {
		_, err := db.Exec("DELETE FROM stats WHERE created_at < datetime('now', '-1 hour')")
		if err != nil {
			fmt.Println("Gagal cleanup:", err)
		}
		time.Sleep(10 * time.Minute) // Cek setiap 10 menit
	}
}

// Endpoint baru untuk mengambil data 1 jam terakhir
func getHistoryHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT cpu, ram FROM stats WHERE created_at > datetime('now', '-1 hour') ORDER BY created_at ASC")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer rows.Close()

	var history []Stats
	for rows.Next() {
		var s Stats
		rows.Scan(&s.CPU, &s.RAM)
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
	fmt.Println("Server berjalan di http://localhost:8086")
	http.ListenAndServe(":8086", nil)
}
