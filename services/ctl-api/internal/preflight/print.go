package preflight

import (
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"
)

type palette struct {
	pass, warn, fail, skip, dim, reset string
}

func paletteFor(w io.Writer) palette {
	if os.Getenv("NO_COLOR") != "" {
		return palette{}
	}

	f, ok := w.(*os.File)
	if !ok || !term.IsTerminal(int(f.Fd())) {
		return palette{}
	}

	return palette{
		pass:  "\033[32m",
		warn:  "\033[33m",
		fail:  "\033[31m",
		skip:  "\033[90m",
		dim:   "\033[90m",
		reset: "\033[0m",
	}
}

func (p palette) status(s Status) string {
	switch s {
	case StatusPass:
		return p.pass + "✓ pass" + p.reset
	case StatusWarn:
		return p.warn + "! warn" + p.reset
	case StatusFail:
		return p.fail + "✗ fail" + p.reset
	default:
		return p.skip + "- skip" + p.reset
	}
}

func PrintResults(w io.Writer, results []Result) int {
	p := paletteFor(w)
	s := summarize(results)
	width := nameWidth(results)

	fmt.Fprintln(w)
	fmt.Fprintf(w, "  %-*s  %-6s  %s\n", width, "CHECK", "STATUS", "DETAIL")
	for _, r := range results {
		fmt.Fprintf(w, "  %-*s  %s  %s\n", width, r.Name, p.status(r.Status), r.Detail)

		if r.Failed() {
			for _, f := range r.Fields {
				fmt.Fprintf(w, "  %-*s  %s%s%s\n", width, "", p.dim, fieldLine(f), p.reset)
			}
		}
	}

	fmt.Fprintln(w)
	fmt.Fprintf(w, "  %d passed, %d warned, %d failed, %d skipped\n",
		s.Passed, s.Warned, s.Failed, s.Skipped)

	if s.Failed > 0 {
		return 1
	}

	return 0
}

func PrintChecks(w io.Writer, results []Result) {
	p := paletteFor(w)

	for i, r := range results {
		if i > 0 {
			fmt.Fprintln(w)
		}

		if _, ok := Lookup(r.Name); !ok {
			fmt.Fprintf(w, "  %s%s%s%s — %s%s\n", p.dim, r.Name, p.reset, p.fail, r.Detail, p.reset)
			continue
		}

		fmt.Fprintf(w, "  %s%s%s — %s\n", p.dim, r.Name, p.reset, r.Description)
		if r.Status == StatusSkipped {
			fmt.Fprintf(w, "    %sskipped: %s%s\n", p.skip, r.Detail, p.reset)
		}

		for _, f := range r.Fields {
			fmt.Fprintf(w, "    %s\n", fieldLine(f))
		}
	}
}

func fieldLine(f Field) string {
	req := "optional"
	if f.Required {
		req = "required"
	}

	return fmt.Sprintf("%-32s %-8s %s", f.Name, req, f.Display())
}

func nameWidth(results []Result) int {
	width := len("CHECK")
	for _, r := range results {
		if len(r.Name) > width {
			width = len(r.Name)
		}
	}

	return width
}

func truncate(s string, max int) string {
	s = strings.TrimSpace(s)
	if len(s) <= max {
		return s
	}

	return s[:max-3] + "..."
}
