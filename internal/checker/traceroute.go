package checker

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

// protocolICMP is the IANA protocol number for ICMPv4, used by
// icmp.ParseMessage to interpret received packets.
const protocolICMP = 1

// TracerouteOptions configures a traceroute check.
type TracerouteOptions struct {
	MaxHops      int           // maximum TTL to probe before giving up
	ProbeTimeout time.Duration // how long to wait for each hop's reply
	ResolveNames bool          // reverse-DNS each responding hop
}

func (o TracerouteOptions) withDefaults() TracerouteOptions {
	if o.MaxHops <= 0 {
		o.MaxHops = 30
	}
	if o.ProbeTimeout <= 0 {
		o.ProbeTimeout = time.Second
	}
	return o
}

// CheckTraceroute discovers the IPv4 path to host, one hop per
// incrementing TTL, using ICMP echo requests.
//
// Unlike CheckPing, traceroute needs to observe ICMP "time exceeded"
// replies from routers along the path, which the Linux/BSD unprivileged
// "ping socket" does not surface. A raw ICMP socket is required, which in
// turn requires root privileges (or the CAP_NET_RAW capability on Linux,
// or Administrator on Windows) -- the same requirement as the standard
// traceroute/tracert tools.
func CheckTraceroute(ctx context.Context, host string, opts TracerouteOptions) Result {
	opts = opts.withDefaults()
	start := time.Now()
	res := Result{
		Target:    host,
		Type:      TypeTraceroute,
		Timestamp: start,
	}

	dst, err := net.ResolveIPAddr("ip4", host)
	if err != nil {
		res.Duration = Duration(time.Since(start))
		res.Error = fmt.Sprintf("resolve: %v", err)
		return res
	}

	hops, reached, err := traceroute(ctx, dst, opts)
	res.Duration = Duration(time.Since(start))
	res.Traceroute = &TracerouteDetails{Hops: hops, Reached: reached}
	if err != nil {
		res.Error = err.Error()
		return res
	}
	res.Success = reached
	if !reached {
		res.Error = "destination not reached within max hops"
	}
	return res
}

func traceroute(ctx context.Context, dst *net.IPAddr, opts TracerouteOptions) ([]Hop, bool, error) {
	conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		if isPermissionErr(err) {
			return nil, false, fmt.Errorf("traceroute requires elevated privileges: %w (try running as root, granting CAP_NET_RAW with 'setcap cap_net_raw+ep' on Linux, or as Administrator on Windows)", err)
		}
		return nil, false, err
	}
	defer func() { _ = conn.Close() }()

	p4 := conn.IPv4PacketConn()
	id := os.Getpid() & 0xffff
	hops := make([]Hop, 0, opts.MaxHops)
	reached := false

	for ttl := 1; ttl <= opts.MaxHops; ttl++ {
		select {
		case <-ctx.Done():
			return hops, reached, ctx.Err()
		default:
		}

		hop, ok, err := probeHop(conn, p4, dst, id, ttl, opts.ProbeTimeout)
		if err != nil {
			return hops, reached, err
		}
		if opts.ResolveNames && hop.Address != "" {
			hop.Hostname = reverseLookup(hop.Address)
		}
		hops = append(hops, hop)
		if ok {
			reached = true
			break
		}
	}

	return hops, reached, nil
}

func probeHop(conn *icmp.PacketConn, p4 *ipv4.PacketConn, dst *net.IPAddr, id, ttl int, timeout time.Duration) (Hop, bool, error) {
	hop := Hop{Number: ttl, TimedOut: true}

	if err := p4.SetTTL(ttl); err != nil {
		return hop, false, fmt.Errorf("set ttl: %w", err)
	}

	msg := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &icmp.Echo{
			ID:   id,
			Seq:  ttl,
			Data: []byte("netwatch"),
		},
	}
	wb, err := msg.Marshal(nil)
	if err != nil {
		return hop, false, fmt.Errorf("marshal probe: %w", err)
	}

	sentAt := time.Now()
	if _, err := conn.WriteTo(wb, dst); err != nil {
		return hop, false, fmt.Errorf("send probe: %w", err)
	}
	if err := conn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
		return hop, false, fmt.Errorf("set read deadline: %w", err)
	}

	rb := make([]byte, 1500)
	for {
		n, peer, err := conn.ReadFrom(rb)
		if err != nil {
			var nerr net.Error
			if errors.As(err, &nerr) && nerr.Timeout() {
				return hop, false, nil
			}
			return hop, false, fmt.Errorf("read reply: %w", err)
		}
		rtt := time.Since(sentAt)

		rm, err := icmp.ParseMessage(protocolICMP, rb[:n])
		if err != nil {
			continue
		}

		peerAddr := peer.String()
		switch rm.Type {
		case ipv4.ICMPTypeTimeExceeded:
			hop = Hop{Number: ttl, Address: peerAddr, RTT: Duration(rtt)}
			return hop, false, nil
		case ipv4.ICMPTypeEchoReply:
			echo, ok := rm.Body.(*icmp.Echo)
			if !ok || echo.ID != id {
				continue
			}
			hop = Hop{Number: ttl, Address: peerAddr, RTT: Duration(rtt)}
			return hop, peerAddr == dst.IP.String(), nil
		default:
			continue
		}
	}
}

func reverseLookup(ip string) string {
	names, err := net.LookupAddr(ip)
	if err != nil || len(names) == 0 {
		return ""
	}
	return strings.TrimSuffix(names[0], ".")
}

func isPermissionErr(err error) bool {
	return errors.Is(err, os.ErrPermission)
}
