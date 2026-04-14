package collector

import (
	"context"
	"fmt"
	"math"

	"github.com/shirou/gopsutil/v3/mem"
)

func collectMemory(ctx context.Context) (MemoryInfo, error) {
	vm, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return MemoryInfo{}, fmt.Errorf("virtual memory: %w", err)
	}

	return MemoryInfo{
		TotalBytes:     vm.Total,
		UsedBytes:      vm.Used,
		AvailableBytes: vm.Available,
		UsedPercent:    round2(vm.UsedPercent),
	}, nil
}

// round2 rounds a float64 to 2 decimal places.
func round2(v float64) float64 {
	return math.Round(v*100) / 100
}
