// Package monitor orchestrates running checks concurrently across a set
// of targets, either once or on a repeating interval.
package monitor

import (
	"context"
	"sync"
	"time"

	"github.com/Blizoman/netwatch/internal/checker"
	"github.com/Blizoman/netwatch/internal/target"
)

// Options controls how checks are executed.
type Options struct {
	Timeout     time.Duration
	PingCount   int
	MaxHops     int
	Concurrency int
}

func (o Options) withDefaults() Options {
	if o.Timeout <= 0 {
		o.Timeout = 5 * time.Second
	}
	if o.PingCount <= 0 {
		o.PingCount = 5
	}
	if o.MaxHops <= 0 {
		o.MaxHops = 30
	}
	if o.Concurrency <= 0 {
		o.Concurrency = 8
	}
	return o
}

// Monitor runs the configured checks against a fixed set of targets.
type Monitor struct {
	targets []target.Target
	opts    Options
}

// New builds a Monitor for the given targets.
func New(targets []target.Target, opts Options) *Monitor {
	return &Monitor{targets: targets, opts: opts.withDefaults()}
}

type job struct {
	target target.Target
	check  target.Check
}

// RunOnce runs every configured check against every target concurrently,
// bounded by Options.Concurrency, and returns results ordered by target
// then check so output is stable across runs.
func (m *Monitor) RunOnce(ctx context.Context) []checker.Result {
	var jobs []job
	for _, t := range m.targets {
		for _, c := range t.Checks {
			jobs = append(jobs, job{target: t, check: c})
		}
	}

	results := make([]checker.Result, len(jobs))
	sem := make(chan struct{}, m.opts.Concurrency)
	var wg sync.WaitGroup

	for i, j := range jobs {
		wg.Add(1)
		go func(i int, j job) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			results[i] = m.runCheck(ctx, j.target, j.check)
		}(i, j)
	}

	wg.Wait()
	return results
}

func (m *Monitor) runCheck(ctx context.Context, t target.Target, c target.Check) checker.Result {
	switch c {
	case target.CheckDNS:
		return checker.CheckDNS(ctx, checker.DefaultResolver(), t.Host)
	case target.CheckTCP:
		return checker.CheckTCP(ctx, t.Host, t.Port, m.opts.Timeout)
	case target.CheckPing:
		return checker.CheckPing(ctx, t.Host, checker.PingOptions{
			Count:   m.opts.PingCount,
			Timeout: m.opts.Timeout + time.Duration(m.opts.PingCount)*time.Second,
		})
	case target.CheckTraceroute:
		return checker.CheckTraceroute(ctx, t.Host, checker.TracerouteOptions{
			MaxHops:      m.opts.MaxHops,
			ProbeTimeout: m.opts.Timeout,
		})
	default:
		return checker.Result{
			Target:  t.Name(),
			Success: false,
			Error:   "unknown check type: " + string(c),
		}
	}
}

// Watch runs RunOnce immediately and then every interval, invoking fn with
// each round's results, until ctx is cancelled.
func (m *Monitor) Watch(ctx context.Context, interval time.Duration, fn func([]checker.Result)) {
	fn(m.RunOnce(ctx))

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			fn(m.RunOnce(ctx))
		}
	}
}
