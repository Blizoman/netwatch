package cli

import (
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/Blizoman/netwatch/internal/monitor"
	"github.com/Blizoman/netwatch/internal/output"
)

func newCheckCmd() *cobra.Command {
	f := &sharedFlags{}
	cmd := &cobra.Command{
		Use:   "check",
		Short: "Run every configured check once and exit",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := f.resolve(cmd)
			if err != nil {
				return err
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
			results := mon.RunOnce(ctx)
			if err := w.Write(output.Round{Timestamp: time.Now(), Results: results}); err != nil {
				return err
			}

			if _, failed := output.Summarize(results); failed > 0 {
				cmd.SilenceErrors = true
				return ErrChecksFailed
			}
			return nil
		},
	}
	addSharedFlags(cmd, f)
	return cmd
}
