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

	client := http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(url)
	
	duration := time.Since(startTime)
	milliseconds := duration.Milliseconds()

	if err != nil {
		return ProbeResult{
			StatusCode:  0, 
			LatencyMs:   milliseconds,
			NetworkErr:  true,
		}
	}
	defer resp.Body.Close()

	return ProbeResult{
		StatusCode:  resp.StatusCode,
		LatencyMs:   milliseconds,
		NetworkErr:  false,
	}
}
