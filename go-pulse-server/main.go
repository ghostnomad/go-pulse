package main

import (
	"database/sql"
	"encoding/json"
	"fmt" // <-- Added to fix formatting
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
)

type ServerReport struct {
	ID        int       `json:"id"`
	Hostname  string    `json:"hostname"`
	IPAddress string    `json:"ip_address"`
	Timestamp time.Time `json:"timestamp"`
	CPUModel  string    `json:"cpu_model"`
	CPUCount  int       `json:"cpu_count"`
	RAMTotal  uint64    `json:"ram_total"`
	RAMFree   uint64    `json:"ram_free"`
	DiskTotal uint64    `json:"disk_total"`
	DiskFree  uint64    `json:"disk_free"`
	OS        string    `json:"os"`
	Platform  string    `json:"platform"`
}

// Custom template helper methods for formatting units in the HTML view
var tmplFuncs = template.FuncMap{
	"bytesToGB": func(b uint64) string {
		gb := float64(b) / 1024 / 1024 / 1024
		return fmt.Sprintf("%.1f GB", gb)
	},
	"formatTime": func(t time.Time) string {
		return t.Format("2006-01-02 15:04:05")
	},
	"isStale": func(t time.Time) bool {
		// If the last check-in was longer than 12 hours ago, mark it as stale/offline
		// (11 hours expected interval + 1 hour grace period buffer)
		threshold := 11 * time.Hour
		return time.Since(t) > threshold
	},
}

var db *sql.DB

func main() {
	var err error
	db, err = sql.Open("sqlite3", "./pulse.db")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	if err := initSchema(); err != nil {
		log.Fatalf("Failed to initialize database schema: %v", err)
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Post("/api/report", handleReport)

	// Updated root URL route to display our dashboard
	r.Get("/", handleDashboard)

	port := getEnv("PORT", "8080")
	log.Printf("Collector server starting on port %s...", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func initSchema() error {
	query := `
    CREATE TABLE IF NOT EXISTS server_reports (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        hostname TEXT NOT NULL,
        ip_address TEXT,
        timestamp DATETIME NOT NULL,
        cpu_model TEXT,
        cpu_count INTEGER,
        ram_total INTEGER,
        ram_free INTEGER,
        disk_total INTEGER,
        disk_free INTEGER,
        os TEXT,
        platform TEXT
    );
    CREATE INDEX IF NOT EXISTS idx_hostname_time ON server_reports(hostname, timestamp DESC);
    `
	_, err := db.Exec(query)
	return err
}

func handleReport(w http.ResponseWriter, r *http.Request) {
	var report ServerReport
	err := json.NewDecoder(r.Body).Decode(&report)
	if err != nil {
		log.Printf("Error decoding JSON payload: %v", err)
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	query := `
    INSERT INTO server_reports (
        hostname, ip_address, timestamp, cpu_model, cpu_count, ram_total, ram_free, disk_total, disk_free, os, platform
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err = db.Exec(query,
		report.Hostname,
		report.IPAddress,
		report.Timestamp,
		report.CPUModel,
		report.CPUCount,
		report.RAMTotal,
		report.RAMFree,
		report.DiskTotal,
		report.DiskFree,
		report.OS,
		report.Platform,
	)

	if err != nil {
		log.Printf("Failed to save report to database: %v", err)
		http.Error(w, "Internal server database error", http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully saved metrics for host: %s (%s)", report.Hostname, report.IPAddress)
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(`{"status": "success"}`))
}

// Fetches the single latest record for every unique server hostname
// FIXED: Reads timestamps as text strings first to shield against underlying SQLite driver quirks
func getLatestServerStatuses() ([]ServerReport, error) {
	query := `
    SELECT id, hostname, ip_address, timestamp, cpu_model, cpu_count, ram_total, ram_free, disk_total, disk_free, os, platform
    FROM server_reports
    WHERE id IN (
        SELECT MAX(id) FROM server_reports GROUP BY hostname
    )
    ORDER BY hostname ASC`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reports []ServerReport
	for rows.Next() {
		var r ServerReport
		var timeStr string

		err := rows.Scan(
			&r.ID, &r.Hostname, &r.IPAddress, &timeStr, &r.CPUModel,
			&r.CPUCount, &r.RAMTotal, &r.RAMFree, &r.DiskTotal, &r.DiskFree, &r.OS, &r.Platform,
		)
		if err != nil {
			return nil, err
		}

		// Re-hydrate text representation back into a robust Go time.Time object
		parsedTime, parseErr := time.Parse(time.RFC3339, timeStr)
		if parseErr != nil {
			// Fallback formatting parsing profile
			parsedTime, _ = time.Parse("2006-01-02 15:04:05-07:00", timeStr)
		}
		r.Timestamp = parsedTime

		reports = append(reports, r)
	}
	return reports, nil
}

func handleDashboard(w http.ResponseWriter, r *http.Request) {
	statuses, err := getLatestServerStatuses()
	if err != nil {
		log.Printf("Database read error: %v", err)
		http.Error(w, "Failed to load dashboard data", http.StatusInternalServerError)
		return
	}

	// Parse and display the index.html view
	tmpl, err := template.New("index.html").Funcs(tmplFuncs).ParseFiles("index.html")
	if err != nil {
		log.Printf("Template compilation error: %v", err)
		http.Error(w, "Failed to parse dashboard template", http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, statuses)
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
