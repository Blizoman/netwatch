// Package config loads netwatch's YAML configuration file and applies
// sane defaults.
package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/Blizoman/netwatch/internal/target"
)

const (
	DefaultTimeout     = 5 * time.Second
	DefaultInterval    = 30 * time.Second
	DefaultPingCount   = 5
	DefaultMaxHops     = 30
	DefaultConcurrency = 8
)

// Duration wraps time.Duration so it can be parsed from YAML strings like
// "5s" or "1m30s" instead of raw nanoseconds.
type Duration time.Duration

func (d Duration) Duration() time.Duration { return time.Duration(d) }

func (d *Duration) UnmarshalYAML(node *yaml.Node) error {
	var s string
	if err := node.Decode(&s); err != nil {
		return err
	}
	parsed, err := time.ParseDuration(s)
	if err != nil {
		return fmt.Errorf("invalid duration %q: %w", s, err)
	}
	*d = Duration(parsed)
	return nil
}

func (d Duration) MarshalYAML() (interface{}, error) {
	return time.Duration(d).String(), nil
}

// Config is the top-level configuration file schema.
type Config struct {
	Interval    Duration        `yaml:"interval,omitempty"`
	Timeout     Duration        `yaml:"timeout,omitempty"`
	PingCount   int             `yaml:"ping_count,omitempty"`
	MaxHops     int             `yaml:"max_hops,omitempty"`
	Concurrency int             `yaml:"concurrency,omitempty"`
	Output      string          `yaml:"output,omitempty"`
	NoColor     bool            `yaml:"no_color,omitempty"`
	Targets     []target.Target `yaml:"targets"`
}

// Default returns a Config populated with netwatch's built-in defaults.
func Default() Config {
	return Config{
		Interval:    Duration(DefaultInterval),
		Timeout:     Duration(DefaultTimeout),
		PingCount:   DefaultPingCount,
		MaxHops:     DefaultMaxHops,
		Concurrency: DefaultConcurrency,
		Output:      "human",
	}
}

// Load reads and parses a YAML config file at path, filling in defaults
// for any field left unset.
func Load(path string) (Config, error) {
	cfg := Default()

	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}

	loaded := Config{}
	if err := yaml.Unmarshal(data, &loaded); err != nil {
		return Config{}, fmt.Errorf("parse config: %w", err)
	}

	if loaded.Interval > 0 {
		cfg.Interval = loaded.Interval
	}
	if loaded.Timeout > 0 {
		cfg.Timeout = loaded.Timeout
	}
	if loaded.PingCount > 0 {
		cfg.PingCount = loaded.PingCount
	}
	if loaded.MaxHops > 0 {
		cfg.MaxHops = loaded.MaxHops
	}
	if loaded.Concurrency > 0 {
		cfg.Concurrency = loaded.Concurrency
	}
	if loaded.Output != "" {
		cfg.Output = loaded.Output
	}
	cfg.NoColor = loaded.NoColor
	cfg.Targets = loaded.Targets

	for i, t := range cfg.Targets {
		if len(t.Checks) == 0 {
			cfg.Targets[i].Checks = []target.Check{target.CheckDNS, target.CheckTCP, target.CheckPing}
		}
		if err := cfg.Targets[i].Validate(); err != nil {
			return Config{}, fmt.Errorf("config: %w", err)
		}
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

// Validate checks global settings for sane values.
func (c Config) Validate() error {
	if c.Output != "human" && c.Output != "json" {
		return fmt.Errorf("output must be \"human\" or \"json\", got %q", c.Output)
	}
	if c.Timeout.Duration() <= 0 {
		return fmt.Errorf("timeout must be positive")
	}
	if c.Interval.Duration() <= 0 {
		return fmt.Errorf("interval must be positive")
	}
	if c.Concurrency <= 0 {
		return fmt.Errorf("concurrency must be positive")
	}
	return nil
}
