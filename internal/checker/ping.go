package checker

import (
	"context"
	"fmt"
	"time"

	probing "github.com/prometheus-community/pro-bing"
)

// PingOptions configures an ICMP echo check.
type PingOptions struct {
	Count    int           // number of echo requests to send
	Interval time.Duration // delay between echo requests
	Timeout  time.Duration // overall deadline for the whole check
}

func (o PingOptions) withDefaults() PingOptions {
	if o.Count <= 0 {
		o.Count = 5
	}
	if o.Interval <= 0 {
		o.Interval = time.Second
	}
	if o.Timeout <= 0 {
		o.Timeout = time.Duration(o.Count)*o.Interval + 5*time.Second
	}
	return o
}

// CheckPing sends ICMP echo requests to host and reports round-trip time
// and packet-loss statistics.
//
// It first tries an unprivileged (SOCK_DGRAM) ICMP socket, which needs no
// special permissions on Linux when net.ipv4.ping_group_range permits it
// and works out of the box on macOS. If that fails, it falls back to a
// privileged raw socket, which requires root or CAP_NET_RAW.
func CheckPing(ctx context.Context, host string, opts PingOptions) Result {
	opts = opts.withDefaults()
	start := time.Now()
	res := Result{
		Target:    host,
		Type:      TypePing,
		Timestamp: start,
	}

	stats, privileged, err := runPing(ctx, host, opts)
	res.Duration = Duration(time.Since(start))
	if err != nil {
		res.Success = false
		res.Error = err.Error()
		return res
	}

	res.Ping = &PingDetails{
		PacketsSent:       stats.PacketsSent,
		PacketsRecv:       stats.PacketsRecv,
		PacketLossPercent: stats.PacketLoss,
		MinRTT:            Duration(stats.MinRtt),
		AvgRTT:            Duration(stats.AvgRtt),
		MaxRTT:            Duration(stats.MaxRtt),
		StdDevRTT:         Duration(stats.StdDevRtt),
		Privileged:        privileged,
	}
	res.Success = stats.PacketsRecv > 0
	if !res.Success {
		res.Error = "100% packet loss"
	}
	return res
}

func runPing(ctx context.Context, host string, opts PingOptions) (*probing.Statistics, bool, error) {
	var lastErr error
	for _, privileged := range []bool{false, true} {
		stats, err := doPing(ctx, host, opts, privileged)
		if err == nil {
			return stats, privileged, nil
		}
		lastErr = err
	}
	return nil, false, fmt.Errorf("icmp ping failed (tried unprivileged and privileged modes, the latter requires root/CAP_NET_RAW): %w", lastErr)
}

func doPing(ctx context.Context, host string, opts PingOptions, privileged bool) (*probing.Statistics, error) {
	pinger, err := probing.NewPinger(host)
	if err != nil {
		return nil, err
	}
	pinger.SetPrivileged(privileged)
	pinger.Count = opts.Count
	pinger.Interval = opts.Interval
	pinger.Timeout = opts.Timeout

	if err := pinger.RunWithContext(ctx); err != nil {
		return nil, err
	}
	return pinger.Statistics(), nil
}
