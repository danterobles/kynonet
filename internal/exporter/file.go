package exporter

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/danterobles/recolector/internal/collector"
)

// FileExporter writes JSON to a file path, overwriting on each export cycle.
type FileExporter struct {
	path   string
	pretty bool
}

// NewFileExporter returns a FileExporter that writes to path.
func NewFileExporter(path string, pretty bool) *FileExporter {
	return &FileExporter{path: path, pretty: pretty}
}

// Export marshals info to JSON and writes it to the configured file.
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
