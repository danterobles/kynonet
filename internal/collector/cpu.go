package collector

import (
	"context"
	"fmt"

	"github.com/shirou/gopsutil/v3/cpu"
)

func collectCPU(ctx context.Context) (CPUInfo, error) {
	phys, err := cpu.CountsWithContext(ctx, false)
	if err != nil {
		return CPUInfo{}, fmt.Errorf("cpu physical count: %w", err)
	}

	logical, err := cpu.CountsWithContext(ctx, true)
	if err != nil {
		return CPUInfo{}, fmt.Errorf("cpu logical count: %w", err)
	}

	infos, err := cpu.InfoWithContext(ctx)
	if err != nil {
		return CPUInfo{}, fmt.Errorf("cpu info: %w", err)
	}

	info := CPUInfo{
		PhysicalCores: phys,
		LogicalCores:  logical,
	}

	// cpu.Info() may return an empty slice on some VMs/containers.
	if len(infos) > 0 {
		info.ModelName = infos[0].ModelName
		info.FrequencyMHz = round2(infos[0].Mhz)
	}

	return info, nil
}
