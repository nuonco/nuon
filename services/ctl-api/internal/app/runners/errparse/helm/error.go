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
//
// The anchored cause is then classified into a specific helm failure type
// (helm.immutable_field, helm.ownership_conflict, ...) which carries a distinct
// discriminator (for UI badging and the parse-coverage metric) and retry hints
// (deterministic config errors set skip_auto_retry so the orchestrator parks
// the step for manual retry instead of burning attempts). An unclassified helm
// failure still yields the generic helm.error so it beats the raw dump.
package helm

import (
	"strings"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/errparse"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
)

const (
	// HelmErrorType is the fallback discriminator for a helm failure that no
	// specific classifier recognised.
	HelmErrorType compositeerrors.Type = "helm.error"
	// HelmImmutableFieldType is a rejected patch to an immutable/forbidden field
	// (e.g. a StatefulSet selector), which a blind retry can never fix.
	HelmImmutableFieldType compositeerrors.Type = "helm.immutable_field"
	// HelmOwnershipConflictType is a resource that already exists and is not
	// owned by this release, so it cannot be adopted.
	HelmOwnershipConflictType compositeerrors.Type = "helm.ownership_conflict"
	// HelmNameInUseType is a release name still held by another (often stuck)
	// release.
	HelmNameInUseType compositeerrors.Type = "helm.name_in_use"
	// HelmNoDeployedReleaseType is an upgrade with no prior deployed release.
	HelmNoDeployedReleaseType compositeerrors.Type = "helm.no_deployed_release"
	// HelmHookFailedType is a lifecycle hook (pre/post install/upgrade) that
	// failed; this can be transient so it stays auto-retryable.
	HelmHookFailedType compositeerrors.Type = "helm.hook_failed"
	// HelmWaitTimeoutType is a --wait timeout on resources becoming ready; often
	// transient (image pull, scheduling) so it stays auto-retryable.
	HelmWaitTimeoutType compositeerrors.Type = "helm.wait_timeout"
	// HelmRenderErrorType is a chart render / manifest build failure (template
	// execution, YAML, unknown kind), a deterministic config error.
	HelmRenderErrorType compositeerrors.Type = "helm.render_error"
)

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

// causes are verified helm v4 SDK (pkg/action, pkg/kube) error substrings, used
// as a backup anchor when the captured output does not carry the runner wrapper
// (e.g. only a log line was retained). Every entry is helm-specific — generic
// kubernetes phrases like "timed out waiting for the condition" are
// deliberately excluded, since they can appear in streamed pod logs before the
// real cause and would produce a misleading headline. (That phrase is still
// used to classify an already-anchored cause below, just never to anchor one.)
var causes = []string{
	"cannot reuse a name that is still in use",
	"exists and cannot be imported into the current release",
	"invalid ownership metadata",
	"unable to build kubernetes objects from",
	"cannot patch",
	"unable to recognize",
	"failed pre-install",
	"failed post-install",
	"pre-upgrade hooks failed",
	"post-upgrade hooks failed",
	"failed to install CRD",
	"has no deployed releases",
	"another operation (install/upgrade/rollback) is in progress",
	"chart dependencies processing failed",
	"YAML parse error",
}

// classifier maps an anchored cause to a specific helm failure type. The first
// classifier whose match hits wins, so more specific ones are listed first.
type classifier struct {
	typ   compositeerrors.Type
	hints compositeerrors.Hints
	match func(summary string) bool
}

// skipRetry marks a deterministic failure a blind retry cannot fix, so the
// orchestrator parks the step for manual retry instead of burning attempts.
var skipRetry = compositeerrors.Hints{compositeerrors.HintSkipAutoRetry: "true"}

var classifiers = []classifier{
	{HelmOwnershipConflictType, skipRetry, contains("exists and cannot be imported into the current release", "invalid ownership metadata")},
	{HelmImmutableFieldType, skipRetry, contains("field is immutable", "Forbidden: updates to")},
	{HelmNameInUseType, skipRetry, contains("cannot reuse a name that is still in use")},
	{HelmNoDeployedReleaseType, skipRetry, contains("has no deployed releases")},
	{HelmRenderErrorType, skipRetry, contains("unable to build kubernetes objects from", "YAML parse error", "template:", "unable to recognize")},
	{HelmHookFailedType, nil, contains("hooks failed", "failed pre-install", "failed post-install")},
	{HelmWaitTimeoutType, nil, contains("timed out waiting for the condition")},
}

// contains returns a matcher that hits when the summary contains any of subs.
func contains(subs ...string) func(string) bool {
	return func(summary string) bool {
		for _, s := range subs {
			if strings.Contains(summary, s) {
				return true
			}
		}
		return false
	}
}

// HelmError is the tool-layer payload: the classified helm failure. Summary is
// the SDK error with the runner's wrapper stripped; Reason is the specific
// failure class (empty for the generic fallback); Output is the cleaned context.
type HelmError struct {
	Reason  string `json:"reason,omitempty"`
	Summary string `json:"summary"`
	Output  string `json:"output,omitempty"`

	typ   compositeerrors.Type
	hints compositeerrors.Hints
}

var (
	_ compositeerrors.CompositeError = (*HelmError)(nil)
	_ compositeerrors.HintsProvider  = (*HelmError)(nil)
)

func (e *HelmError) Error() string                      { return truncate(e.Summary, maxHeadline) }
func (e *HelmError) Type() compositeerrors.Type         { return e.typ }
func (e *HelmError) Severity() compositeerrors.Severity { return compositeerrors.SeverityError }
func (e *HelmError) Hints() compositeerrors.Hints       { return e.hints }

func (e *HelmError) Sections() []compositeerrors.Section {
	if e.Output == "" {
		return nil
	}
	return []compositeerrors.Section{
		compositeerrors.CodeSection("Output", e.Output),
	}
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

	typ, hints := HelmErrorType, compositeerrors.Hints(nil)
	for _, c := range classifiers {
		if c.match(summary) {
			typ, hints = c.typ, c.hints
			break
		}
	}

	e := &HelmError{
		Summary: summary,
		Output:  truncate(strings.Join(lines, "\n"), maxBody),
		typ:     typ,
		hints:   hints,
	}
	if typ != HelmErrorType {
		e.Reason = strings.TrimPrefix(string(typ), "helm.")
	}
	return e
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
