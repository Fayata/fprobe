package handler

import (
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"test/database" // PERBAIKAN: Dihapus 'internal/'
	"test/models"    // PERBAIKAN: Dihapus 'internal/'
	"test/scheduler" // PERBAIKAN: Dihapus 'internal/'
	"time"

	"github.com/gorilla/mux"
	"github.com/robfig/cron/v3"
)

// Application adalah struct utama yang menampung semua state aplikasi
type Application struct {
	Store     *database.Store
	Templates *template.Template
	Scheduler *cron.Cron
	JobID     cron.EntryID
}

// Handlers adalah struct untuk menampung App
type Handlers struct {
	App *Application
}

// NewHandlers membuat struct Handlers baru
func NewHandlers(app *Application) *Handlers {
	return &Handlers{App: app}
}

// === HANDLER HALAMAN ===

// DashboardPage menangani halaman utama ('/')
func (h *Handlers) DashboardPage(w http.ResponseWriter, r *http.Request) {
	// PERBAIKAN: Mencegah error 'superfluous response.WriteHeader'
	// Hanya proses request untuk path root '/'
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	// 1. Ambil semua URL (untuk dropdown chart dan stats)
	urls, err := h.App.Store.GetAllURLs()
	if err != nil {
		log.Printf("Gagal mengambil URL: %v", err)
		http.Error(w, "Gagal mengambil data", http.StatusInternalServerError)
		return
	}

	// 2. Ambil URL yang dipilih dari query param
	selectedURLIDStr := r.URL.Query().Get("url_id")
	selectedID, _ := strconv.Atoi(selectedURLIDStr) // Konversi ke int

	// Jika tidak ada ID, atau ID tidak valid, pakai ID pertama dari daftar
	if selectedID == 0 && len(urls) > 0 {
		selectedID = urls[0].ID
	}

	// 3. Ambil data history probe (untuk chart)
	var historyData []models.ProbeHistory
	if selectedID > 0 { // Hanya ambil jika ada ID yang valid
		historyData, err = h.App.Store.GetProbeHistory(selectedID, 30) // Ambil history untuk ID yang dipilih
		if err != nil {
			log.Printf("Gagal mengambil data history: %v", err)
		}
	}

	// 4. Siapkan PageData untuk dikirim ke template
	data := models.PageData{
		Page:             "dashboard",
		URLs:             urls,
		GlobalAvgLatency: calculateGlobalAvgLatency(urls),
		LastCheckedTime:  getLatestProbeTime(urls),
		HistoryData:      historyData,
		SelectedURLID:    selectedID, // Kirim ID yang dipilih ke template
	}

	// 5. Render template
	// 5. Render template
	err = h.App.Templates.ExecuteTemplate(w, "layout", data)
	if err != nil {
    	log.Printf("Error rendering template: %v", err)
    	return
	}
}

// URLsPage menangani halaman '/urls'
func (h *Handlers) URLsPage(w http.ResponseWriter, r *http.Request) {
	urls, err := h.App.Store.GetAllURLs()
	if err != nil {
		log.Printf("Gagal mengambil URL: %v", err)
		http.Error(w, "Gagal mengambil data", http.StatusInternalServerError)
		return
	}

	data := models.PageData{
		Page:            "urls",
		URLs:            urls,
		LastCheckedTime: getLatestProbeTime(urls),
	}

	err = h.App.Templates.ExecuteTemplate(w, "layout", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// SchedulerPage menangani halaman '/scheduler'
func (h *Handlers) SchedulerPage(w http.ResponseWriter, r *http.Request) {
	// 1. Ambil interval saat ini
	interval, err := h.App.Store.GetScheduleInterval()
	if err != nil {
		log.Printf("Gagal mengambil interval: %v", err)
		http.Error(w, "Gagal mengambil data", http.StatusInternalServerError)
		return
	}
	urls, _ := h.App.Store.GetAllURLs()

	// 2. AMBIL RIWAYAT PROBE TERBARU (FITUR BARU)
	historyData, err := h.App.Store.GetAllProbeHistory(50) // Ambil 50 log terakhir
	if err != nil {
		log.Printf("Gagal mengambil semua history: %v", err)
	}

	// 3. Siapkan PageData
	data := models.PageData{
		Page:            "scheduler",
		CurrentInterval: interval,
		LastCheckedTime: getLatestProbeTime(urls),
		HistoryData:     historyData, // Kirim data riwayat ke template
	}

	// 4. Render template
	err = h.App.Templates.ExecuteTemplate(w, "layout", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// === HANDLER AKSI (FORM) ===

// AddURL menangani form 'Tambah URL'
func (h *Handlers) AddURL(w http.ResponseWriter, r *http.Request) {
	url := r.FormValue("url")
	if url == "" {
		http.Redirect(w, r, "/urls", http.StatusSeeOther)
		return
	}
	if !((strings.HasPrefix(url, "http://")) || (strings.HasPrefix(url, "https://"))) {
		url = "https://" + url
	}
	err := h.App.Store.AddURL(url)
	if err != nil {
		log.Printf("Gagal menambah URL: %v", err)
	}
	http.Redirect(w, r, "/urls", http.StatusSeeOther)
}

// DeleteURL menangani link 'Hapus'
func (h *Handlers) DeleteURL(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID tidak valid", http.StatusBadRequest)
		return
	}

	// Hapus history dulu
	err = h.App.Store.DeleteProbeHistory(id)
	if err != nil {
		log.Printf("Gagal menghapus history URL: %v", err)
	}

	// Baru hapus URL-nya
	err = h.App.Store.DeleteURL(id)
	if err != nil {
		log.Printf("Gagal menghapus URL: %v", err)
	}
	http.Redirect(w, r, "/urls", http.StatusSeeOther)
}

// UpdateSettings menangani form 'Simpan Jadwal'
func (h *Handlers) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	interval := r.FormValue("interval")

	// Validasi input
	validIntervals := map[string]bool{
		"@every 1m":  true,
		"@every 5m":  true,
		"@every 10m": true,
		"@every 30m": true,
	}
	if !validIntervals[interval] {
		log.Println("Percobaan input interval tidak valid:", interval)
		http.Redirect(w, r, "/scheduler", http.StatusSeeOther)
		return
	}

	// Simpan ke DB
	err := h.App.Store.SetScheduleInterval(interval)
	if err != nil {
		log.Println("Gagal menyimpan interval:", err)
		http.Redirect(w, r, "/scheduler", http.StatusSeeOther)
		return
	}

	// Restart Cron Job
	log.Printf("Mengubah jadwal scheduler ke: %s", interval)
	h.App.Scheduler.Remove(h.App.JobID) // Hapus job lama
	newID, err := h.App.Scheduler.AddFunc(interval, scheduler.CreateJob(h.App.Store)) // Buat job baru
	if err != nil {
		log.Println("Gagal menambah job cron baru:", err)
	}
	h.App.JobID = newID // Simpan ID job baru

	http.Redirect(w, r, "/scheduler", http.StatusSeeOther)
}

// === FUNGSI HELPER ===

// calculateGlobalAvgLatency menghitung rata-rata dari semua URL
func calculateGlobalAvgLatency(urls []models.TargetURL) int64 {
	var totalSum, totalCount int64
	for _, u := range urls {
		totalSum += u.TotalLatencySum
		totalCount += u.TotalProbeCount
	}
	if totalCount == 0 {
		return 0
	}
	return totalSum / totalCount
}

// getLatestProbeTime mencari waktu probe terbaru dari semua URL
func getLatestProbeTime(urls []models.TargetURL) time.Time {
	var latest time.Time
	for _, u := range urls {
		if u.LastChecked.After(latest) {
			latest = u.LastChecked
		}
	}
	return latest
}

