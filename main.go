package main

import (
	"html/template"
	"log"
	"net/http"
	"test/database" // Sesuaikan 'test' dengan nama modul Anda
	"test/handler"  // Sesuaikan 'test'
	"test/scheduler" // Sesuaikan 'test'

	"github.com/gorilla/mux"
)

// main adalah titik masuk aplikasi.
func main() {
	// 1. Inisialisasi Database
	store := database.NewStore("probe.db")
	log.Println("Database terhubung dan tabel siap.")

	// 2. Muat SEMUA Template HTML
	// Inilah perbaikannya. ParseGlob memuat semua file .html di folder templates.
	tpl, err := template.ParseGlob("templates/*.html")
	if err != nil {
		log.Fatalf("Gagal memuat template (ParseGlob): %v", err)
	}

	// 3. Ambil interval awal dari DB
	initialInterval, err := store.GetScheduleInterval()
	if err != nil {
		log.Fatalf("Gagal mengambil interval awal: %v", err)
	}

	// 4. Buat struct 'app'
	app := &handler.Application{
		Store:     store,
		Templates: tpl,
	}

	// 5. Mulai Scheduler dan simpan state-nya ke 'app'
	app.Scheduler, app.JobID = scheduler.StartScheduler(initialInterval, app.Store)

	// 6. Setup Handlers (kirim 'app' ke handlers)
	h := handler.NewHandlers(app)
	r := mux.NewRouter()

	// Routing untuk Halaman
	r.HandleFunc("/", h.DashboardPage).Methods("GET")
	r.HandleFunc("/urls", h.URLsPage).Methods("GET")
	r.HandleFunc("/scheduler", h.SchedulerPage).Methods("GET")

	// Routing untuk Aksi (POST/GET)
	r.HandleFunc("/add", h.AddURL).Methods("POST")
	r.HandleFunc("/delete/{id:[0-9]+}", h.DeleteURL).Methods("GET")
	r.HandleFunc("/settings", h.UpdateSettings).Methods("POST")

	// Routing untuk file statis (CSS, JS, Gambar)
	fs := http.FileServer(http.Dir("./static/"))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))

	// 7. Jalankan Web Server
	port := ":8080"
	log.Printf("Server Merah Putih berjalan di http://localhost%s\n", port)
	log.Fatal(http.ListenAndServe(port, r))
}

