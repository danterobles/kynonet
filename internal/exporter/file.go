package exporter

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/danterobles/recolector/internal/collector"
)

// FileExporter serializa el SystemInfo como JSON y lo escribe en un archivo del sistema.
// En cada ciclo de recolección el archivo se sobreescribe completamente, de forma que
// siempre refleja la instantánea más reciente. Esto es el comportamiento esperado
// cuando se usa junto con -interval para monitoreo continuo.
type FileExporter struct {
	path   string // ruta absoluta o relativa del archivo de salida
	pretty bool   // si es true, usa json.MarshalIndent con 2 espacios
}

// NewFileExporter construye un FileExporter que escribirá en path.
// No crea ni valida el archivo al momento de la construcción; los errores
// de ruta o permisos se reportan en la primera llamada a Export.
func NewFileExporter(path string, pretty bool) *FileExporter {
	return &FileExporter{path: path, pretty: pretty}
}

// Export serializa info a JSON y lo escribe en el archivo configurado con
// permisos 0644 (lectura para todos, escritura solo para el propietario).
// Si el archivo ya existe es sobreescrito; si el directorio padre no existe,
// la operación falla con un error descriptivo.
func (e *FileExporter) Export(_ context.Context, info *collector.SystemInfo) error {
	var (
		data []byte
		err  error
	)

	if e.pretty {
		data, err = json.MarshalIndent(info, "", "  ")
	} else {
		data, err = json.Marshal(info)
	}
	if err != nil {
		return fmt.Errorf("marshal json: %w", err)
	}

	if err := os.WriteFile(e.path, data, 0o644); err != nil {
		return fmt.Errorf("write file %s: %w", e.path, err)
	}

	return nil
}
