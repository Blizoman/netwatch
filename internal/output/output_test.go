package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/Blizoman/netwatch/internal/checker"
)

func sampleRound() Round {
	return Round{
		Timestamp: time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC),
		Results: []checker.Result{
			{
				Target:  "example.com",
				Type:    checker.TypeDNS,
				Success: true,
				DNS:     &checker.DNSDetails{Addresses: []string{"1.2.3.4"}},
			},
			{
				Target:  "example.com:443",
				Type:    checker.TypeTCP,
				Success: false,
				Error:   "connection refused",
			},
		},
	}
}

func TestJSONWriter(t *testing.T) {
	var buf bytes.Buffer
	w := New(&buf, "json", true)

	if err := w.Write(sampleRound()); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	var decoded Round
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, buf.String())
	}
	if len(decoded.Results) != 2 {
		t.Fatalf("len(Results) = %d, want 2", len(decoded.Results))
	}
	if decoded.Results[0].DNS == nil || decoded.Results[0].DNS.Addresses[0] != "1.2.3.4" {
		t.Errorf("decoded DNS details = %+v", decoded.Results[0].DNS)
	}
}

func TestHumanWriterNoColor(t *testing.T) {
	var buf bytes.Buffer
	w := New(&buf, "human", true)

	if err := w.Write(sampleRound()); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	out := buf.String()
	if strings.Contains(out, "\x1b[") {
		t.Errorf("expected no ANSI escape codes with noColor=true, got:\n%s", out)
	}
	if !strings.Contains(out, "OK") || !strings.Contains(out, "FAIL") {
		t.Errorf("expected both OK and FAIL statuses, got:\n%s", out)
	}
	if !strings.Contains(out, "connection refused") {
		t.Errorf("expected error detail in output, got:\n%s", out)
	}
	if !strings.Contains(out, "1 ok, 1 failed") {
		t.Errorf("expected summary line, got:\n%s", out)
	}
}

func TestSummarize(t *testing.T) {
	ok, failed := Summarize(sampleRound().Results)
	if ok != 1 || failed != 1 {
		t.Errorf("Summarize() = (%d, %d), want (1, 1)", ok, failed)
	}
}
