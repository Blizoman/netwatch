// Command netwatch is a cross-platform CLI for network diagnostics:
// DNS resolution, TCP connectivity, ICMP latency/packet loss and
// traceroute, in one-shot or continuous monitoring modes.
package main

import (
	"os"

	"github.com/Blizoman/netwatch/internal/cli"
)

func main() {
	if err := cli.NewRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
