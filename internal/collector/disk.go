package collector

import (
	"context"
	"fmt"

	"github.com/shirou/gopsutil/v3/disk"
)

// collectDisks enumera las particiones físicas del sistema y recopila
// las estadísticas de uso de cada una.
//
// Solo se incluyen particiones físicas (all=false en PartitionsWithContext),
// lo que excluye sistemas de archivos virtuales como tmpfs, devfs o procfs en Linux.
//
// Las particiones que no se pueden leer (unidades ópticas sin medio, volúmenes
// desmontados, puntos de montaje con permisos insuficientes) son silenciosamente
// omitidas para no abortar la recolección completa. El slice resultante puede
// estar vacío si ninguna partición es accesible.
//
// La representación varía por plataforma:
//   - Linux/macOS: Mountpoint es una ruta tipo "/", "/home", "/Volumes/Data"
//   - Windows: Mountpoint es una letra de unidad tipo "C:\", "D:\"
func collectDisks(ctx context.Context) ([]DiskInfo, error) {
	partitions, err := disk.PartitionsWithContext(ctx, false)
	if err != nil {
		return nil, fmt.Errorf("disk partitions: %w", err)
	}

	// Pre-alocar con capacidad conocida para evitar reasignaciones del slice.
	result := make([]DiskInfo, 0, len(partitions))

	for _, p := range partitions {
		usage, err := disk.UsageWithContext(ctx, p.Mountpoint)
		if err != nil {
			// Omitir particiones inaccesibles (error "soft"); continuar con las demás.
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
