package exporter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/danterobles/recolector/internal/collector"
)

// HTTPExporter sends JSON via HTTP POST to a configurable endpoint.
type HTTPExporter struct {
	endpoint string
	token    string
	client   *http.Client
	pretty   bool
}

// HTTPOption configures an HTTPExporter.
type HTTPOption func(*HTTPExporter)

// WithHTTPClient replaces the default http.Client (useful for testing).
func WithHTTPClient(c *http.Client) HTTPOption {
	return func(e *HTTPExporter) {
		e.client = c
	}
}

// WithHTTPTimeout sets the HTTP client timeout.
func WithHTTPTimeout(d time.Duration) HTTPOption {
	return func(e *HTTPExporter) {
		e.client = &http.Client{Timeout: d}
	}
}

// NewHTTPExporter returns an HTTPExporter ready to use.
func NewHTTPExporter(endpoint, token string, pretty bool, opts ...HTTPOption) *HTTPExporter {
	e := &HTTPExporter{
		endpoint: endpoint,
		token:    token,
		pretty:   pretty,
		client:   &http.Client{Timeout: 10 * time.Second},
	}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

// Export marshals info to JSON and POSTs it to the configured endpoint.
func (e *HTTPExporter) Export(ctx context.Context, info *collector.SystemInfo) error {
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

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, e.endpoint, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if e.token != "" {
		req.Header.Set("Authorization", "Bearer "+e.token)
	}

	resp, err := e.client.Do(req)
	if err != nil {
		return fmt.Errorf("post to %s: %w", e.endpoint, err)
	}
	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("server returned %d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	return nil
}
