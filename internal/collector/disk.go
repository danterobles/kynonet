package collector

import (
	"context"
	"fmt"

	"github.com/shirou/gopsutil/v3/disk"
)

func collectDisks(ctx context.Context) ([]DiskInfo, error) {
	partitions, err := disk.PartitionsWithContext(ctx, false)
	if err != nil {
		return nil, fmt.Errorf("disk partitions: %w", err)
	}

	result := make([]DiskInfo, 0, len(partitions))

	for _, p := range partitions {
		usage, err := disk.UsageWithContext(ctx, p.Mountpoint)
		if err != nil {
			// Skip partitions that can't be read (optical drives, unmounted volumes, etc.)
			continue
		}

		result = append(result, DiskInfo{
			Device:      p.Device,
			Mountpoint:  p.Mountpoint,
			Fstype:      p.Fstype,
			TotalBytes:  usage.Total,
			UsedBytes:   usage.Used,
			FreeBytes:   usage.Free,
			UsedPercent: round2(usage.UsedPercent),
		})
	}

	return result, nil
}
