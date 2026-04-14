package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/danterobles/recolector/internal/collector"
	"github.com/danterobles/recolector/internal/config"
	"github.com/danterobles/recolector/internal/exporter"
)

func main() {
	cfg, err := config.FromFlags(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "recolector: %v\n", err)
		os.Exit(1)
	}

	exp, err := buildExporter(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "recolector: %v\n", err)
		os.Exit(1)
	}

	col := collector.New()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := run(ctx, col, exp, cfg.Interval); err != nil && !errors.Is(err, context.Canceled) {
		fmt.Fprintf(os.Stderr, "recolector: %v\n", err)
		os.Exit(1)
	}
}

func buildExporter(cfg config.Config) (exporter.Exporter, error) {
	switch cfg.Output {
	case "stdout", "":
		return exporter.NewStdoutExporter(cfg.Pretty), nil
	case "file":
		return exporter.NewFileExporter(cfg.FilePath, cfg.Pretty), nil
	case "post":
		return exporter.NewHTTPExporter(cfg.Endpoint, cfg.Token, cfg.Pretty), nil
	default:
		return nil, fmt.Errorf("unknown output mode %q", cfg.Output)
	}
}

func run(ctx context.Context, col collector.Collector, exp exporter.Exporter, interval time.Duration) error {
	collect := func() error {
		info, err := col.Collect(ctx)
		if err != nil {
			return fmt.Errorf("collect: %w", err)
		}
		if err := exp.Export(ctx, info); err != nil {
			return fmt.Errorf("export: %w", err)
		}
		return nil
	}

	if err := collect(); err != nil {
		return err
	}
	if interval == 0 {
		return nil
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := collect(); err != nil {
				return err
			}
		}
	}
}
