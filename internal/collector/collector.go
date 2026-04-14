// Package collector recopila métricas del sistema operativo y las agrupa
// en una instantánea (snapshot) estructurada lista para ser serializada.
//
// La interfaz Collector es el único contrato público del paquete; SystemCollector
// es su implementación de producción. Los cinco sub-colectores (host, memory, cpu,
// disk, process) son funciones privadas del paquete y no forman parte de ninguna
// interfaz, lo que mantiene la superficie de API mínima.
package collector

import (
	"context"
	"fmt"
	"time"
)

// Collector define el contrato para cualquier implementación que recopile
// información del sistema. El contexto permite propagar cancelaciones y
// timeouts a todas las llamadas al sistema operativo subyacentes.
type Collector interface {
	Collect(ctx context.Context) (*SystemInfo, error)
}

// SystemInfo es la estructura raíz que contiene la instantánea completa
// del sistema en un momento dado. Es directamente serializable a JSON.
type SystemInfo struct {
	// CollectedAt es la marca de tiempo UTC en que se completó la recolección.
	CollectedAt time.Time `json:"collectedAt"`
	// Host contiene los datos de identidad del equipo.
	Host HostInfo `json:"host"`
	// Memory contiene el estado actual de la memoria virtual.
	Memory MemoryInfo `json:"memory"`
	// CPU contiene información estática del procesador.
	CPU CPUInfo `json:"cpu"`
	// Disks contiene una entrada por cada partición de disco accesible.
	Disks []DiskInfo `json:"disks"`
	// Processes contiene una entrada por cada proceso en ejecución al momento de la recolección.
	Processes []ProcessInfo `json:"processes"`
}

// HostInfo agrupa los campos que identifican de forma única al equipo.
type HostInfo struct {
	// Hostname es el nombre de red del equipo (equivalente a `hostname` en la terminal).
	Hostname string `json:"hostname"`
	// OS es el nombre del sistema operativo (ej. "darwin", "linux", "windows").
	OS string `json:"os"`
	// Platform es la distribución o variante del OS (ej. "ubuntu", "alma").
	Platform string `json:"platform"`
	// Arch es la arquitectura del kernel (ej. "x86_64", "arm64").
	Arch string `json:"arch"`
	// MachineID es un UUID estable generado a partir de hardware/SO,
	// útil para identificar el equipo entre reinicios. Puede estar vacío
	// en contenedores o entornos con permisos restringidos.
	MachineID string `json:"machineId"`
}

// MemoryInfo contiene estadísticas de la memoria RAM virtual del sistema.
// Todos los valores en bytes; el porcentaje está redondeado a 2 decimales.
type MemoryInfo struct {
	// TotalBytes es la capacidad total de RAM instalada.
	TotalBytes uint64 `json:"totalBytes"`
	// UsedBytes es la memoria actualmente en uso por el sistema y los procesos.
	UsedBytes uint64 `json:"usedBytes"`
	// AvailableBytes es la memoria disponible para nuevas asignaciones sin necesidad de swap.
	AvailableBytes uint64 `json:"availableBytes"`
	// UsedPercent es el porcentaje de uso respecto al total (redondeado a 2 decimales).
	UsedPercent float64 `json:"usedPercent"`
}

// CPUInfo contiene información descriptiva y de capacidad del procesador.
// En sistemas con múltiples sockets físicos, los valores corresponden al primer socket detectado.
type CPUInfo struct {
	// ModelName es el nombre del modelo del procesador (ej. "Apple M2 Max", "Intel Core i7-12700K").
	ModelName string `json:"modelName"`
	// PhysicalCores es el número de núcleos físicos del procesador.
	PhysicalCores int `json:"physicalCores"`
	// LogicalCores es el número de hilos lógicos, incluyendo hyper-threading si está habilitado.
	LogicalCores int `json:"logicalCores"`
	// FrequencyMHz es la frecuencia base del procesador en megahertz (redondeada a 2 decimales).
	FrequencyMHz float64 `json:"frequencyMhz"`
}

// DiskInfo contiene el estado de uso de una partición de disco individual.
// Las particiones que no son accesibles (unidades ópticas, volúmenes no montados) son omitidas.
type DiskInfo struct {
	// Device es el identificador del dispositivo (ej. "/dev/sda1", "C:").
	Device string `json:"device"`
	// Mountpoint es el punto de montaje en el sistema de archivos (ej. "/", "/home", "C:\").
	Mountpoint string `json:"mountpoint"`
	// Fstype es el tipo de sistema de archivos (ej. "ext4", "apfs", "ntfs").
	Fstype string `json:"fstype"`
	// TotalBytes es la capacidad total de la partición en bytes.
	TotalBytes uint64 `json:"totalBytes"`
	// UsedBytes es el espacio actualmente ocupado en bytes.
	UsedBytes uint64 `json:"usedBytes"`
	// FreeBytes es el espacio disponible para escritura en bytes.
	FreeBytes uint64 `json:"freeBytes"`
	// UsedPercent es el porcentaje de uso del disco (redondeado a 2 decimales).
	UsedPercent float64 `json:"usedPercent"`
}

// ProcessInfo contiene una instantánea de los recursos consumidos por un proceso.
// Los campos de memoria y CPU pueden ser cero si el proceso no concedió acceso
// antes de terminar durante la recolección.
type ProcessInfo struct {
	// PID es el identificador de proceso asignado por el sistema operativo.
	PID int32 `json:"pid"`
	// Name es el nombre del ejecutable del proceso.
	Name string `json:"name"`
	// MemoryRSS es el Resident Set Size: la memoria física real que ocupa el proceso en bytes.
	MemoryRSS uint64 `json:"memoryRssBytes"`
	// CPUPercent es el porcentaje de CPU utilizado desde la última medición (redondeado a 2 decimales).
	CPUPercent float64 `json:"cpuPercent"`
}

// SystemCollector es la implementación de producción de la interfaz Collector.
// Orquesta los cinco sub-colectores privados y combina sus resultados en un SystemInfo.
type SystemCollector struct{}

// New devuelve un SystemCollector listo para usar.
// No requiere ningún parámetro de configuración; las dependencias del sistema
// operativo son resueltas en tiempo de ejecución por cada sub-colector.
func New() *SystemCollector {
	return &SystemCollector{}
}

// Collect ejecuta los cinco sub-colectores en secuencia y devuelve un SystemInfo
// completamente poblado con la marca de tiempo UTC del momento de la recolección.
// Si cualquier sub-colector falla con un error irrecuperable, Collect retorna
// inmediatamente con ese error envuelto en contexto.
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
