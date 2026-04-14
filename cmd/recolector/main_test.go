package main

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/danterobles/recolector/internal/collector"
)

// fakeCollector returns a canned SystemInfo on every call.
type fakeCollector struct {
	calls atomic.Int64
}

func (f *fakeCollector) Collect(_ context.Context) (*collector.SystemInfo, error) {
	f.calls.Add(1)
	return &collector.SystemInfo{CollectedAt: time.Now()}, nil
}

// fakeExporter records every export call.
type fakeExporter struct {
	calls atomic.Int64
}

func (f *fakeExporter) Export(_ context.Context, _ *collector.SystemInfo) error {
	f.calls.Add(1)
	return nil
}

func TestRun_singleShot(t *testing.T) {
	col := &fakeCollector{}
	exp := &fakeExporter{}

	err := run(context.Background(), col, exp, 0)
	if err != nil {
		t.Fatalf("run() error: %v", err)
	}
	if col.calls.Load() != 1 {
		t.Errorf("Collect called %d times, want 1", col.calls.Load())
	}
	if exp.calls.Load() != 1 {
		t.Errorf("Export called %d times, want 1", exp.calls.Load())
	}
}

func TestRun_interval_cancelClean(t *testing.T) {
	col := &fakeCollector{}
	exp := &fakeExporter{}

	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()

	err := run(ctx, col, exp, 50*time.Millisecond)
	if err != nil && !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("run() unexpected error: %v", err)
	}

	// At least the initial call + a couple of ticks should have fired.
	if exp.calls.Load() < 2 {
		t.Errorf("Export called %d times, want >= 2", exp.calls.Load())
	}
}

func TestRun_collectError_propagates(t *testing.T) {
	bad := &badCollector{}
	exp := &fakeExporter{}

	err := run(context.Background(), bad, exp, 0)
	if err == nil {
		t.Fatal("expected error from bad collector, got nil")
	}
}

type badCollector struct{}

func (b *badCollector) Collect(_ context.Context) (*collector.SystemInfo, error) {
	return nil, errors.New("simulated collect failure")
}
