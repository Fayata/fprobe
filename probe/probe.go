package probe

import (
	"net/http"
	"time"
)

// ProbeResult mendefinisikan data untuk hasil satu kali probe.
type ProbeResult struct {
	StatusCode  int
	LatencyMs   int64
	NetworkErr  bool
}

// DoProbe menjalankan satu kali HTTP GET probe dan mengukur waktu.
func DoProbe(url string) ProbeResult {
	startTime := time.Now()

	// Buat client dengan timeout 5 detik
	client := http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(url)
	
	// Hitung durasi
	duration := time.Since(startTime)
	milliseconds := duration.Milliseconds()

	// Kasus 1: Error Jaringan (misal: DNS, timeout, koneksi ditolak)
	if err != nil {
		return ProbeResult{
			StatusCode:  0, // Kita beri status 0 untuk error jaringan
			LatencyMs:   milliseconds,
			NetworkErr:  true,
		}
	}
	defer resp.Body.Close()

	// Kasus 2: Request berhasil, server memberi balasan
	return ProbeResult{
		StatusCode:  resp.StatusCode,
		LatencyMs:   milliseconds,
		NetworkErr:  false,
	}
}
