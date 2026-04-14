package collector

import (
	"context"
	"testing"
)

func TestSystemCollector_Collect(t *testing.T) {
	col := New()
	info, err := col.Collect(context.Background())
	if err != nil {
		t.Fatalf("Collect() error: %v", err)
	}

	if info == nil {
		t.Fatal("Collect() returned nil SystemInfo")
	}
	if info.CollectedAt.IsZero() {
		t.Error("CollectedAt is zero")
	}
	if info.Host.Hostname == "" {
		t.Error("Host.Hostname is empty")
	}
	if info.CPU.LogicalCores <= 0 {
		t.Errorf("CPU.LogicalCores = %d, want > 0", info.CPU.LogicalCores)
	}
	if info.Memory.TotalBytes == 0 {
		t.Error("Memory.TotalBytes is 0")
	}
	if len(info.Disks) == 0 {
		t.Error("Disks is empty")
	}
}
