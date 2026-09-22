// Package target defines the checks netwatch runs against a host.
package target

import (
	"fmt"
	"strconv"
	"strings"
)

// Check names a diagnostic check that can be run against a target.
type Check string

const (
	CheckDNS        Check = "dns"
	CheckTCP        Check = "tcp"
	CheckPing       Check = "ping"
	CheckTraceroute Check = "traceroute"
)

// AllChecks is every check netwatch knows how to run.
var AllChecks = []Check{CheckDNS, CheckTCP, CheckPing, CheckTraceroute}

// Target is a single host to monitor and the checks to run against it.
type Target struct {
	Host   string  `yaml:"host" json:"host"`
	Port   int     `yaml:"port,omitempty" json:"port,omitempty"`
	Checks []Check `yaml:"checks,omitempty" json:"checks,omitempty"`
}

// Name returns a human-friendly identifier for the target.
func (t Target) Name() string {
	if t.Port != 0 {
		return fmt.Sprintf("%s:%d", t.Host, t.Port)
	}
	return t.Host
}

// Validate checks that the target is well-formed and that every requested
// check can actually run (e.g. tcp requires a port).
func (t Target) Validate() error {
	if strings.TrimSpace(t.Host) == "" {
		return fmt.Errorf("target host must not be empty")
	}
	if len(t.Checks) == 0 {
		return fmt.Errorf("target %q must specify at least one check", t.Host)
	}
	for _, c := range t.Checks {
		switch c {
		case CheckDNS, CheckPing, CheckTraceroute:
			// no extra requirements
		case CheckTCP:
			if t.Port <= 0 || t.Port > 65535 {
				return fmt.Errorf("target %q: tcp check requires a valid port", t.Host)
			}
		default:
			return fmt.Errorf("target %q: unknown check %q", t.Host, c)
		}
	}
	return nil
}

// ParseSpec parses a "host", "host:port" or "host:port/check1,check2" spec
// as accepted by the --target CLI flag.
func ParseSpec(spec string, defaultChecks []Check) (Target, error) {
	spec = strings.TrimSpace(spec)
	if spec == "" {
		return Target{}, fmt.Errorf("empty target spec")
	}

	hostPort := spec
	var checks []Check
	if i := strings.LastIndex(spec, "/"); i != -1 {
		hostPort = spec[:i]
		checks = ParseChecks(spec[i+1:])
	}

	host, port := hostPort, 0
	if h, p, err := splitHostPort(hostPort); err == nil {
		host, port = h, p
	}

	if len(checks) == 0 {
		checks = defaultChecks
		// tcp needs an explicit port; drop it from the *implicit* default
		// set rather than rejecting an otherwise-valid "host"-only spec.
		if port == 0 {
			checks = withoutCheck(checks, CheckTCP)
		}
	}

	t := Target{Host: host, Port: port, Checks: checks}
	return t, t.Validate()
}

// splitHostPort splits "host:port" without requiring the brackets
// net.SplitHostPort demands for bare IPv6 addresses.
func splitHostPort(hostPort string) (string, int, error) {
	i := strings.LastIndex(hostPort, ":")
	if i == -1 {
		return "", 0, fmt.Errorf("no port")
	}
	host := hostPort[:i]
	port, err := strconv.Atoi(hostPort[i+1:])
	if err != nil {
		return "", 0, fmt.Errorf("invalid port: %w", err)
	}
	return host, port, nil
}

func withoutCheck(checks []Check, drop Check) []Check {
	out := make([]Check, 0, len(checks))
	for _, c := range checks {
		if c != drop {
			out = append(out, c)
		}
	}
	return out
}

// ParseChecks parses a comma-separated check list, e.g. "dns,tcp,ping".
func ParseChecks(s string) []Check {
	var out []Check
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(strings.ToLower(part))
		if part == "" {
			continue
		}
		out = append(out, Check(part))
	}
	return out
}
