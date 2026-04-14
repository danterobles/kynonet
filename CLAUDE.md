# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

`recolector` is a cross-platform CLI utility (macOS, Windows, Linux) that collects system metrics — CPU, memory, disk, processes, and machine identity — and exports them as JSON to stdout, a file, or via HTTP POST.

## Commands

```bash
# Run (development)
go run ./cmd/recolector -pretty
go run ./cmd/recolector -output file -file /tmp/snap.json
go run ./cmd/recolector -output post -endpoint http://localhost:8080/collect -token mytoken
go run ./cmd/recolector -interval 5 -pretty          # repeat every 5s, Ctrl-C to stop

# Build
go build -o recolector ./cmd/recolector

# Test all packages
go test ./...
go test -race ./...

# Single test by name
go test -run TestFunctionName ./internal/collector/...

# Lint and vet
go vet ./...

# Cross-platform compilation check
GOOS=linux   GOARCH=amd64 go build ./...
GOOS=windows GOARCH=amd64 go build ./...
GOOS=darwin  GOARCH=arm64 go build ./...

# Tidy modules
go mod tidy
```

> **macOS note**: The compiled binary requires code-signing to enumerate process info from the OS. Use `go run` during development. In production, sign the binary with the `com.apple.security.get-task-allow` entitlement.

## Architecture

Two interfaces are the core seam of the system:

```go
// internal/collector/collector.go
type Collector interface {
    Collect(ctx context.Context) (*SystemInfo, error)
}

// internal/exporter/exporter.go
type Exporter interface {
    Export(ctx context.Context, info *collector.SystemInfo) error
}
```

`cmd/recolector/main.go` wires everything: parses flags → builds the right `Exporter` via `buildExporter()` → creates `collector.New()` → calls `run()`. The `run()` function takes interfaces, making it independently testable without OS calls.

### Packages

| Package | Responsibility |
|---------|---------------|
| `internal/config` | Flag parsing and validation (`FromFlags`) |
| `internal/collector` | `SystemCollector` aggregates 5 sub-collectors: `collectHost`, `collectMemory`, `collectCPU`, `collectDisks`, `collectProcesses` |
| `internal/exporter` | `StdoutExporter`, `FileExporter`, `HTTPExporter` — each takes an injectable `io.Writer` or `*http.Client` for testing |

### Key design rules
- Sub-collector functions (`collectHost`, etc.) are package-private; `SystemCollector` is the only exported type in `collector`
- Per-process errors in `collectProcesses` are soft — a process that disappears mid-collection is skipped, not fatal
- Disk partitions that fail `Usage()` (optical drives, unmounted volumes) are silently skipped
- All percentage fields are rounded to 2 decimal places via `round2()`
- HTTP exporter always drains the response body before closing to allow TCP connection reuse

## Dependencies

- `github.com/shirou/gopsutil/v3` — cross-platform system metrics (CPU, memory, disk, host, processes)
- `github.com/denisbrodbeck/machineid` — stable machine UUID (reads `/etc/machine-id` on Linux, registry on Windows, `ioreg` on macOS)
