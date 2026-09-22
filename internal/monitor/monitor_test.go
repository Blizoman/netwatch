package monitor

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/Blizoman/netwatch/internal/checker"
	"github.com/Blizoman/netwatch/internal/target"
)

func listenerTarget(t *testing.T) (target.Target, func()) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			_ = conn.Close()
		}
	}()
	port := ln.Addr().(*net.TCPAddr).Port
	return target.Target{
		Host:   "127.0.0.1",
		Port:   port,
		Checks: []target.Check{target.CheckDNS, target.CheckTCP},
	}, func() { _ = ln.Close() }
}

func TestRunOnce(t *testing.T) {
	tg, cleanup := listenerTarget(t)
	defer cleanup()

	m := New([]target.Target{tg}, Options{Concurrency: 2})
	results := m.RunOnce(context.Background())

	if len(results) != 2 {
		t.Fatalf("len(results) = %d, want 2", len(results))
	}
	// Order must follow target.Checks (dns, then tcp) for stable output.
	if results[0].Type != checker.TypeDNS {
		t.Errorf("results[0].Type = %v, want dns", results[0].Type)
	}
	if results[1].Type != checker.TypeTCP {
		t.Errorf("results[1].Type = %v, want tcp", results[1].Type)
	}
	for _, r := range results {
		if !r.Success {
			t.Errorf("check %s failed: %s", r.Type, r.Error)
		}
	}
}

func TestRunOnceUnknownCheck(t *testing.T) {
	m := New([]target.Target{{
		Host:   "example.com",
		Checks: []target.Check{"bogus"},
	}}, Options{})
	results := m.RunOnce(context.Background())

	if len(results) != 1 || results[0].Success {
		t.Fatalf("expected a single failed result, got %+v", results)
	}
}

func TestWatchStopsOnCancel(t *testing.T) {
	tg, cleanup := listenerTarget(t)
	defer cleanup()

	m := New([]target.Target{tg}, Options{})
	ctx, cancel := context.WithCancel(context.Background())

	var rounds int
	done := make(chan struct{})
	go func() {
		m.Watch(ctx, 10*time.Millisecond, func(results []checker.Result) {
			rounds++
			if rounds == 2 {
				cancel()
			}
		})
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Watch did not stop after context cancellation")
	}
	if rounds < 2 {
		t.Errorf("rounds = %d, want at least 2", rounds)
	}
}
