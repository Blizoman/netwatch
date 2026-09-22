package checker

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestPingOptionsDefaults(t *testing.T) {
	o := PingOptions{}.withDefaults()
	if o.Count != 5 {
		t.Errorf("Count = %d, want 5", o.Count)
	}
	if o.Interval != time.Second {
		t.Errorf("Interval = %v, want 1s", o.Interval)
	}
	if o.Timeout <= 0 {
		t.Errorf("Timeout = %v, want positive", o.Timeout)
	}
}

// TestCheckPingLoopback exercises the real ICMP path against 127.0.0.1.
// It is skipped in sandboxed CI environments that permit neither
// unprivileged ("ping socket") nor privileged (raw socket) ICMP, since
// that is an environment limitation rather than a netwatch bug.
func TestCheckPingLoopback(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	res := CheckPing(ctx, "127.0.0.1", PingOptions{Count: 2, Interval: 100 * time.Millisecond, Timeout: 5 * time.Second})

	if !res.Success && looksLikePermissionError(res.Error) {
		t.Skipf("ICMP not permitted in this environment: %s", res.Error)
	}
	if !res.Success {
		t.Fatalf("expected success pinging loopback, got error: %s", res.Error)
	}
	if res.Ping == nil || res.Ping.PacketsRecv == 0 {
		t.Errorf("Ping details = %+v", res.Ping)
	}
}

func looksLikePermissionError(msg string) bool {
	return strings.Contains(msg, "operation not permitted") ||
		strings.Contains(msg, "permission denied") ||
		strings.Contains(msg, "socket type not supported")
}
