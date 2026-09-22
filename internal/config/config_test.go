package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Blizoman/netwatch/internal/target"
)

func writeTemp(t *testing.T, contents string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "netwatch.yaml")
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadFillsDefaults(t *testing.T) {
	path := writeTemp(t, `
targets:
  - host: example.com
    port: 443
`)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Timeout.Duration() != DefaultTimeout {
		t.Errorf("Timeout = %v, want default %v", cfg.Timeout.Duration(), DefaultTimeout)
	}
	if cfg.Interval.Duration() != DefaultInterval {
		t.Errorf("Interval = %v, want default %v", cfg.Interval.Duration(), DefaultInterval)
	}
	if cfg.Output != "human" {
		t.Errorf("Output = %q, want %q", cfg.Output, "human")
	}
	if len(cfg.Targets) != 1 {
		t.Fatalf("len(Targets) = %d, want 1", len(cfg.Targets))
	}
	want := []target.Check{target.CheckDNS, target.CheckTCP, target.CheckPing}
	got := cfg.Targets[0].Checks
	if len(got) != len(want) {
		t.Fatalf("default checks = %v, want %v", got, want)
	}
}

func TestLoadExplicitValues(t *testing.T) {
	path := writeTemp(t, `
interval: 10s
timeout: 2s
ping_count: 3
max_hops: 15
concurrency: 4
output: json
no_color: true
targets:
  - host: 8.8.8.8
    checks: [ping, traceroute]
`)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Interval.Duration() != 10*time.Second {
		t.Errorf("Interval = %v", cfg.Interval.Duration())
	}
	if cfg.PingCount != 3 || cfg.MaxHops != 15 || cfg.Concurrency != 4 {
		t.Errorf("got PingCount=%d MaxHops=%d Concurrency=%d", cfg.PingCount, cfg.MaxHops, cfg.Concurrency)
	}
	if cfg.Output != "json" || !cfg.NoColor {
		t.Errorf("got Output=%q NoColor=%v", cfg.Output, cfg.NoColor)
	}
}

func TestLoadRejectsInvalidOutput(t *testing.T) {
	path := writeTemp(t, `
output: xml
targets:
  - host: example.com
    checks: [dns]
`)
	if _, err := Load(path); err == nil {
		t.Fatal("expected error for invalid output format")
	}
}

func TestLoadRejectsInvalidTarget(t *testing.T) {
	path := writeTemp(t, `
targets:
  - host: example.com
    checks: [tcp]
`)
	if _, err := Load(path); err == nil {
		t.Fatal("expected error for tcp check without a port")
	}
}

func TestLoadMissingFile(t *testing.T) {
	if _, err := Load("/nonexistent/path/netwatch.yaml"); err == nil {
		t.Fatal("expected error for missing file")
	}
}
