package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
)

// 1. Definisikan Struktur Data di Luar
type Stats struct {
	CPU    float64 `json:"cpu"`
	RAM    float64 `json:"ram"`
	Disk   float64 `json:"disk"`
	NetIn  uint64  `json:"net_in"`
	NetOut uint64  `json:"net_out"`
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

	fmt.Println("Client terhubung!")

	for {
		// Panggil fungsi pengambil data tadi
		stats := getSystemStats()

		// Kirim data
		err := ws.WriteJSON(stats)
		if err != nil {
			break // Keluar loop jika client putus
		}

		time.Sleep(1 * time.Second)
	}
}

// 5. FUNGSI UTAMA (ENTRY POINT)
func main() {
	http.Handle("/", http.FileServer(http.Dir("./public")))
	http.HandleFunc("/ws", handleConnections)

	fmt.Println("Server berjalan di http://localhost:8086")
	http.ListenAndServe(":8086", nil)
}
