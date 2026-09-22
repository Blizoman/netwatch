package cli

import (
	"bytes"
	"errors"
	"net"
	"strconv"
	"strings"
	"testing"
)

func startEchoListener(t *testing.T) (int, func()) {
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
	return ln.Addr().(*net.TCPAddr).Port, func() { _ = ln.Close() }
}

func TestCheckCommandSuccess(t *testing.T) {
	port, cleanup := startEchoListener(t)
	defer cleanup()

	var buf bytes.Buffer
	cmd := NewRootCmd()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"check", "-t", "127.0.0.1:" + strconv.Itoa(port) + "/tcp", "--no-color"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v, output:\n%s", err, buf.String())
	}
	if !strings.Contains(buf.String(), "OK") {
		t.Errorf("expected OK in output, got:\n%s", buf.String())
	}
}

func TestCheckCommandFailureIsSilent(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	_ = ln.Close() // nothing listening now

	var buf bytes.Buffer
	cmd := NewRootCmd()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"check", "-t", "127.0.0.1:" + strconv.Itoa(port) + "/tcp", "--no-color"})

	err = cmd.Execute()
	if !errors.Is(err, ErrChecksFailed) {
		t.Fatalf("expected ErrChecksFailed, got %v", err)
	}
	if strings.Contains(buf.String(), "Error:") {
		t.Errorf("expected no redundant Error: line since the report already shows the failure, got:\n%s", buf.String())
	}
	if !strings.Contains(buf.String(), "FAIL") {
		t.Errorf("expected FAIL in output, got:\n%s", buf.String())
	}
}

func TestCheckCommandNoTargets(t *testing.T) {
	var buf bytes.Buffer
	cmd := NewRootCmd()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"check"})

	if err := cmd.Execute(); err == nil {
		t.Fatal("expected an error when no targets are given")
	}
}
