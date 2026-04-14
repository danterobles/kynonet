package collector

import (
	"context"
	"fmt"
	"math"

	"github.com/shirou/gopsutil/v3/mem"
)

// collectMemory consulta las estadísticas de memoria virtual del sistema operativo
// y las mapea al tipo MemoryInfo. Los valores reflejan el estado de la RAM
// en el momento exacto de la llamada.
//
// Utiliza mem.VirtualMemoryWithContext para respetar cancelaciones de contexto,
// lo que es importante en el modo de ejecución con intervalo (-interval).
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

// round2 redondea un valor float64 a exactamente 2 decimales.
// Se aplica a todos los campos de porcentaje para evitar que el JSON
// contenga secuencias largas de decimales innecesarios (ej. 47.599999999...).
func round2(v float64) float64 {
	return math.Round(v*100) / 100
}
