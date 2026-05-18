# Go-Pulse Agent

A lightweight, zero-dependency system monitoring agent written in Go. It collects hardware metrics (CPU, RAM, Disk, OS, and active IP address) and reports them back to a central Go-Pulse Collector server on a configurable interval.

## Installation & Setup

### 1. Prerequisites

Ensure you have Go installed on the target machine (version 1.22+ recommended).

### 2. Clone and Build the Binary

Clone the repository to your server and compile the standalone binary:

```bash
git clone https://github.com/yourusername/go-pulse-agent.git
cd go-pulse-agent
go mod tidy
go build -o go-pulse-agent main.go
```

### 3. Configuration & Execution

The agent is configured entirely via environment variables passed inline at startup. Run the agent by executing the following command:

```bash
SERVER_URL="http://<YOUR_SERVER_IP>:8080/api/report" REPORT_INTERVAL="6h" ./go-pulse-agent
```

#### Configuration Options

| Variable | Description |
|---|---|
| `SERVER_URL` | The full HTTP endpoint of your central collector server (e.g., `http://192.168.1.50:8080/api/report`). |
| `REPORT_INTERVAL` | How often the agent sweeps and reports metrics. Supports standard duration strings like `10s` (10 seconds), `30m` (30 minutes), or `6h` (6 hours). |

#### Alternative: Using a `.env` File

If you prefer not to pass variables inline, you can create a file named `.env` in the same folder as your binary:

```env
SERVER_URL=http://<YOUR_SERVER_IP>:8080/api/report
REPORT_INTERVAL=6h
```

Once the file is saved, you can simply launch the agent with:

```bash
./go-pulse-agent
```