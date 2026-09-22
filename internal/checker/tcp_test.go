package checker

import (
	"context"
	"net"
	"testing"
	"time"
)

func TestCheckTCPSuccess(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer func() { _ = ln.Close() }()
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			_ = conn.Close()
		}
	}()

	addr := ln.Addr().(*net.TCPAddr)
	res := CheckTCP(context.Background(), "127.0.0.1", addr.Port, time.Second)

	if !res.Success {
		t.Fatalf("expected success, got error: %s", res.Error)
	}
	if res.TCP == nil || res.TCP.Port != addr.Port {
		t.Errorf("TCP details = %+v", res.TCP)
	}
}

func TestCheckTCPConnectionRefused(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	_ = ln.Close() // free the port immediately so nothing is listening

	res := CheckTCP(context.Background(), "127.0.0.1", port, time.Second)

	if res.Success {
		t.Fatal("expected failure for closed port")
	}
	if res.Error == "" {
		t.Error("expected a non-empty error message")
	}
}

func TestCheckTCPTimeout(t *testing.T) {
	// 192.0.2.0/24 is TEST-NET-1 (RFC 5737): reserved for documentation,
	// so connection attempts there reliably time out rather than being
	// refused or answered.
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	res := CheckTCP(ctx, "192.0.2.1", 81, 50*time.Millisecond)

	if res.Success {
		t.Fatal("expected failure for unreachable host")
	}
}
