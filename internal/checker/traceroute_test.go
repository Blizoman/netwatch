package checker

import (
	"context"
	"testing"
	"time"
)

func TestTracerouteOptionsDefaults(t *testing.T) {
	o := TracerouteOptions{}.withDefaults()
	if o.MaxHops != 30 {
		t.Errorf("MaxHops = %d, want 30", o.MaxHops)
	}
	if o.ProbeTimeout != time.Second {
		t.Errorf("ProbeTimeout = %v, want 1s", o.ProbeTimeout)
	}
}

func TestCheckTracerouteInvalidHost(t *testing.T) {
	res := CheckTraceroute(context.Background(), "this.host.does.not.exist.invalid", TracerouteOptions{})
	if res.Success {
		t.Fatal("expected failure resolving a bogus host")
	}
	if res.Error == "" {
		t.Error("expected a non-empty error message")
	}
}

// TestCheckTracerouteLoopback exercises the real raw-ICMP path. It is
// skipped when the process lacks CAP_NET_RAW/root, since that is an
// environment limitation rather than a netwatch bug -- the same
// privilege the standard traceroute/tracert tools require.
func TestCheckTracerouteLoopback(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	res := CheckTraceroute(ctx, "127.0.0.1", TracerouteOptions{MaxHops: 5, ProbeTimeout: time.Second})

	if !res.Success && looksLikePermissionError(res.Error) {
		t.Skipf("raw ICMP not permitted in this environment: %s", res.Error)
	}
	if !res.Success {
		t.Fatalf("expected to reach loopback, got error: %s", res.Error)
	}
	if res.Traceroute == nil || len(res.Traceroute.Hops) == 0 {
		t.Errorf("Traceroute details = %+v", res.Traceroute)
	}
}
