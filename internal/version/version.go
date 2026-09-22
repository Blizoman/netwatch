// Package version holds build-time version metadata, injected via
// -ldflags at release build time (see Makefile / .goreleaser.yaml).
package version

import (
	"fmt"
	"runtime"
)

var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

// String renders a one-line version banner.
func String() string {
	return fmt.Sprintf("netwatch %s (commit %s, built %s, %s/%s, %s)",
		Version, Commit, Date, runtime.GOOS, runtime.GOARCH, runtime.Version())
}
