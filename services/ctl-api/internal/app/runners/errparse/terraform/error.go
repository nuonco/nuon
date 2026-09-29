package terraform

import (
	"fmt"
	"strings"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/errparse"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
)

const TerraformErrorType compositeerrors.Type = "terraform.error"

const (
	maxHeadline = 240
	maxBody     = 8000
	maxErrors   = 20
)

type TerraformError struct {
	Summary string   `json:"summary"`
	Errors  []string `json:"errors,omitempty"`
	Output  string   `json:"output,omitempty"`
}

var _ compositeerrors.CompositeError = (*TerraformError)(nil)

// why: Error returns the first error summary as the headline, noting the count when
// terraform reported several. The summary is truncated first so the "+N more"
// suffix is never cut off.
func (e *TerraformError) Error() string {
	h := truncate(e.Summary, maxHeadline)
	if n := len(e.Errors); n > 1 {
		h = fmt.Sprintf("%s (+%d more errors)", h, n-1)
	}
	return h
}

func (e *TerraformError) Type() compositeerrors.Type { return TerraformErrorType }
func (e *TerraformError) Severity() compositeerrors.Severity {
	return compositeerrors.SeverityError
}

func (e *TerraformError) Sections() []compositeerrors.Section {
	var sections []compositeerrors.Section

	if len(e.Errors) > 1 {
		sections = append(sections, compositeerrors.TextSection("Errors", strings.Join(e.Errors, "\n")))
	}

	if e.Output != "" {
		sections = append(sections, compositeerrors.CodeSection("Output", e.Output))
	}

	return sections
}

func parseError(ctx *errparse.ParseContext) compositeerrors.CompositeError {
	lines := cleanedLines(ctx.Raw)
	summaries := errorSummaries(lines)
	if len(summaries) == 0 {
		return nil
	}

	output := truncate(strings.Join(lines, "\n"), maxBody)

	e := &TerraformError{
		Summary: summaries[0],
		Output:  output,
	}
	if len(summaries) > 1 {
		e.Errors = summaries
	}
	return e
}

func init() {
	errparse.Register(errparse.NewParser(errparse.LayerTool, parseError,
		errparse.WithTools(errparse.ToolTerraform),
		errparse.WithSignals("Error:"),
	))
}

func errorSummaries(lines []string) []string {
	var out []string
	seen := map[string]bool{}
	for _, l := range lines {
		s, ok := strings.CutPrefix(l, "Error:")
		if !ok {
			continue
		}
		s = strings.TrimSpace(s)
		if s == "" || seen[s] {
			continue
		}
		seen[s] = true
		out = append(out, s)
		if len(out) >= maxErrors {
			break
		}
	}
	return out
}

func cleanedLines(raw string) []string {
	var out []string
	for _, line := range strings.Split(raw, "\n") {
		t := strings.TrimSpace(line)
		t = strings.TrimPrefix(t, "│")
		t = strings.TrimSpace(t)
		if t == "" {
			continue
		}
		out = append(out, t)
	}
	return out
}

// why: truncate caps s to n runes (not bytes) so it never splits a multi-byte rune
// into invalid UTF-8, appending an ellipsis when it cuts.
func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n]) + "…"
}
