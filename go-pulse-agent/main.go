package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
)

type ServerReport struct {
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

func main() {
	// Load the .env file if it exists.
	// If it doesn't exist, it safely skips it and reads system variables.
	if err := godotenv.Load(); err != nil {
		log.Println("No local .env file found, relying on system environment variables")
	}

	serverURL := getEnv("SERVER_URL", "http://localhost:8080/api/report")
	intervalStr := getEnv("REPORT_INTERVAL", "6h")

	interval, err := time.ParseDuration(intervalStr)
	if err != nil {
		log.Fatalf("Invalid REPORT_INTERVAL layout: %v", err)
	}

	log.Printf("Go-Pulse Agent started. Reporting to %s every %s", serverURL, interval)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Initial check-in right away on boot
	sendMetrics(serverURL)

	for range ticker.C {
		sendMetrics(serverURL)
	}
}

func sendMetrics(url string) {
	log.Println("Gathering system metrics...")
	report, err := collectMetrics()
	if err != nil {
		log.Printf("Error gathering hardware statistics: %v", err)
		return
	}

	jsonData, err := json.Marshal(report)
	if err != nil {
		log.Printf("Error processing payload JSON: %v", err)
		return
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Error sending report to server: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		log.Printf("Server rejected metrics payload with status: %s", resp.Status)
		return
	}

	log.Println("Metrics successfully delivered to central collector")
}

func collectMetrics() (ServerReport, error) {
	var r ServerReport
	r.Timestamp = time.Now()

	hInfo, err := host.Info()
	if err != nil {
		return r, err
	}
	r.Hostname = hInfo.Hostname
	r.OS = hInfo.OS
	r.Platform = hInfo.Platform

	r.IPAddress = getLocalIP()

	vMem, err := mem.VirtualMemory()
	if err != nil {
		return r, err
	}
	r.RAMTotal = vMem.Total
	r.RAMFree = vMem.Available

	dUsage, err := disk.Usage("/")
	if err != nil {
		return r, err
	}
	r.DiskTotal = dUsage.Total
	r.DiskFree = dUsage.Free

	cInfo, err := cpu.Info()
	if err == nil && len(cInfo) > 0 {
		r.CPUModel = cInfo[0].ModelName
	} else {
		r.CPUModel = "Unknown CPU"
	}

	cores, err := cpu.Counts(true)
	if err == nil {
		r.CPUCount = cores
	} else {
		r.CPUCount = 1
	}

	return r, nil
}

func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "127.0.0.1"
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return "127.0.0.1"
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
