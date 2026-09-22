package target

import (
	"reflect"
	"testing"
)

func TestParseSpec(t *testing.T) {
	defaultChecks := []Check{CheckDNS, CheckTCP, CheckPing}

	tests := []struct {
		name    string
		spec    string
		want    Target
		wantErr bool
	}{
		{
			name: "host only drops tcp from the implicit defaults (no port given)",
			spec: "example.com",
			want: Target{Host: "example.com", Checks: []Check{CheckDNS, CheckPing}},
		},
		{
			name: "host and port",
			spec: "example.com:443",
			want: Target{Host: "example.com", Port: 443, Checks: defaultChecks},
		},
		{
			name: "host with explicit checks",
			spec: "example.com/dns,ping",
			want: Target{Host: "example.com", Checks: []Check{CheckDNS, CheckPing}},
		},
		{
			name: "host, port and explicit checks",
			spec: "example.com:443/tcp",
			want: Target{Host: "example.com", Port: 443, Checks: []Check{CheckTCP}},
		},
		{
			name:    "empty spec",
			spec:    "",
			wantErr: true,
		},
		{
			name:    "tcp check without a port",
			spec:    "example.com/tcp",
			wantErr: true,
		},
		{
			name:    "unknown check",
			spec:    "example.com/bogus",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseSpec(tt.spec, defaultChecks)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseSpec(%q) error = %v, wantErr %v", tt.spec, err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseSpec(%q) = %+v, want %+v", tt.spec, got, tt.want)
			}
		})
	}
}

func TestTargetName(t *testing.T) {
	if got := (Target{Host: "example.com"}).Name(); got != "example.com" {
		t.Errorf("Name() = %q, want %q", got, "example.com")
	}
	if got := (Target{Host: "example.com", Port: 443}).Name(); got != "example.com:443" {
		t.Errorf("Name() = %q, want %q", got, "example.com:443")
	}
}

func TestTargetValidate(t *testing.T) {
	cases := []struct {
		name    string
		target  Target
		wantErr bool
	}{
		{"empty host", Target{Checks: []Check{CheckDNS}}, true},
		{"no checks", Target{Host: "h"}, true},
		{"valid dns", Target{Host: "h", Checks: []Check{CheckDNS}}, false},
		{"tcp missing port", Target{Host: "h", Checks: []Check{CheckTCP}}, true},
		{"tcp with port", Target{Host: "h", Port: 80, Checks: []Check{CheckTCP}}, false},
		{"tcp port out of range", Target{Host: "h", Port: 70000, Checks: []Check{CheckTCP}}, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := c.target.Validate()
			if (err != nil) != c.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, c.wantErr)
			}
		})
	}
}
