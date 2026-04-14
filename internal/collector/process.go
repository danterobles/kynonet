package collector

import (
	"context"
	"fmt"

	"github.com/shirou/gopsutil/v3/process"
)

// collectProcesses obtiene la lista completa de procesos en ejecución y
// recopila los datos de cada uno usando una estrategia de errores "soft":
// los procesos individuales que fallen no abortan la recolección global.
//
// Es normal que algunos procesos desaparezcan entre el momento en que se
// obtiene la lista y el momento en que se leen sus atributos (race condition
// inherente a /proc en Linux). Estos procesos efímeros son simplemente omitidos.
func collectProcesses(ctx context.Context) ([]ProcessInfo, error) {
	procs, err := process.ProcessesWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("process list: %w", err)
	}

	// Pre-alocar con la capacidad de la lista original como estimación superior.
	result := make([]ProcessInfo, 0, len(procs))

	for _, p := range procs {
		info, ok := gatherProcess(ctx, p)
		if !ok {
			continue // proceso efímero o sin permisos; se omite silenciosamente
		}
		result = append(result, info)
	}

	return result, nil
}

// gatherProcess recopila los atributos de un proceso individual.
// Devuelve (info, false) si el proceso ya no existe o si no se puede leer su nombre,
// lo que indica que debe ser omitido de los resultados.
//
// Los campos MemoryRSS y CPUPercent son opcionales: si no se pueden leer
// (por permisos o porque el proceso terminó entre llamadas), se dejan en cero
// en lugar de descartar el proceso completo.
//
// CPUPercent devuelve el porcentaje de uso desde la última invocación para
// ese proceso. En la primera llamada siempre devuelve 0.0, lo cual es esperado.
func gatherProcess(ctx context.Context, p *process.Process) (ProcessInfo, bool) {
	name, err := p.NameWithContext(ctx)
	if err != nil {
		// El proceso probablemente ya terminó. Se omite del resultado.
		return ProcessInfo{}, false
	}

	info := ProcessInfo{
		PID:  p.Pid,
		Name: name,
	}

	// Memoria RSS: memoria física real ocupada por el proceso.
	if memInfo, err := p.MemoryInfoWithContext(ctx); err == nil {
		info.MemoryRSS = memInfo.RSS
	}

	// Porcentaje de CPU: se usa el valor acumulado desde la última medición.
	if cpuPct, err := p.CPUPercentWithContext(ctx); err == nil {
		info.CPUPercent = round2(cpuPct)
	}

	return info, true
}
