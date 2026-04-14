package collector

import (
	"context"
	"fmt"

	"github.com/shirou/gopsutil/v3/cpu"
)

// collectCPU recopila información descriptiva y de capacidad del procesador.
// Realiza tres llamadas independientes a gopsutil para obtener el conteo
// de núcleos físicos, el conteo de hilos lógicos y los detalles del modelo.
//
// Comportamiento en casos especiales:
//   - VMs y contenedores: cpu.Info() puede devolver un slice vacío en algunos
//     hipervisores; en ese caso ModelName y FrequencyMHz quedan en sus valores cero.
//   - Sistemas multi-socket: cpu.Info() devuelve una entrada por socket en macOS
//     y una por hilo lógico en Linux. En ambos casos se usa infos[0] para el
//     modelo y la frecuencia, que son iguales para todos los núcleos del mismo chip.
func collectCPU(ctx context.Context) (CPUInfo, error) {
	// Núcleos físicos (sin hyper-threading).
	phys, err := cpu.CountsWithContext(ctx, false)
	if err != nil {
		return CPUInfo{}, fmt.Errorf("cpu physical count: %w", err)
	}

	// Hilos lógicos (incluye hyper-threading si está habilitado).
	logical, err := cpu.CountsWithContext(ctx, true)
	if err != nil {
		return CPUInfo{}, fmt.Errorf("cpu logical count: %w", err)
	}

	// Información detallada del modelo: nombre y frecuencia base.
	infos, err := cpu.InfoWithContext(ctx)
	if err != nil {
		return CPUInfo{}, fmt.Errorf("cpu info: %w", err)
	}

	info := CPUInfo{
		PhysicalCores: phys,
		LogicalCores:  logical,
	}

	// Guardar contra slice vacío en VMs o contenedores ligeros.
	if len(infos) > 0 {
		info.ModelName = infos[0].ModelName
		info.FrequencyMHz = round2(infos[0].Mhz)
	}

	return info, nil
}
