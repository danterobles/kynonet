package config

import (
	"flag"
	"fmt"
	"net/url"
	"time"
)

// Config holds all runtime configuration parsed from CLI flags.
type Config struct {
	Output   string        // "stdout" | "file" | "post"
	FilePath string        // used when Output == "file"
	Endpoint string        // used when Output == "post"
	Token    string        // Bearer token for HTTP POST, may be empty
	Pretty   bool          // pretty-print JSON
	Interval time.Duration // 0 = run once
}

// FromFlags parses args (os.Args[1:]) using a local FlagSet and returns a validated Config.
func FromFlags(args []string) (Config, error) {
	fs := flag.NewFlagSet("recolector", flag.ContinueOnError)

	var (
		output   = fs.String("output", "stdout", `Output mode: stdout | file | post`)
		filePath = fs.String("file", "", "File path for file output")
		endpoint = fs.String("endpoint", "", "HTTP endpoint for POST output")
		token    = fs.String("token", "", "Bearer token for POST auth")
		pretty   = fs.Bool("pretty", false, "Pretty-print JSON")
		interval = fs.Int("interval", 0, "Repeat collection every N seconds (0 = run once)")
	)

	if err := fs.Parse(args); err != nil {
		return Config{}, err
	}

	cfg := Config{
		Output:   *output,
		FilePath: *filePath,
		Endpoint: *endpoint,
		Token:    *token,
		Pretty:   *pretty,
		Interval: time.Duration(*interval) * time.Second,
	}

	if err := cfg.validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (c Config) validate() error {
	switch c.Output {
	case "stdout":
		// nothing required
	case "file":
		if c.FilePath == "" {
			return fmt.Errorf("output=file requires -file <path>")
		}
	case "post":
		if c.Endpoint == "" {
			return fmt.Errorf("output=post requires -endpoint <url>")
		}
		if _, err := url.ParseRequestURI(c.Endpoint); err != nil {
			return fmt.Errorf("invalid -endpoint URL %q: %w", c.Endpoint, err)
		}
	default:
		return fmt.Errorf("unknown output mode %q: must be stdout, file, or post", c.Output)
	}

	if c.Interval < 0 {
		return fmt.Errorf("-interval must be >= 0")
	}

	return nil
}
