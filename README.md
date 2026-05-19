# Go-Pulse

Go-Pulse is a lightweight, self-hosted system monitoring ecosystem written in Go. It consists of two components: a remote **Agent** that collects hardware metrics from target machines, and a central **Collector Server** that ingests those metrics, stores them in an embedded SQLite database, and serves a web UI dashboard for visualization.

## Architecture Overview

```
[ Target Machine ]           [ Main Server ]
  go-pulse-agent   ───────►  go-pulse-server  ───► Web Dashboard
  (CPU, RAM, Disk,            (REST API + SQLite     http://<ip>:8080
   OS, IP metrics)             + Web UI)
```

## Components

### 🖥️ Go-Pulse Agent
A zero-dependency binary deployed on each machine you want to monitor. It collects hardware metrics (CPU, RAM, Disk, OS, and active IP address) and reports them to the Collector Server on a configurable interval.

- Configured via environment variables or a `.env` file
- Supports flexible reporting intervals (`10s`, `30m`, `6h`, etc.)
- **Repo:** `https://github.com/ghostnomad/go-pulse-agent`

### 📡 Go-Pulse Collector Server
The central hub of the ecosystem. It exposes a REST API to receive telemetry from agents, persists historical reports in an embedded SQLite database (`pulse.db`), and hosts a minimal web UI dashboard.

- Listens on port `8080` by default
- No external database required — SQLite is embedded
- **Repo:** `https://github.com/ghostnomad/go-pulse-server`

## Quick Start

### 1. Set Up the Collector Server

```bash
git clone https://github.com/ghostnomad/go-pulse-server.git
cd go-pulse-server
go mod tidy
go build -o go-pulse-server main.go
PORT=8080 ./go-pulse-server
```

Then open `http://<YOUR_SERVER_IP>:8080` in your browser to view the dashboard.

### 2. Deploy the Agent on Target Machines

```bash
git clone https://github.com/ghostnomad/go-pulse-agent.git
cd go-pulse-agent
go mod tidy
go build -o go-pulse-agent main.go
SERVER_URL="http://<YOUR_SERVER_IP>:8080/api/report" REPORT_INTERVAL="6h" ./go-pulse-agent
```

## Requirements

- Go 1.22+ on all machines running either component

## Usage

### Agent Configuration

The agent is configured via a `.env` file placed in the same directory as the binary. Create the file before running the agent:

```env
SERVER_URL=http://<YOUR_SERVER_IP>:8080/api/report
REPORT_INTERVAL=6h
```

| Variable | Description |
|---|---|
| `SERVER_URL` | The full HTTP endpoint of your Collector Server (e.g., `http://192.168.1.50:8080/api/report`). |
| `REPORT_INTERVAL` | How often the agent collects and reports metrics. Supports duration strings like `10s`, `30m`, or `6h`. |

Once the `.env` file is saved, start the agent with:

```bash
./go-pulse-agent
```

### Collector Server Configuration

The server reads its port from an environment variable. You can also define this in a `.env` file in the server's directory:

```env
PORT=8080
```

Then start the server with:

```bash
./go-pulse-server
```

The dashboard will be available at `http://<YOUR_SERVER_IP>:8080`.

## Motivation

Managing a homelab means keeping track of a growing number of machines, and existing monitoring solutions tend to be heavy, complex, or overkill for personal use. Go-Pulse was built to solve that — a simple, lightweight way to keep an eye on all assets in a homelab without the overhead of enterprise tooling. The goal was something easy to deploy, easy to understand, and easy to maintain.

## Contributors

| Name | Role |
|---|---|
| @ghostnomad | Creator & Maintainer |

Contributions are welcome! Feel free to open an issue or submit a pull request.

## Further Reading

- [Agent Setup & Configuration](https://github.com/ghostnomad/go-pulse-agent)
- [Collector Server Setup](https://github.com/ghostnomad/go-pulse-server)