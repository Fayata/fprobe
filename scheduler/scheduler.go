package scheduler

import (
	"database/sql"
	"log"
	"test/database" // Ganti 'test' jika nama modul Anda berbeda
	"test/probe"    // Ganti 'test' jika nama modul Anda berbeda
	"time"

	"github.com/robfig/cron/v3"
)

// CreateJob adalah fungsi yang mengembalikan (return) fungsi job
// Ini agar kita bisa memanggil ulang logikanya saat restart
func CreateJob(store *database.Store) func() {
	return func() {
		log.Println("[CRON] Memulai probe...")

		urls, err := store.GetAllURLs()
		if err != nil {
			log.Printf("[CRON] Gagal mengambil URL: %v\n", err)
			return
		}

		if len(urls) == 0 {
			log.Println("[CRON] Tidak ada URL untuk di-probe.")
			return
		}

		// Jalankan probe untuk setiap URL
		for _, u := range urls {
			result := probe.DoProbe(u.URL)

			// --- LOGIKA UPTIME ---
			var newFirstUpTime sql.NullTime = u.FirstUpTime
			wasUp := (u.LastStatus == 200)
			isNowUp := (result.StatusCode == 200)

			if !wasUp && isNowUp {
				// Kasus 1: Baru saja UP (sebelumnya Down)
				newFirstUpTime = sql.NullTime{Time: time.Now(), Valid: true}
			} else if wasUp && !isNowUp {
				// Kasus 3: Baru saja DOWN (sebelumnya Up)
				newFirstUpTime = sql.NullTime{Time: time.Time{}, Valid: false} // Set jadi NULL
			}
			// --- END LOGIKA UPTIME ---

			// --- UBAH LOGIKA UPDATE DB DI SINI ---
			if result.StatusCode > 0 {
				// Dapat balasan HTTP (200, 404, 503, dll.)
				// Gunakan 'UpdateProbeStats' untuk update rata-rata
				err = store.UpdateProbeStats(u.ID, result.StatusCode, result.LatencyMs, newFirstUpTime)
				// Tambah ke history HANYA jika probe sukses (status > 0)
				if err == nil {
					// === INI PERBAIKANNYA ===
					// Mengganti 'InsertProbeHistory' menjadi 'AddProbeHistory'
					err = store.AddProbeHistory(u.ID, result.LatencyMs)
					// === END PERBAIKAN ===
				}
			} else {
				// Gagal (Network error, status code 0)
				// Gunakan 'UpdateProbeNetworkError' (TIDAK update rata-rata)
				err = store.UpdateProbeNetworkError(u.ID, result.LatencyMs, newFirstUpTime)
			}
			// --- END UBAHAN ---

			if err != nil {
				log.Printf("[CRON] Gagal update DB untuk %s: %v\n", u.URL, err)
			} else {
				log.Printf("[CRON] Probe %s -> Status: %d, Latency: %dms\n", u.URL, result.StatusCode, result.LatencyMs)
			}
		}
		log.Println("[CRON] Probe selesai.")
	}
}

// StartScheduler memulai cron job
func StartScheduler(interval string, store *database.Store) (*cron.Cron, cron.EntryID) {
	log.Printf("Menjalankan scheduler (setiap %s)...", interval)
	c := cron.New()

	// Gunakan 'interval' dari argumen
	id, _ := c.AddFunc(interval, CreateJob(store))

	// Mulai cron job di background
	c.Start()
	
	// Kembalikan objek cron dan ID job-nya
	return c, id
}

