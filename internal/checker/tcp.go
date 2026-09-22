package checker

import (
	"context"
	"fmt"
	"net"
	"time"
)

// CheckTCP attempts a TCP connection to host:port and reports whether the
// handshake succeeded and how long it took.
func CheckTCP(ctx context.Context, host string, port int, timeout time.Duration) Result {
	start := time.Now()
	res := Result{
		Target:    fmt.Sprintf("%s:%d", host, port),
		Type:      TypeTCP,
		Timestamp: start,
	}

	dialCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var d net.Dialer
	conn, err := d.DialContext(dialCtx, "tcp", net.JoinHostPort(host, fmt.Sprintf("%d", port)))
	res.Duration = Duration(time.Since(start))
	if err != nil {
		res.Success = false
		res.Error = err.Error()
		res.TCP = &TCPDetails{Port: port}
		return res
	}
	defer func() { _ = conn.Close() }()

	res.Success = true
	res.TCP = &TCPDetails{Port: port, RemoteAddr: conn.RemoteAddr().String()}
	return res
}
