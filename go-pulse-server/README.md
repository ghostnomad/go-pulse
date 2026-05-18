# Go-Pulse Collector Server

The central collector and visualization dashboard for the Go-Pulse monitoring ecosystem. It exposes a lightweight REST API to ingest telemetry updates from remote agents, saves historical reports into an embedded SQLite database, and hosts a minimal web UI dashboard.

## Installation & Setup

### 1. Prerequisites

Ensure you have Go installed on your main server (version 1.22+ recommended).

### 2. Clone and Build the Server

Clone the repository and compile the binary:

```bash
git clone https://github.com/yourusername/go-pulse-server.git
cd go-pulse-server
go mod tidy
go build -o go-pulse-server main.go
```

### 3. Running the Server

By default, the server listens on port `8080` and initializes an embedded database file named `pulse.db` automatically in its local directory.

Run the compiled binary:

```bash
PORT=8080 ./go-pulse-server
```

### 4. View the Dashboard

Open your web browser and navigate to the server's IP address:

```
http://<YOUR_SERVER_IP>:8080
```

> If running on the same machine as the server, you can use `http://localhost:8080`