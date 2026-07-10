// Package terraform holds tool-layer CompositeError parsers for terraform jobs.
// They register at errparse.LayerTool so a provider-specific cause (e.g. an AWS
// IAM denial parsed at LayerProvider) still wins, but any other terraform
// diagnostic yields a clean, structured error instead of falling through to the
// raw generic dump. This is the first tool-layer parser; more specific
// terraform causes (state lock, backend init, ...) can register alongside it.
package terraform

import (
	"fmt"
	"strings"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/errparse"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
)

// TerraformErrorType is the discriminator for a terraform diagnostic that no
// provider-layer parser recognised.
const TerraformErrorType compositeerrors.Type = "terraform.error"

const (
	// maxHeadline bounds the one-line message; full detail lives in the output.
	maxHeadline = 240
	// maxBody bounds the stored output so a pathological log can't bloat the
	// JSONB payload.
	maxBody = 8000
	// maxErrors caps how many distinct summaries we keep for the list section.
	maxErrors = 20
)

// TerraformError is the tool-layer payload: the terraform "Error:" summaries
// extracted from the diagnostics, plus the cleaned output for context.
type TerraformError struct {
	// Summary is the first terraform error summary (the text after "Error:").
	Summary string `json:"summary"`
	// Errors lists every distinct error summary when terraform reported more
	// than one. Empty when there is only the single Summary.
	Errors []string `json:"errors,omitempty"`
	// Output is the cleaned, possibly-truncated diagnostic output.
	Output string `json:"output,omitempty"`
}

var _ compositeerrors.CompositeError = (*TerraformError)(nil)

// Error returns the first error summary as the headline, noting the count when
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

// Sections lists every summary when there are several (the headline shows only
// the first), then the cleaned diagnostic output for full context.
func (e *TerraformError) Sections() []compositeerrors.Section {
	var sections []compositeerrors.Section

	if len(e.Errors) > 1 {
		var b strings.Builder
		for _, s := range e.Errors {
			b.WriteString("- ")
			b.WriteString(s)
			b.WriteString("\n")
		}
		sections = append(sections, compositeerrors.Section{
			Heading: "Errors",
			Body:    strings.TrimRight(b.String(), "\n"),
		})
	}

	if e.Output != "" {
		sections = append(sections, compositeerrors.Section{
			Heading: "Output",
			Body:    "```\n" + e.Output + "\n```",
		})
	}

	return sections
}

// errorParser recognises terraform "Error:" diagnostics in a terraform job's
// raw output.
type errorParser struct{}

func (errorParser) Layer() errparse.Layer  { return errparse.LayerTool }
func (errorParser) Tools() []errparse.Tool { return []errparse.Tool{errparse.ToolTerraform} }

// Signals gates on the presence of a terraform diagnostic. "Error:" is broad,
// but the parser is already bucketed to terraform jobs, where its presence
// reliably marks a diagnostic block.
func (errorParser) Signals() []string                      { return []string{"Error:"} }
func (errorParser) Applicable(*errparse.ParseContext) bool { return true }

func (errorParser) Parse(ctx *errparse.ParseContext) compositeerrors.CompositeError {
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
	errparse.Register(errorParser{})
}

// errorSummaries returns the distinct text following each terraform "Error:"
// diagnostic line, in order of first appearance.
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

// cleanedLines returns the non-blank lines of raw, each trimmed of surrounding
// space and terraform's "│" box-drawing prefix.
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

// truncate caps s to n runes (not bytes) so it never splits a multi-byte rune
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
