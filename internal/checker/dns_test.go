package checker

import (
	"context"
	"errors"
	"testing"
)

type fakeResolver struct {
	addrs []string
	err   error
}

func (f fakeResolver) LookupHost(ctx context.Context, host string) ([]string, error) {
	return f.addrs, f.err
}

func TestCheckDNSSuccess(t *testing.T) {
	r := fakeResolver{addrs: []string{"93.184.216.34"}}
	res := CheckDNS(context.Background(), r, "example.com")

	if !res.Success {
		t.Fatalf("expected success, got error: %s", res.Error)
	}
	if res.Type != TypeDNS {
		t.Errorf("Type = %v, want %v", res.Type, TypeDNS)
	}
	if res.DNS == nil || len(res.DNS.Addresses) != 1 || res.DNS.Addresses[0] != "93.184.216.34" {
		t.Errorf("DNS details = %+v", res.DNS)
	}
}

func TestCheckDNSFailure(t *testing.T) {
	r := fakeResolver{err: errors.New("no such host")}
	res := CheckDNS(context.Background(), r, "invalid.invalid")

	if res.Success {
		t.Fatal("expected failure")
	}
	if res.Error != "no such host" {
		t.Errorf("Error = %q", res.Error)
	}
	if res.DNS != nil {
		t.Errorf("expected nil DNS details on failure, got %+v", res.DNS)
	}
}
