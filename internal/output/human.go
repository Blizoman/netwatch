package output

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/fatih/color"

	"github.com/Blizoman/netwatch/internal/checker"
)

// humanWriter renders a round of results as a colored, aligned table.
type humanWriter struct {
	w       io.Writer
	noColor bool
}

func (h *humanWriter) Write(round Round) error {
	ok := color.New(color.FgGreen, color.Bold)
	fail := color.New(color.FgRed, color.Bold)
	dim := color.New(color.FgHiBlack)
	if h.noColor {
		ok.DisableColor()
		fail.DisableColor()
		dim.DisableColor()
	}

	if _, err := fmt.Fprintf(h.w, "%s\n", dim.Sprintf("── %s ──", round.Timestamp.Format(time.RFC3339))); err != nil {
		return err
	}

	tw := tabwriter.NewWriter(h.w, 0, 2, 2, ' ', 0)
	for _, r := range round.Results {
		status := ok.Sprint("OK")
		if !r.Success {
			status = fail.Sprint("FAIL")
		}
		if _, err := fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n", status, r.Type, r.Target, time.Duration(r.Duration).Round(time.Millisecond), detailLine(r)); err != nil {
			return err
		}
	}
	if err := tw.Flush(); err != nil {
		return err
	}

	for _, r := range round.Results {
		if r.Type == checker.TypeTraceroute && r.Traceroute != nil {
			if err := writeHops(h.w, dim, r); err != nil {
				return err
			}
		}
	}

	okCount, failCount := Summarize(round.Results)
	summary := fmt.Sprintf("%d ok, %d failed (%d total)", okCount, failCount, len(round.Results))
	if failCount > 0 {
		summary = fail.Sprint(summary)
	} else {
		summary = ok.Sprint(summary)
	}
	_, err := fmt.Fprintf(h.w, "%s\n\n", summary)
	return err
}

func writeHops(w io.Writer, dim *color.Color, r checker.Result) error {
	if _, err := fmt.Fprintf(w, "  %s\n", dim.Sprintf("traceroute %s:", r.Target)); err != nil {
		return err
	}
	tw := tabwriter.NewWriter(w, 0, 2, 2, ' ', 0)
	for _, hop := range r.Traceroute.Hops {
		addr := hop.Address
		if addr == "" {
			addr = "*"
		}
		rtt := "*"
		if !hop.TimedOut {
			rtt = time.Duration(hop.RTT).Round(time.Millisecond).String()
		}
		name := hop.Hostname
		if _, err := fmt.Fprintf(tw, "    %d\t%s\t%s\t%s\n", hop.Number, addr, name, rtt); err != nil {
			return err
		}
	}
	return tw.Flush()
}

func detailLine(r checker.Result) string {
	if !r.Success && r.Error != "" {
		return r.Error
	}
	switch r.Type {
	case checker.TypeDNS:
		if r.DNS != nil {
			return strings.Join(r.DNS.Addresses, ", ")
		}
	case checker.TypeTCP:
		if r.TCP != nil {
			return fmt.Sprintf("connected to %s", r.TCP.RemoteAddr)
		}
	case checker.TypePing:
		if r.Ping != nil {
			mode := "unprivileged"
			if r.Ping.Privileged {
				mode = "privileged"
			}
			return fmt.Sprintf("loss=%.1f%% min=%s avg=%s max=%s (%s)",
				r.Ping.PacketLossPercent,
				time.Duration(r.Ping.MinRTT).Round(time.Millisecond),
				time.Duration(r.Ping.AvgRTT).Round(time.Millisecond),
				time.Duration(r.Ping.MaxRTT).Round(time.Millisecond),
				mode,
			)
		}
	case checker.TypeTraceroute:
		if r.Traceroute != nil {
			return fmt.Sprintf("%d hops, reached=%t", len(r.Traceroute.Hops), r.Traceroute.Reached)
		}
	}
	return ""
}
