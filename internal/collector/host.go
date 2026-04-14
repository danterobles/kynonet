package collector

import (
	"context"
	"fmt"

	"github.com/denisbrodbeck/machineid"
	"github.com/shirou/gopsutil/v3/host"
)

func collectHost(ctx context.Context) (HostInfo, error) {
	info, err := host.InfoWithContext(ctx)
	if err != nil {
		return HostInfo{}, fmt.Errorf("host info: %w", err)
	}

	mid, err := machineid.ID()
	if err != nil {
		// Non-fatal: machine ID may be unavailable in containers or restricted envs.
		mid = ""
	}

	return HostInfo{
		Hostname:  info.Hostname,
		OS:        info.OS,
		Platform:  info.Platform,
		Arch:      info.KernelArch,
		MachineID: mid,
	}, nil
}
