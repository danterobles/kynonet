package exporter

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/danterobles/recolector/internal/collector"
)

// StdoutExporter writes JSON to an io.Writer (default: os.Stdout).
type StdoutExporter struct {
	pretty bool
	w      io.Writer
}

// StdoutOption configures a StdoutExporter.
type StdoutOption func(*StdoutExporter)

// WithWriter replaces the default os.Stdout writer (useful for testing).
func WithWriter(w io.Writer) StdoutOption {
	return func(e *StdoutExporter) {
		e.w = w
	}
}

// NewStdoutExporter returns a StdoutExporter ready to use.
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

// Export marshals info to JSON and writes it to the configured writer.
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
