# recolector kynonet 

Utilería CLI multiplataforma escrita en Go para recolectar métricas del sistema operativo y exportarlas en formato JSON.

Compatible con **macOS**, **Linux** y **Windows** (amd64 / arm64).

## ¿Qué recolecta?

| Categoría | Datos |
|-----------|-------|
| **Host** | Hostname, sistema operativo, plataforma, arquitectura del kernel, UUID del equipo |
| **Memoria** | Total, en uso, disponible, porcentaje de uso |
| **CPU** | Modelo, núcleos físicos, hilos lógicos, frecuencia (MHz) |
| **Discos** | Por partición: dispositivo, punto de montaje, sistema de archivos, total/usado/libre, porcentaje |
| **Procesos** | PID, nombre, memoria RSS, porcentaje de CPU |

## Modos de salida

| Flag | Descripción |
|------|-------------|
| `-output stdout` | Imprime el JSON en la consola *(predeterminado)* |
| `-output file -file <ruta>` | Escribe el JSON en un archivo (sobreescribe en cada ciclo) |
| `-output post -endpoint <url>` | Envía el JSON vía HTTP POST al endpoint indicado |

## Instalación

### Desde el código fuente

Requiere Go 1.21 o superior.

```bash
git clone https://github.com/danterobles/recolector.git
cd recolector
go build -o recolector ./cmd/recolector
```

### Compilación cruzada

```bash
# Linux (x86_64)
GOOS=linux GOARCH=amd64 go build -o recolector-linux ./cmd/recolector

# Windows (x86_64)
GOOS=windows GOARCH=amd64 go build -o recolector.exe ./cmd/recolector

# macOS Apple Silicon
GOOS=darwin GOARCH=arm64 go build -o recolector-darwin ./cmd/recolector
```

## Uso

```bash
# Una sola recolección a consola con formato legible
./recolector -pretty

# Guardar en archivo
./recolector -output file -file /tmp/snapshot.json

# Enviar a un servidor de recolección con token de autenticación
./recolector -output post -endpoint http://servidor:8080/collect -token mi-token-secreto

# Recolección continua cada 10 segundos (Ctrl-C para detener)
./recolector -interval 10 -output file -file /var/log/metricas.json
```

### Flags completos

| Flag | Tipo | Predeterminado | Descripción |
|------|------|---------------|-------------|
| `-output` | string | `stdout` | Destino: `stdout`, `file`, `post` |
| `-file` | string | — | Ruta del archivo (requerido con `-output file`) |
| `-endpoint` | string | — | URL del endpoint (requerido con `-output post`) |
| `-token` | string | — | Bearer token para autenticación HTTP |
| `-pretty` | bool | `false` | Formato JSON indentado (2 espacios) |
| `-interval` | int | `0` | Segundos entre recolecciones; `0` = una sola vez |

## Ejemplo de salida

```json
{
  "collectedAt": "2026-04-14T22:00:00Z",
  "host": {
    "hostname": "mi-servidor",
    "os": "linux",
    "platform": "alma",
    "arch": "x86_64",
    "machineId": "a1b2c3d4-e5f6-..."
  },
  "memory": {
    "totalBytes": 34359738368,
    "usedBytes": 16357154816,
    "availableBytes": 18002583552,
    "usedPercent": 47.6
  },
  "cpu": {
    "modelName": "Intel Xeon E5-2680 v4",
    "physicalCores": 14,
    "logicalCores": 28,
    "frequencyMhz": 2400
  },
  "disks": [
    {
      "device": "/dev/sda1",
      "mountpoint": "/",
      "fstype": "xfs",
      "totalBytes": 107374182400,
      "usedBytes": 52010680320,
      "freeBytes": 55363502080,
      "usedPercent": 48.47
    }
  ],
  "processes": [
    {
      "pid": 1,
      "name": "systemd",
      "memoryRssBytes": 12288000,
      "cpuPercent": 0.01
    }
  ]
}
```

## Arquitectura

```
recolector/
├── cmd/recolector/        # Punto de entrada CLI, wiring y loop de intervalo
├── internal/
│   ├── config/            # Parseo y validación de flags (FromFlags)
│   ├── collector/         # Recolección de métricas del SO (5 sub-colectores)
│   └── exporter/          # Exportadores: stdout, file, HTTP POST
├── go.mod
└── go.sum
```

Las dos interfaces centrales son:

```go
type Collector interface {
    Collect(ctx context.Context) (*SystemInfo, error)
}

type Exporter interface {
    Export(ctx context.Context, info *collector.SystemInfo) error
}
```

`main.go` conecta todo: parsea flags → construye el Exporter → instancia el Collector → ejecuta `run()`. La función `run()` acepta interfaces, lo que la hace testeable sin llamadas reales al sistema operativo.

## Dependencias

| Paquete | Versión | Uso |
|---------|---------|-----|
| `github.com/shirou/gopsutil/v3` | v3.24.5 | Métricas de CPU, memoria, disco, procesos y host |
| `github.com/denisbrodbeck/machineid` | v1.0.1 | UUID estable del hardware del equipo |

## Desarrollo

```bash
# Ejecutar en modo desarrollo (recomendado en macOS)
go run ./cmd/recolector -pretty

# Tests con detector de condiciones de carrera
go test -race ./...

# Verificación estática
go vet ./...

# Limpiar dependencias
go mod tidy
```

> **Nota macOS:** El binario compilado requiere firma de código para enumerar
> procesos del sistema. Durante el desarrollo usar `go run`. En producción,
> firmar con el entitlement `com.apple.security.get-task-allow`.

## Licencia

MIT
