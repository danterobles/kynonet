# recolector kynonet

Cross-platform CLI utility written in Go to collect operating system metrics and export them in JSON format.

Compatible with **macOS**, **Linux**, and **Windows** (amd64 / arm64).

## What does it collect?

| Category | Data |
|----------|------|
| **Host** | Hostname, operating system, platform, kernel architecture, machine UUID |
| **Memory** | Total, used, available, usage percentage |
| **CPU** | Model, physical cores, logical threads, frequency (MHz) |
| **Disks** | Per partition: device, mount point, filesystem type, total/used/free, percentage |
| **Processes** | PID, name, RSS memory, CPU percentage |

## Output modes

| Flag | Description |
|------|-------------|
| `-output stdout` | Prints JSON to the console *(default)* |
| `-output file -file <path>` | Writes JSON to a file (overwrites on each cycle) |
| `-output post -endpoint <url>` | Sends JSON via HTTP POST to the specified endpoint |

## Installation

### From source

Requires Go 1.21 or higher.

```bash
git clone https://github.com/danterobles/recolector.git
cd recolector
go build -o recolector ./cmd/recolector
```

### Cross-compilation

```bash
# Linux (x86_64)
GOOS=linux GOARCH=amd64 go build -o recolector-linux ./cmd/recolector

# Windows (x86_64)
GOOS=windows GOARCH=amd64 go build -o recolector.exe ./cmd/recolector

# macOS Apple Silicon
GOOS=darwin GOARCH=arm64 go build -o recolector-darwin ./cmd/recolector
```

## Usage

```bash
# Single collection to console with readable formatting
./recolector -pretty

# Save to file
./recolector -output file -file /tmp/snapshot.json

# Send to a collection server with authentication token
./recolector -output post -endpoint http://server:8080/collect -token my-secret-token

# Continuous collection every 10 seconds (Ctrl-C to stop)
./recolector -interval 10 -output file -file /var/log/metrics.json
```

### All flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-output` | string | `stdout` | Destination: `stdout`, `file`, `post` |
| `-file` | string | — | File path (required with `-output file`) |
| `-endpoint` | string | — | Endpoint URL (required with `-output post`) |
| `-token` | string | — | Bearer token for HTTP authentication |
| `-pretty` | bool | `false` | Indented JSON output (2 spaces) |
| `-interval` | int | `0` | Seconds between collections; `0` = run once |

## Output example

```json
{
  "collectedAt": "2026-04-14T22:00:00Z",
  "host": {
    "hostname": "my-server",
    "os": "linux",
    "platform": "alma",
    "arch": "x86_64",
    "machineId": "a1b2c3d4-e5f6-..."
  },
  "memory": {
    "totalBytes": 34359738368,
    "usedBytes": 16357154816,
    "availableBytes": 18002583552,
    "usedPercent": 47.6
  },
  "cpu": {
    "modelName": "Intel Xeon E5-2680 v4",
    "physicalCores": 14,
    "logicalCores": 28,
    "frequencyMhz": 2400
  },
  "disks": [
    {
      "device": "/dev/sda1",
      "mountpoint": "/",
      "fstype": "xfs",
      "totalBytes": 107374182400,
      "usedBytes": 52010680320,
      "freeBytes": 55363502080,
      "usedPercent": 48.47
    }
  ],
  "processes": [
    {
      "pid": 1,
      "name": "systemd",
      "memoryRssBytes": 12288000,
      "cpuPercent": 0.01
    }
  ]
}
```

## Architecture

```
recolector/
├── cmd/recolector/        # CLI entry point, wiring, and interval loop
├── internal/
│   ├── config/            # Flag parsing and validation (FromFlags)
│   ├── collector/         # OS metrics collection (5 sub-collectors)
│   └── exporter/          # Exporters: stdout, file, HTTP POST
├── go.mod
└── go.sum
```

The two core interfaces are:

```go
type Collector interface {
    Collect(ctx context.Context) (*SystemInfo, error)
}

type Exporter interface {
    Export(ctx context.Context, info *collector.SystemInfo) error
}
```

`main.go` wires everything together: parses flags → builds the Exporter → instantiates the Collector → calls `run()`. The `run()` function accepts interfaces, making it testable without real OS calls.

## Dependencies

| Package | Version | Purpose |
|---------|---------|---------|
| `github.com/shirou/gopsutil/v3` | v3.24.5 | Cross-platform CPU, memory, disk, process, and host metrics |
| `github.com/denisbrodbeck/machineid` | v1.0.1 | Stable hardware UUID for the machine |

## Development

```bash
# Run in development mode (recommended on macOS)
go run ./cmd/recolector -pretty

# Tests with race condition detector
go test -race ./...

# Static analysis
go vet ./...

# Clean up dependencies
go mod tidy
```

> **macOS note:** The compiled binary requires code signing to enumerate system
> processes. Use `go run` during development. In production, sign the binary
> with the `com.apple.security.get-task-allow` entitlement.

## License

MIT
