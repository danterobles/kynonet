package collector

import (
	"context"
	"fmt"

	"github.com/shirou/gopsutil/v3/process"
)

func collectProcesses(ctx context.Context) ([]ProcessInfo, error) {
	procs, err := process.ProcessesWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("process list: %w", err)
	}

	result := make([]ProcessInfo, 0, len(procs))

	for _, p := range procs {
		info, ok := gatherProcess(ctx, p)
		if !ok {
			continue
		}
		result = append(result, info)
	}

	return result, nil
}

// gatherProcess collects data for a single process.
// Returns (info, false) if the process has disappeared or Name cannot be read.
func gatherProcess(ctx context.Context, p *process.Process) (ProcessInfo, bool) {
	name, err := p.NameWithContext(ctx)
	if err != nil {
		// Process likely exited; skip it.
		return ProcessInfo{}, false
	}

	info := ProcessInfo{
		PID:  p.Pid,
		Name: name,
	}

	if memInfo, err := p.MemoryInfoWithContext(ctx); err == nil {
		info.MemoryRSS = memInfo.RSS
	}

	if cpuPct, err := p.CPUPercentWithContext(ctx); err == nil {
		info.CPUPercent = round2(cpuPct)
	}

	return info, true
}
