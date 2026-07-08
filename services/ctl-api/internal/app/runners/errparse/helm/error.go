// Package helm holds tool-layer CompositeError parsers for helm jobs. They
// register at errparse.LayerTool so a provider-specific cause (e.g. an AWS IAM
// denial parsed at LayerProvider) still wins, but any recognised helm failure
// yields a clean, structured error instead of falling through to the raw
// generic dump.
//
// The runner drives helm through the Go SDK (helm.sh/helm/v4 pkg/action), not
// the CLI, so the familiar "INSTALLATION FAILED"/"UPGRADE FAILED" phase
// prefixes never appear — those are added only by helm's cmd layer. What is
// reliably present is the runner's own wrapper around the SDK error
// ("unable to upgrade helm release: <sdk error>", "unable to execute with
// dry-run: <sdk error>"). The parser leads the headline at the SDK error by
// stripping that wrapper, and falls back to a set of verified helm v4 SDK cause
// strings when the wrapper is absent from the captured output.
package helm

import (
	"strings"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/errparse"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
)

// HelmErrorType is the discriminator for a helm failure that no provider-layer
// parser recognised.
const HelmErrorType compositeerrors.Type = "helm.error"

const (
	// maxHeadline bounds the one-line message; full detail lives in the output.
	maxHeadline = 240
	// maxBody bounds the stored output so a pathological log can't bloat the
	// JSONB payload.
	maxBody = 8000
)

// wrappers are the runner's own error-wrap prefixes around the helm SDK error.
// They are the highest-confidence anchor: the runner always wraps a failed helm
// action with one of these, so the text following the wrapper is the real SDK
// cause with the runner's nesting stripped. See bins/runner/internal/jobs/
// deploy/helm (operation_install.go / operation_upgrade.go).
var wrappers = []string{"helm release:", "with dry-run:"}

// causes are verified helm v4 SDK (pkg/action) error substrings, used as a
// backup anchor when the captured output does not carry the runner wrapper
// (e.g. only a log line was retained). Every entry is helm-specific — generic
// kubernetes phrases like "timed out waiting for the condition" are
// deliberately excluded, since they can appear in streamed pod logs before the
// real cause and would produce a misleading headline.
var causes = []string{
	"cannot reuse a name that is still in use",
	"exists and cannot be imported into the current release",
	"invalid ownership metadata",
	"unable to build kubernetes objects from",
	"failed pre-install",
	"failed post-install",
	"pre-upgrade hooks failed",
	"post-upgrade hooks failed",
	"failed to install CRD",
	"has no deployed releases",
	"another operation (install/upgrade/rollback) is in progress",
	"chart dependencies processing failed",
}

// HelmError is the tool-layer payload: the helm failure summary (the SDK error
// with the runner's wrapper stripped) plus the cleaned output for context.
type HelmError struct {
	// Summary is the helm failure cause, with the runner's wrapper prefix
	// removed so it leads with the SDK error rather than the nesting.
	Summary string `json:"summary"`
	// Output is the cleaned, possibly-truncated diagnostic output.
	Output string `json:"output,omitempty"`
}

var _ compositeerrors.CompositeError = (*HelmError)(nil)

func (e *HelmError) Error() string                      { return truncate(e.Summary, maxHeadline) }
func (e *HelmError) Type() compositeerrors.Type         { return HelmErrorType }
func (e *HelmError) Severity() compositeerrors.Severity { return compositeerrors.SeverityError }

func (e *HelmError) Sections() []compositeerrors.Section {
	if e.Output == "" {
		return nil
	}
	return []compositeerrors.Section{{
		Heading: "Output",
		Body:    "```\n" + e.Output + "\n```",
	}}
}

// errorParser recognises helm failures in a helm job's raw output.
type errorParser struct{}

func (errorParser) Layer() errparse.Layer  { return errparse.LayerTool }
func (errorParser) Tools() []errparse.Tool { return []errparse.Tool{errparse.ToolHelm} }

func (errorParser) Signals() []string {
	return append(append([]string{}, wrappers...), causes...)
}

func (errorParser) Applicable(*errparse.ParseContext) bool { return true }

func (errorParser) Parse(ctx *errparse.ParseContext) compositeerrors.CompositeError {
	lines := cleanedLines(ctx.Raw)

	summary := wrapperSummary(lines)
	if summary == "" {
		summary = causeSummary(lines)
	}
	if summary == "" {
		return nil
	}

	return &HelmError{
		Summary: summary,
		Output:  truncate(strings.Join(lines, "\n"), maxBody),
	}
}

func init() {
	errparse.Register(errorParser{})
}

// wrapperSummary returns the SDK error that follows the runner's helm wrapper on
// the first line that carries one, or "" when no wrapper is present. When a line
// nests several wrappers the rightmost one is used, so the deepest (real) cause
// leads. The wrapper phrase itself is only ever on the actual error line, never
// on streamed pod-log lines, which is why scanning from the first match is safe.
func wrapperSummary(lines []string) string {
	for _, l := range lines {
		cut := -1
		for _, w := range wrappers {
			if i := strings.LastIndex(l, w); i >= 0 {
				if end := i + len(w); end > cut {
					cut = end
				}
			}
		}
		if cut >= 0 {
			if s := strings.TrimSpace(l[cut:]); s != "" {
				return s
			}
		}
	}
	return ""
}

// causeSummary returns the text from the earliest verified helm cause marker on
// the first line that carries one, or "" when none is present.
func causeSummary(lines []string) string {
	for _, l := range lines {
		best := -1
		for _, c := range causes {
			if i := strings.Index(l, c); i >= 0 && (best == -1 || i < best) {
				best = i
			}
		}
		if best >= 0 {
			return strings.TrimSpace(l[best:])
		}
	}
	return ""
}

// cleanedLines returns the non-blank lines of raw, each trimmed of surrounding
// space (helm SDK errors carry no box-drawing prefix, unlike terraform).
func cleanedLines(raw string) []string {
	var out []string
	for _, line := range strings.Split(raw, "\n") {
		if t := strings.TrimSpace(line); t != "" {
			out = append(out, t)
		}
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
