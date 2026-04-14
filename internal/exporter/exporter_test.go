package exporter

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/danterobles/recolector/internal/collector"
)

// sampleInfo returns a minimal SystemInfo for use in tests.
func sampleInfo() *collector.SystemInfo {
	return &collector.SystemInfo{
		CollectedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		Host: collector.HostInfo{
			Hostname: "testhost",
			OS:       "linux",
		},
		CPU: collector.CPUInfo{
			ModelName:    "Test CPU",
			LogicalCores: 4,
		},
	}
}

// --- StdoutExporter ---

func TestStdoutExporter_compact(t *testing.T) {
	var buf bytes.Buffer
	exp := NewStdoutExporter(false, WithWriter(&buf))

	if err := exp.Export(context.Background(), sampleInfo()); err != nil {
		t.Fatalf("Export() error: %v", err)
	}

	if !json.Valid(buf.Bytes()) {
		t.Errorf("output is not valid JSON: %s", buf.String())
	}

	// Compact output must not contain newlines inside the JSON object.
	line := bytes.TrimRight(buf.Bytes(), "\n")
	if bytes.ContainsRune(line, '\n') {
		t.Error("compact output contains unexpected newlines")
	}
}

func TestStdoutExporter_pretty(t *testing.T) {
	var buf bytes.Buffer
	exp := NewStdoutExporter(true, WithWriter(&buf))

	if err := exp.Export(context.Background(), sampleInfo()); err != nil {
		t.Fatalf("Export() error: %v", err)
	}

	if !json.Valid(buf.Bytes()) {
		t.Errorf("output is not valid JSON: %s", buf.String())
	}

	// Pretty output must contain indentation.
	if !bytes.Contains(buf.Bytes(), []byte("  ")) {
		t.Error("pretty output missing indentation")
	}
}

// --- FileExporter ---

func TestFileExporter(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.json")
	exp := NewFileExporter(path, false)

	if err := exp.Export(context.Background(), sampleInfo()); err != nil {
		t.Fatalf("Export() error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error: %v", err)
	}
	if !json.Valid(data) {
		t.Errorf("file content is not valid JSON: %s", data)
	}
}

func TestFileExporter_overwrites(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.json")
	exp := NewFileExporter(path, false)

	info := sampleInfo()
	info.Host.Hostname = "first"
	_ = exp.Export(context.Background(), info)

	info.Host.Hostname = "second"
	_ = exp.Export(context.Background(), info)

	data, _ := os.ReadFile(path)
	if !bytes.Contains(data, []byte("second")) {
		t.Error("file was not overwritten on second Export")
	}
}

// --- HTTPExporter ---

func TestHTTPExporter_success(t *testing.T) {
	var gotBody []byte
	var gotContentType, gotAuth string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotContentType = r.Header.Get("Content-Type")
		gotAuth = r.Header.Get("Authorization")
		buf := new(bytes.Buffer)
		_, _ = buf.ReadFrom(r.Body)
		gotBody = buf.Bytes()
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	exp := NewHTTPExporter(srv.URL, "mytoken", false, WithHTTPClient(srv.Client()))

	if err := exp.Export(context.Background(), sampleInfo()); err != nil {
		t.Fatalf("Export() error: %v", err)
	}

	if gotContentType != "application/json" {
		t.Errorf("Content-Type: got %q, want application/json", gotContentType)
	}
	if gotAuth != "Bearer mytoken" {
		t.Errorf("Authorization: got %q, want Bearer mytoken", gotAuth)
	}
	if !json.Valid(gotBody) {
		t.Errorf("body is not valid JSON: %s", gotBody)
	}
}

func TestHTTPExporter_noToken(t *testing.T) {
	var gotAuth string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	exp := NewHTTPExporter(srv.URL, "", false, WithHTTPClient(srv.Client()))
	_ = exp.Export(context.Background(), sampleInfo())

	if gotAuth != "" {
		t.Errorf("expected no Authorization header, got %q", gotAuth)
	}
}

func TestHTTPExporter_serverError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	exp := NewHTTPExporter(srv.URL, "", false, WithHTTPClient(srv.Client()))
	err := exp.Export(context.Background(), sampleInfo())

	if err == nil {
		t.Fatal("expected error for 500 response, got nil")
	}
}
