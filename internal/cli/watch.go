package cli

import (
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/Blizoman/netwatch/internal/checker"
	"github.com/Blizoman/netwatch/internal/config"
	"github.com/Blizoman/netwatch/internal/monitor"
	"github.com/Blizoman/netwatch/internal/output"
)

func newWatchCmd() *cobra.Command {
	f := &sharedFlags{}
	var interval time.Duration

	cmd := &cobra.Command{
		Use:   "watch",
		Short: "Continuously re-run every configured check until interrupted",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := f.resolve(cmd)
			if err != nil {
				return err
			}
			if cmd.Flags().Changed("interval") {
				cfg.Interval = config.Duration(interval)
			}

			ctx, cancel := signal.NotifyContext(cmd.Context(), syscall.SIGINT, syscall.SIGTERM)
			defer cancel()

			mon := monitor.New(cfg.Targets, monitor.Options{
				Timeout:     cfg.Timeout.Duration(),
				PingCount:   cfg.PingCount,
				MaxHops:     cfg.MaxHops,
				Concurrency: cfg.Concurrency,
			})
			w := output.New(cmd.OutOrStdout(), cfg.Output, cfg.NoColor)

			var anyFailed atomic.Bool
			var writeErr error
			mon.Watch(ctx, cfg.Interval.Duration(), func(results []checker.Result) {
				if _, failed := output.Summarize(results); failed > 0 {
					anyFailed.Store(true)
				}
				if err := w.Write(output.Round{Timestamp: time.Now(), Results: results}); err != nil {
					writeErr = err
				}
			})

			if writeErr != nil {
				return writeErr
			}
			if anyFailed.Load() {
				cmd.SilenceErrors = true
				return ErrChecksFailed
			}
			return nil
		},
	}
	addSharedFlags(cmd, f)
	cmd.Flags().DurationVarP(&interval, "interval", "i", config.DefaultInterval, "how often to re-run all checks")
	return cmd
}
