# Changelog

All notable changes to this project are documented in this file.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](https://semver.org/).

---

## [0.1.0] - 2026-04-14

### Initial release — complete core implementation

First functional version of `recolector`. Includes the full collection and export architecture for operating system metrics.

---

### New packages

#### `cmd/recolector`
- CLI tool entry point.
- Flag parsing via `config.FromFlags`.
- `buildExporter` function: factory that selects the exporter based on `-output`.
- `run` function: collection loop with support for single execution and configurable interval. Accepts interfaces to be testable without OS access.
- Graceful shutdown via `signal.NotifyContext` with `SIGINT` and `SIGTERM`.

#### `internal/config`
- `Config` struct with all runtime parameters: `Output`, `FilePath`, `Endpoint`, `Token`, `Pretty`, `Interval`.
- `FromFlags(args []string) (Config, error)` function: parsing with a local `flag.FlagSet` (independently testable).
- Flag combination validation: `file` requires `-file`; `post` requires `-endpoint` with a valid URL.
- Integer-to-`time.Duration` conversion for the `Interval` field.

#### `internal/collector`
- `Collector` interface with `Collect(ctx) (*SystemInfo, error)` method.
- `SystemInfo` type: root JSON structure with fields `CollectedAt`, `Host`, `Memory`, `CPU`, `Disks`, `Processes`.
- Data types: `HostInfo`, `MemoryInfo`, `CPUInfo`, `DiskInfo`, `ProcessInfo`.
- `SystemCollector` struct (the only exported type from the package): orchestrates the five private sub-collectors.
- `New() *SystemCollector` constructor: no parameters required.
- `collectHost`: gathers hostname, OS, platform, kernel architecture, and machine UUID.
- `collectMemory`: reads total, used, available, and percentage of virtual memory.
- `collectCPU`: obtains physical cores, logical threads, model name, and base frequency.
- `collectDisks`: enumerates physical partitions and their usage stats; skips inaccessible partitions.
- `collectProcesses`: lists all running processes with PID, name, RSS memory, and CPU percentage; soft error handling per process.
- Private utility function `round2`: rounds percentage values to 2 decimal places.

#### `internal/exporter`
- `Exporter` interface with `Export(ctx, *SystemInfo) error` method.
- `StdoutExporter`: writes JSON to an injectable `io.Writer` (default `os.Stdout`). `WithWriter` option for testing.
- `FileExporter`: writes JSON to a file with `0644` permissions; overwrites on each cycle.
- `HTTPExporter`: sends JSON via HTTP POST. Supports Bearer token, configurable timeout (`WithHTTPTimeout`), and injectable client (`WithHTTPClient`). Drains the response body for TCP connection reuse. Returns an error on HTTP responses >= 400.

---

### External dependencies added

| Package | Version | Purpose |
|---------|---------|---------|
| `github.com/shirou/gopsutil/v3` | v3.24.5 | Cross-platform CPU, memory, disk, process, and host metrics |
| `github.com/denisbrodbeck/machineid` | v1.0.1 | Stable hardware UUID for the machine |

### Verified platforms

| OS | Architecture | Status |
|----|-------------|--------|
| macOS (Darwin) | arm64 | ✓ Build and execution verified |
| Linux | amd64 | ✓ Cross-compilation verified |
| Windows | amd64 | ✓ Cross-compilation verified |

### Tests included

| Package | Coverage |
|---------|----------|
| `internal/config` | 9 table-driven cases: defaults, valid modes, validation errors, interval conversion |
| `internal/collector` | Integration test with real OS: verifies hostname, cores, total memory, presence of disks |
| `internal/exporter` | 7 unit tests: compact/pretty stdout, file write and overwrite, HTTP with token, HTTP without token, HTTP 500 error |
| `cmd/recolector` | 3 tests with mocks: single execution, loop with context cancellation, collector error propagation |

---

## [Unreleased]

### Planned improvements
- Code signing support for macOS compiled binary.
- `-collect-timeout` flag to set a maximum timeout per collection cycle.
- CSV output format in addition to JSON.
- Process filtering by name or minimum PID.
- Network metrics (interfaces, bytes sent/received).
