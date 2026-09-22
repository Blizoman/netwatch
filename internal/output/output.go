// Package output renders check results as human-readable or JSON text.
package output

import (
	"io"
	"time"

	"github.com/Blizoman/netwatch/internal/checker"
)

// Round bundles one pass of results with the time it was produced, so
// continuous "watch" mode output is self-describing.
type Round struct {
	Timestamp time.Time        `json:"timestamp"`
	Results   []checker.Result `json:"results"`
}

// Writer renders one Round of results to its underlying stream.
type Writer interface {
	Write(round Round) error
}

// New returns a Writer for the given format ("human" or "json"). noColor
// forces ANSI color off regardless of terminal detection; when false,
// color is still auto-disabled for non-terminal output or NO_COLOR (see
// github.com/fatih/color's own detection).
func New(w io.Writer, format string, noColor bool) Writer {
	if format == "json" {
		return &jsonWriter{w: w}
	}
	return &humanWriter{w: w, noColor: noColor}
}

// Summarize counts how many results in a round succeeded and failed.
func Summarize(results []checker.Result) (ok, failed int) {
	for _, r := range results {
		if r.Success {
			ok++
		} else {
			failed++
		}
	}
	return ok, failed
}
