package config

import (
	"testing"
	"time"
)

func TestFromFlags(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    Config
		wantErr string
	}{
		{
			name: "defaults to stdout",
			args: []string{},
			want: Config{Output: "stdout"},
		},
		{
			name: "stdout explicit",
			args: []string{"-output", "stdout", "-pretty"},
			want: Config{Output: "stdout", Pretty: true},
		},
		{
			name: "file output",
			args: []string{"-output", "file", "-file", "/tmp/out.json"},
			want: Config{Output: "file", FilePath: "/tmp/out.json"},
		},
		{
			name:    "file output missing path",
			args:    []string{"-output", "file"},
			wantErr: "output=file requires -file",
		},
		{
			name: "post output with token",
			args: []string{"-output", "post", "-endpoint", "http://localhost:8080/collect", "-token", "secret"},
			want: Config{Output: "post", Endpoint: "http://localhost:8080/collect", Token: "secret"},
		},
		{
			name:    "post output missing endpoint",
			args:    []string{"-output", "post"},
			wantErr: "output=post requires -endpoint",
		},
		{
			name:    "post output invalid URL",
			args:    []string{"-output", "post", "-endpoint", "not-a-url"},
			wantErr: "invalid -endpoint URL",
		},
		{
			name:    "unknown output mode",
			args:    []string{"-output", "ftp"},
			wantErr: "unknown output mode",
		},
		{
			name: "interval conversion",
			args: []string{"-interval", "5"},
			want: Config{Output: "stdout", Interval: 5 * time.Second},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FromFlags(tt.args)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErr)
				}
				if msg := err.Error(); len(msg) == 0 {
					t.Fatalf("error message is empty")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Output != tt.want.Output {
				t.Errorf("Output: got %q, want %q", got.Output, tt.want.Output)
			}
			if got.FilePath != tt.want.FilePath {
				t.Errorf("FilePath: got %q, want %q", got.FilePath, tt.want.FilePath)
			}
			if got.Endpoint != tt.want.Endpoint {
				t.Errorf("Endpoint: got %q, want %q", got.Endpoint, tt.want.Endpoint)
			}
			if got.Token != tt.want.Token {
				t.Errorf("Token: got %q, want %q", got.Token, tt.want.Token)
			}
			if got.Pretty != tt.want.Pretty {
				t.Errorf("Pretty: got %v, want %v", got.Pretty, tt.want.Pretty)
			}
			if got.Interval != tt.want.Interval {
				t.Errorf("Interval: got %v, want %v", got.Interval, tt.want.Interval)
			}
		})
	}
}
