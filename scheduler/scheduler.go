package scheduler

import (
	"database/sql"
	"log"
	"test/database"
	"test/probe"
	"time"

	"github.com/robfig/cron/v3"
)

// CreateJob adalah fungsi yang mengembalikan fungsi job
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
				newFirstUpTime = sql.NullTime{Time: time.Now(), Valid: true}
			} else if wasUp && !isNowUp {
				newFirstUpTime = sql.NullTime{Time: time.Time{}, Valid: false}
			}

			if result.StatusCode > 0 {
				err = store.UpdateProbeStats(u.ID, result.StatusCode, result.LatencyMs, newFirstUpTime)
				if err == nil {
					err = store.AddProbeHistory(u.ID, result.LatencyMs)
				}
			} else {
				err = store.UpdateProbeNetworkError(u.ID, result.LatencyMs, newFirstUpTime)
			}

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
	c.Start()

	return c, id
}
