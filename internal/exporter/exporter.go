// Package exporter define la interfaz Exporter y sus tres implementaciones:
// StdoutExporter (consola), FileExporter (archivo) y HTTPExporter (POST HTTP).
//
// La interfaz se declara aquí, en el lado del consumidor, siguiendo la convención
// de Go: las interfaces pertenecen al paquete que las usa, no al que las implementa.
// Esto facilita sustituir cualquier implementación con un mock en los tests.
package exporter

import (
	"context"

	"github.com/danterobles/recolector/internal/collector"
)

// Exporter escribe una instantánea del sistema (SystemInfo) en algún destino.
// Cada llamada a Export corresponde a un ciclo de recolección completo.
// El contexto permite cancelar operaciones de I/O de larga duración (especialmente útil en HTTPExporter).
type Exporter interface {
	Export(ctx context.Context, info *collector.SystemInfo) error
}
