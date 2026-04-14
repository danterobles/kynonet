package exporter

import (
	"context"

	"github.com/danterobles/recolector/internal/collector"
)

// Exporter writes a SystemInfo snapshot to some destination.
type Exporter interface {
	Export(ctx context.Context, info *collector.SystemInfo) error
}
