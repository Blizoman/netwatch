package checker

import (
	"context"
	"net"
	"time"
)

// Resolver is the subset of net.Resolver used by CheckDNS, allowing tests
// to inject a fake resolver.
type Resolver interface {
	LookupHost(ctx context.Context, host string) (addrs []string, err error)
}

// CheckDNS resolves host and reports the addresses found along with how
// long resolution took.
func CheckDNS(ctx context.Context, resolver Resolver, host string) Result {
	start := time.Now()
	res := Result{
		Target:    host,
		Type:      TypeDNS,
		Timestamp: start,
	}

	addrs, err := resolver.LookupHost(ctx, host)
	res.Duration = Duration(time.Since(start))
	if err != nil {
		res.Success = false
		res.Error = err.Error()
		return res
	}

	res.Success = true
	res.DNS = &DNSDetails{Addresses: addrs}
	return res
}

// DefaultResolver adapts net.DefaultResolver to the Resolver interface.
func DefaultResolver() Resolver {
	return net.DefaultResolver
}
