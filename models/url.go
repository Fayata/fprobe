package models

import (
	"database/sql"
	"fmt"
	"time"
)

// TargetURL merepresentasikan satu URL yang akan kita probe
type TargetURL struct {
	ID              int
	URL             string
	LastStatus      int
	LastLatencyMs   int64
	LastChecked     time.Time
	IsUp            bool
	FirstUpTime     sql.NullTime
	TotalProbeCount int64
	TotalLatencySum int64
}

// ProbeHistory menyimpan satu titik data historis untuk grafik
type ProbeHistory struct {
	URLID     int
	URL       string // <-- TAMBAHAN: Untuk menampilkan nama URL di riwayat
	LatencyMs int64
	Timestamp time.Time
}

// PageData adalah struct untuk passing data ke semua template HTML
type PageData struct {
	Page             string
	URLs             []TargetURL
	CurrentInterval  string
	GlobalAvgLatency int64
	LastCheckedTime  time.Time
	HistoryData      []ProbeHistory
	SelectedURLID    int
}

// === FUNGSI HELPER UNTUK TEMPLATE ===

// GetUptime menghitung durasi uptime yang mudah dibaca
func (tu *TargetURL) GetUptime() string {
	if !tu.FirstUpTime.Valid {
		return "N/A"
	}
	duration := time.Since(tu.FirstUpTime.Time)

	days := int(duration.Hours() / 24)
	hours := int(duration.Hours()) % 24
	minutes := int(duration.Minutes()) % 60

	if days > 0 {
		return fmt.Sprintf("%d hari, %d jam", days, hours)
	}
	if hours > 0 {
		return fmt.Sprintf("%d jam, %d mnt", hours, minutes)
	}
	if minutes > 0 {
		return fmt.Sprintf("%d mnt", minutes)
	}
	if duration.Seconds() < 60 {
		return "Baru saja"
	}
	return "N/A"
}

// GetAverageLatency menghitung rata-rata latency (sebagai string)
func (tu *TargetURL) GetAverageLatency() string {
	if tu.TotalProbeCount == 0 {
		return "N/A"
	}
	avg := tu.TotalLatencySum / tu.TotalProbeCount
	return fmt.Sprintf("%d ms", avg)
}

