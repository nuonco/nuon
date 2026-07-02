// Package generic holds the tool-agnostic, always-matching fallback parser. It
// registers at errparse.LayerGeneric so it only wins when no provider- or
// tool-specific parser recognised the failure: every failed runner job then
// still produces a CompositeError carrying the (cleaned) error output, rather
// than surfacing nothing.
package generic

import (
	"strings"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/errparse"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
)

// GenericErrorType is the discriminator for an unclassified job failure. The
// dashboard renders it with the same generic renderer as every other composite
// error; its lack of a specific type is what marks it as "not yet categorised".
const GenericErrorType compositeerrors.Type = "generic"

const (
	// maxHeadline bounds the one-line message. Full detail lives in the section.
	maxHeadline = 240
	// maxBody bounds the stored error output so a pathological log can't bloat
	// the JSONB payload. Kept generous; the root cause is usually well within.
	maxBody = 8000
)

// GenericError is the fallback CompositeError: a cleaned copy of the raw error
// output with a best-effort headline. Body is the typed payload (what a future
// view would group/inspect); the rendered detail section mirrors it.
type GenericError struct {
	// Body is the cleaned, possibly-truncated error output.
	Body string `json:"body"`
}

var _ compositeerrors.CompositeError = (*GenericError)(nil)

// Error returns a best-effort one-line headline extracted from the body.
func (e *GenericError) Error() string {
	if h := headline(e.Body); h != "" {
		return h
	}
	return "Job failed"
}

func (e *GenericError) Type() compositeerrors.Type { return GenericErrorType }
func (e *GenericError) Severity() compositeerrors.Severity {
	return compositeerrors.SeverityError
}

// Sections renders the full cleaned error output in a code block.
func (e *GenericError) Sections() []compositeerrors.Section {
	if e.Body == "" {
		return nil
	}
	return []compositeerrors.Section{
		{
			Heading: "Error output",
			Body:    "```\n" + e.Body + "\n```",
		},
	}
}

// genericParser is the tool-agnostic, always-candidate fallback. It has no
// signals (always a candidate) and no tools (considered for every job), and
// sits at LayerGeneric so specific parsers always win.
type genericParser struct{}

func (genericParser) Layer() errparse.Layer                  { return errparse.LayerGeneric }
func (genericParser) Tools() []errparse.Tool                 { return nil }
func (genericParser) Signals() []string                      { return nil }
func (genericParser) Applicable(*errparse.ParseContext) bool { return true }

func (genericParser) Parse(ctx *errparse.ParseContext) compositeerrors.CompositeError {
	body := cleanBody(ctx.Raw)
	if body == "" {
		return nil
	}
	return &GenericError{Body: body}
}

func init() {
	errparse.Register(genericParser{})
}

// cleanBody normalises raw error output for display: it strips the terraform
// "│ " box-drawing prefix, drops blank lines, and bounds the total length.
func cleanBody(raw string) string {
	lines := cleanedLines(raw)
	if len(lines) == 0 {
		return ""
	}
	body := strings.Join(lines, "\n")
	return truncate(body, maxBody)
}

// headline picks the most informative single line: the first "Error:" line when
// present (terraform/helm style), otherwise the first cleaned line.
func headline(body string) string {
	lines := strings.Split(body, "\n")
	for _, l := range lines {
		if strings.HasPrefix(l, "Error:") {
			return truncate(l, maxHeadline)
		}
	}
	for _, l := range lines {
		if l != "" {
			return truncate(l, maxHeadline)
		}
	}
	return ""
}

// cleanedLines returns the non-blank lines of raw, each trimmed of surrounding
// space and the terraform box-drawing prefix.
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
	if len(s) <= n { // fast path: byte length <= n implies rune length <= n
		return s
	}
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n]) + "…"
}
