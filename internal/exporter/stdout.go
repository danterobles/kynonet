package exporter

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/danterobles/recolector/internal/collector"
)

// StdoutExporter serializa el SystemInfo como JSON y lo escribe en un io.Writer.
// Por defecto el writer es os.Stdout, pero puede ser reemplazado mediante
// la opción WithWriter para facilitar las pruebas unitarias sin capturar stdout.
type StdoutExporter struct {
	pretty bool      // si es true, usa json.MarshalIndent con 2 espacios
	w      io.Writer // destino de escritura; por defecto os.Stdout
}

// StdoutOption es el tipo funcional para configurar un StdoutExporter.
// Sigue el patrón "functional options" para mantener compatibilidad hacia adelante.
type StdoutOption func(*StdoutExporter)

// WithWriter reemplaza el writer de salida del StdoutExporter.
// Útil en tests para inyectar un bytes.Buffer y verificar el JSON generado
// sin necesidad de redirigir os.Stdout.
func WithWriter(w io.Writer) StdoutOption {
	return func(e *StdoutExporter) {
		e.w = w
	}
}

// NewStdoutExporter construye un StdoutExporter con os.Stdout como destino por defecto.
// Si pretty es true, el JSON de salida se formatea con indentación de 2 espacios.
// Las opciones adicionales permiten sobreescribir el writer en tiempo de construcción.
func NewStdoutExporter(pretty bool, opts ...StdoutOption) *StdoutExporter {
	e := &StdoutExporter{
		pretty: pretty,
		w:      os.Stdout,
	}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

// Export serializa info a JSON y lo escribe en el writer configurado,
// seguido de un salto de línea para facilitar el procesamiento con herramientas
// de línea de comandos como jq o grep.
func (e *StdoutExporter) Export(_ context.Context, info *collector.SystemInfo) error {
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

	_, err = fmt.Fprintf(e.w, "%s\n", data)
	return err
}
