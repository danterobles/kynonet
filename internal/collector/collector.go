package collector

import (
	"context"
	"fmt"
	"time"
)

// Collector is the interface for gathering system information.
type Collector interface {
	Collect(ctx context.Context) (*SystemInfo, error)
}

// SystemInfo is the top-level JSON-serialisable snapshot of the system.
type SystemInfo struct {
	CollectedAt time.Time     `json:"collectedAt"`
	Host        HostInfo      `json:"host"`
	Memory      MemoryInfo    `json:"memory"`
	CPU         CPUInfo       `json:"cpu"`
	Disks       []DiskInfo    `json:"disks"`
	Processes   []ProcessInfo `json:"processes"`
}

// HostInfo holds machine identity fields.
type HostInfo struct {
	Hostname  string `json:"hostname"`
	OS        string `json:"os"`
	Platform  string `json:"platform"`
	Arch      string `json:"arch"`
	MachineID string `json:"machineId"`
}

// MemoryInfo holds virtual memory stats.
type MemoryInfo struct {
	TotalBytes     uint64  `json:"totalBytes"`
	UsedBytes      uint64  `json:"usedBytes"`
	AvailableBytes uint64  `json:"availableBytes"`
	UsedPercent    float64 `json:"usedPercent"`
}

// CPUInfo holds processor information.
type CPUInfo struct {
	ModelName     string  `json:"modelName"`
	PhysicalCores int     `json:"physicalCores"`
	LogicalCores  int     `json:"logicalCores"`
	FrequencyMHz  float64 `json:"frequencyMhz"`
}

// DiskInfo holds per-partition disk usage.
type DiskInfo struct {
	Device      string  `json:"device"`
	Mountpoint  string  `json:"mountpoint"`
	Fstype      string  `json:"fstype"`
	TotalBytes  uint64  `json:"totalBytes"`
	UsedBytes   uint64  `json:"usedBytes"`
	FreeBytes   uint64  `json:"freeBytes"`
	UsedPercent float64 `json:"usedPercent"`
}

// ProcessInfo holds per-process snapshot data.
type ProcessInfo struct {
	PID        int32   `json:"pid"`
	Name       string  `json:"name"`
	MemoryRSS  uint64  `json:"memoryRssBytes"`
	CPUPercent float64 `json:"cpuPercent"`
}

// SystemCollector is the production implementation of Collector.
type SystemCollector struct{}

// New returns a ready-to-use SystemCollector.
func New() *SystemCollector {
	return &SystemCollector{}
}

// Collect gathers all system metrics and returns a SystemInfo snapshot.
func (c *SystemCollector) Collect(ctx context.Context) (*SystemInfo, error) {
	host, err := collectHost(ctx)
	if err != nil {
		return nil, fmt.Errorf("collect host: %w", err)
	}

	mem, err := collectMemory(ctx)
	if err != nil {
		return nil, fmt.Errorf("collect memory: %w", err)
	}

	cpu, err := collectCPU(ctx)
	if err != nil {
		return nil, fmt.Errorf("collect cpu: %w", err)
	}

	disks, err := collectDisks(ctx)
	if err != nil {
		return nil, fmt.Errorf("collect disks: %w", err)
	}

	procs, err := collectProcesses(ctx)
	if err != nil {
		return nil, fmt.Errorf("collect processes: %w", err)
	}

	return &SystemInfo{
		CollectedAt: time.Now().UTC(),
		Host:        host,
		Memory:      mem,
		CPU:         cpu,
		Disks:       disks,
		Processes:   procs,
	}, nil
}
