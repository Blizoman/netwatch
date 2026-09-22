// Package checker implements the individual network diagnostic checks
// (DNS resolution, TCP connect, ICMP ping, traceroute) used by netwatch.
package checker

import (
	"encoding/json"
	"time"
)

// Type identifies which diagnostic check produced a Result.
type Type string

const (
	TypeDNS        Type = "dns"
	TypeTCP        Type = "tcp"
	TypePing       Type = "ping"
	TypeTraceroute Type = "traceroute"
)

// Duration wraps time.Duration to marshal as milliseconds in JSON output
// while keeping a human string form for logs.
type Duration time.Duration

func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(d).Seconds() * 1000)
}

func (d Duration) String() string {
	return time.Duration(d).String()
}

// Result is the outcome of a single check against a single target.
type Result struct {
	Target    string    `json:"target"`
	Type      Type      `json:"type"`
	Success   bool      `json:"success"`
	Timestamp time.Time `json:"timestamp"`
	Duration  Duration  `json:"duration_ms"`
	Error     string    `json:"error,omitempty"`

	DNS        *DNSDetails        `json:"dns,omitempty"`
	TCP        *TCPDetails        `json:"tcp,omitempty"`
	Ping       *PingDetails       `json:"ping,omitempty"`
	Traceroute *TracerouteDetails `json:"traceroute,omitempty"`
}

// DNSDetails carries the outcome of a DNS resolution check.
type DNSDetails struct {
	Addresses []string `json:"addresses"`
}

// TCPDetails carries the outcome of a TCP connect check.
type TCPDetails struct {
	Port       int    `json:"port"`
	RemoteAddr string `json:"remote_addr,omitempty"`
}

// PingDetails carries latency and packet-loss statistics from an ICMP
// echo check.
type PingDetails struct {
	PacketsSent       int      `json:"packets_sent"`
	PacketsRecv       int      `json:"packets_recv"`
	PacketLossPercent float64  `json:"packet_loss_percent"`
	MinRTT            Duration `json:"min_rtt_ms"`
	AvgRTT            Duration `json:"avg_rtt_ms"`
	MaxRTT            Duration `json:"max_rtt_ms"`
	StdDevRTT         Duration `json:"stddev_rtt_ms"`
	Privileged        bool     `json:"privileged"`
}

// Hop is a single hop reported by a traceroute check.
type Hop struct {
	Number   int      `json:"number"`
	Address  string   `json:"address,omitempty"`
	Hostname string   `json:"hostname,omitempty"`
	RTT      Duration `json:"rtt_ms,omitempty"`
	TimedOut bool     `json:"timed_out"`
}

// TracerouteDetails carries the hop-by-hop path discovered to a target.
type TracerouteDetails struct {
	Hops    []Hop `json:"hops"`
	Reached bool  `json:"reached"`
}
