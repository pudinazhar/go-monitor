// Tahun 
document.getElementById('year').innerText = new Date().getFullYear();
const badge = document.getElementById('status-badge');

// Konfigurasi Dasar Grafik
const chartOptions = {
    responsive: true,
    maintainAspectRatio: false,
    scales: {
        y: { beginAtZero: true, max: 100, grid: { color: '#374151' } },
        x: { display: false }
    },
    plugins: { legend: { display: false } },
    elements: { line: { tension: 0.4 }, point: { radius: 0 } },
    animation: { duration: 500 }
};

// Inisialisasi Chart Disk (Doughnut)
const diskCtx = document.getElementById('diskChart').getContext('2d');
const diskChart = new Chart(diskCtx, {
    type: 'doughnut',
    data: {
        labels: ['Used', 'Free'],
        datasets: [{
            data: [0, 100], // Default awal
            backgroundColor: ['#60a5fa', 'rgba(96, 165, 250, 0.2)'], // Biru terang
            borderWidth: 0,
            hoverOffset: 4
        }]
    },
    options: {
        responsive: true,
        maintainAspectRatio: false,
        cutout: '70%', // Membuat lubang di tengah lebih besar (lebih modern)
        plugins: {
            legend: { display: false } // Sembunyikan legenda agar bersih
        }
    }
});

// Inisialisasi Chart CPU
const cpuCtx = document.getElementById('cpuChart').getContext('2d');
const cpuChart = new Chart(cpuCtx, {
    type: 'line',
    data: {
        labels: Array(20).fill(''),
        datasets: [{
            label: 'CPU',
            data: Array(20).fill(0),
            borderColor: '#60a5fa',
            backgroundColor: 'rgba(96, 165, 250, 0.2)',
            fill: true
        }]
    },
    options: chartOptions
});

// Inisialisasi Chart RAM
const ramCtx = document.getElementById('ramChart').getContext('2d');
const ramChart = new Chart(ramCtx, {
    type: 'line',
    data: {
        labels: Array(20).fill(''),
        datasets: [{
            label: 'RAM',
            data: Array(20).fill(0),
            borderColor: '#a855f7',
            backgroundColor: 'rgba(168, 85, 247, 0.2)',
            fill: true
        }]
    },
    options: chartOptions
});

function connect() {
    // 1. Deteksi protokol: jika https gunakan wss, jika http gunakan ws
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    
    // 2. Gabungkan protokol dengan host yang sedang aktif
    const socketUrl = `${protocol}//${window.location.host}/ws`;
    
    const socket = new WebSocket(socketUrl);
    const badge = document.getElementById('status-badge');

    socket.onopen = () => {
        console.log("✅ Terhubung!");
        // Update tampilan badge saat Online
        badge.innerText = "● Live System";
        badge.className = "px-3 py-1 rounded-full bg-green-900/30 text-green-400 text-xs border border-green-800 transition-all duration-500";
    };

    socket.onmessage = (event) => {
        const data = JSON.parse(event.data);
    
        // ... kode update CPU & RAM ...
        document.getElementById('disk-val').innerText = data.disk.toFixed(1) + '%';

        // Update Grafik Disk
        const usedPercent = data.disk;
        const freePercent = 100 - data.disk;
        
        diskChart.data.datasets[0].data = [usedPercent, freePercent];
        diskChart.update(); // Update grafik bulat
        
        document.getElementById('cpu-val').innerText = data.cpu.toFixed(1) + '%';
        document.getElementById('ram-val').innerText = data.ram.toFixed(1) + '%';
        
        const mbIn = (data.net_in / 1024 / 1024).toFixed(2);
        const mbOut = (data.net_out / 1024 / 1024).toFixed(2);
        
        // Update ID baru
        document.getElementById('net-in').innerText = `${mbIn} MB`;
        document.getElementById('net-out').innerText = `${mbOut} MB`;

        // Update Charts
        cpuChart.data.datasets[0].data.push(data.cpu);
        cpuChart.data.datasets[0].data.shift();
        cpuChart.update('none');

        ramChart.data.datasets[0].data.push(data.ram);
        ramChart.data.datasets[0].data.shift();
        ramChart.update('none');
    };

    socket.onclose = () => {
        console.log("❌ Terputus!");
        // Update tampilan badge saat Offline/Reconnecting
        badge.innerText = "○ Disconnected - Reconnecting...";
        badge.className = "px-3 py-1 rounded-full bg-red-900/30 text-red-400 text-xs border border-red-800 animate-pulse";
        
        setTimeout(connect, 2000);
    };
}

connect();