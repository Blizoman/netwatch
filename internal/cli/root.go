// Package cli wires netwatch's cobra commands to the config, monitor and
// output packages.
package cli

import (
	"errors"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/Blizoman/netwatch/internal/config"
	"github.com/Blizoman/netwatch/internal/target"
)

// ErrChecksFailed is returned by "check" and "watch" when at least one
// check failed, so main can set a non-zero exit code without cobra
// printing a redundant "Error: ..." line on top of the report already
// written to stdout.
var ErrChecksFailed = errors.New("one or more checks failed")

// NewRootCmd builds the netwatch command tree.
func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "netwatch",
		Short:         "Cross-platform CLI network diagnostics",
		Long:          "netwatch monitors hosts, TCP ports, DNS resolution, latency,\npacket loss and network paths (traceroute), with human-readable\nor JSON output and one-shot or continuous monitoring modes.",
		SilenceUsage:  true,
		SilenceErrors: false,
	}
	root.AddCommand(newCheckCmd(), newWatchCmd(), newVersionCmd())
	return root
}

// sharedFlags are the flags common to "check" and "watch".
type sharedFlags struct {
	configPath  string
	targets     []string
	checks      string
	timeout     time.Duration
	pingCount   int
	maxHops     int
	concurrency int
	output      string
	noColor     bool
}

func addSharedFlags(cmd *cobra.Command, f *sharedFlags) {
	cmd.Flags().StringVarP(&f.configPath, "config", "c", "", "path to a YAML config file (overrides --target)")
	cmd.Flags().StringArrayVarP(&f.targets, "target", "t", nil, "target to check: host, host:port, or host:port/checks (repeatable)")
	cmd.Flags().StringVar(&f.checks, "checks", "dns,tcp,ping", "default checks for --target entries without an explicit /checks suffix")
	cmd.Flags().DurationVar(&f.timeout, "timeout", config.DefaultTimeout, "per-check timeout")
	cmd.Flags().IntVar(&f.pingCount, "ping-count", config.DefaultPingCount, "number of ICMP echo requests per ping check")
	cmd.Flags().IntVar(&f.maxHops, "max-hops", config.DefaultMaxHops, "maximum TTL probed by traceroute")
	cmd.Flags().IntVar(&f.concurrency, "concurrency", config.DefaultConcurrency, "maximum number of checks to run at once")
	cmd.Flags().StringVarP(&f.output, "output", "o", "human", "output format: human or json")
	cmd.Flags().BoolVar(&f.noColor, "no-color", false, "disable colored output")
}

// resolve merges CLI flags and an optional config file into a single
// Config, with explicitly-set CLI flags taking precedence over the file.
func (f *sharedFlags) resolve(cmd *cobra.Command) (config.Config, error) {
	var cfg config.Config

	if f.configPath != "" {
		loaded, err := config.Load(f.configPath)
		if err != nil {
			return config.Config{}, err
		}
		cfg = loaded
	} else {
		cfg = config.Default()
		defaultChecks := target.ParseChecks(f.checks)
		for _, spec := range f.targets {
			t, err := target.ParseSpec(spec, defaultChecks)
			if err != nil {
				return config.Config{}, err
			}
			cfg.Targets = append(cfg.Targets, t)
		}
		if len(cfg.Targets) == 0 {
			return config.Config{}, fmt.Errorf("no targets specified: pass --target/-t one or more times, or --config <file>")
		}
	}

	if cmd.Flags().Changed("timeout") {
		cfg.Timeout = config.Duration(f.timeout)
	}
	if cmd.Flags().Changed("ping-count") {
		cfg.PingCount = f.pingCount
	}
	if cmd.Flags().Changed("max-hops") {
		cfg.MaxHops = f.maxHops
	}
	if cmd.Flags().Changed("concurrency") {
		cfg.Concurrency = f.concurrency
	}
	if cmd.Flags().Changed("output") {
		cfg.Output = f.output
	}
	if cmd.Flags().Changed("no-color") {
		cfg.NoColor = f.noColor
	}

	return cfg, cfg.Validate()
}
