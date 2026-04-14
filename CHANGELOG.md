# Changelog

Todos los cambios notables de este proyecto están documentados en este archivo.

El formato sigue [Keep a Changelog](https://keepachangelog.com/es/1.0.0/)
y el proyecto adhiere a [Semantic Versioning](https://semver.org/lang/es/).

---

## [0.1.0] - 2026-04-14

### Versión inicial — implementación completa del núcleo

Primera versión funcional de `recolector`. Incluye la arquitectura completa de recolección y exportación de métricas del sistema operativo.

---

### Nuevos paquetes

#### `cmd/recolector`
- Punto de entrada de la herramienta CLI.
- Parseo de flags vía `config.FromFlags`.
- Función `buildExporter`: fábrica que selecciona el exportador según `-output`.
- Función `run`: loop de recolección con soporte para ejecución única e intervalo configurable. Acepta interfaces para ser testeable sin acceso al sistema operativo.
- Graceful shutdown mediante `signal.NotifyContext` con `SIGINT` y `SIGTERM`.

#### `internal/config`
- Struct `Config` con todos los parámetros de ejecución: `Output`, `FilePath`, `Endpoint`, `Token`, `Pretty`, `Interval`.
- Función `FromFlags(args []string) (Config, error)`: parseo con `flag.FlagSet` local (testeable en aislamiento).
- Validación de combinaciones de flags: `file` requiere `-file`; `post` requiere `-endpoint` con URL válida.
- Conversión de segundos enteros a `time.Duration` para el campo `Interval`.

#### `internal/collector`
- Interfaz `Collector` con método `Collect(ctx) (*SystemInfo, error)`.
- Tipo `SystemInfo`: estructura raíz JSON con campos `CollectedAt`, `Host`, `Memory`, `CPU`, `Disks`, `Processes`.
- Tipos de datos: `HostInfo`, `MemoryInfo`, `CPUInfo`, `DiskInfo`, `ProcessInfo`.
- Struct `SystemCollector` (única exportación del paquete): orquesta los cinco sub-colectores privados.
- Función `New() *SystemCollector`: constructor sin parámetros.
- `collectHost`: recopila hostname, OS, plataforma, arquitectura del kernel y machine UUID.
- `collectMemory`: lee total, usado, disponible y porcentaje de la memoria virtual.
- `collectCPU`: obtiene núcleos físicos, hilos lógicos, modelo y frecuencia base.
- `collectDisks`: enumera particiones físicas y sus estadísticas de uso; omite particiones inaccesibles.
- `collectProcesses`: lista todos los procesos en ejecución con PID, nombre, RSS y porcentaje de CPU; manejo "soft" de errores por proceso.
- Función utilitaria privada `round2`: redondea porcentajes a 2 decimales.

#### `internal/exporter`
- Interfaz `Exporter` con método `Export(ctx, *SystemInfo) error`.
- `StdoutExporter`: escribe JSON en un `io.Writer` inyectable (por defecto `os.Stdout`). Opción `WithWriter` para pruebas.
- `FileExporter`: escribe JSON en un archivo del sistema con permisos `0644`; sobreescribe en cada ciclo.
- `HTTPExporter`: envía JSON vía HTTP POST. Soporta token Bearer, timeout configurable (`WithHTTPTimeout`) y cliente inyectable (`WithHTTPClient`). Drena el cuerpo de la respuesta para reutilización de conexiones TCP. Retorna error en respuestas HTTP >= 400.

---

### Dependencias externas añadidas

| Paquete | Versión | Propósito |
|---------|---------|-----------|
| `github.com/shirou/gopsutil/v3` | v3.24.5 | Métricas cross-platform de CPU, memoria, disco, procesos y host |
| `github.com/denisbrodbeck/machineid` | v1.0.1 | UUID estable del hardware del equipo |

### Plataformas verificadas

| Sistema | Arquitectura | Estado |
|---------|-------------|--------|
| macOS (Darwin) | arm64 | ✓ Compilación y ejecución verificadas |
| Linux | amd64 | ✓ Compilación cruzada verificada |
| Windows | amd64 | ✓ Compilación cruzada verificada |

### Tests incluidos

| Paquete | Cobertura |
|---------|-----------|
| `internal/config` | 9 casos tabla-driven: defaults, modos válidos, errores de validación, conversión de intervalo |
| `internal/collector` | Test de integración con OS real: verifica hostname, cores, memoria total, presencia de discos |
| `internal/exporter` | 7 tests unitarios: stdout compacto/pretty, escritura y sobreescritura de archivo, HTTP con token, HTTP sin token, HTTP error 500 |
| `cmd/recolector` | 3 tests con mocks: ejecución única, loop con cancelación de contexto, propagación de errores del colector |

---

## [Sin publicar]

### Mejoras planificadas
- Firma de código para macOS en modo binario compilado.
- Flag `-collect-timeout` para establecer un timeout máximo por ciclo de recolección.
- Salida en formato CSV además de JSON.
- Filtro de procesos por nombre o PID mínimo.
- Métricas de red (interfaces, bytes enviados/recibidos).
