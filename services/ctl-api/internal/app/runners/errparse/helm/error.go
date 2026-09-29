// why: Package helm holds tool-layer CompositeError parsers for helm jobs. They
// register at errparse.LayerTool so a provider-specific cause (e.g. an AWS IAM
// denial parsed at LayerProvider) still wins, but any recognised helm failure
// yields a clean, structured error instead of falling through to the raw
// generic dump.
//
// The runner drives helm through the Go SDK (helm.sh/helm/v4 pkg/action), not
// the CLI, so the familiar "INSTALLATION FAILED"/"UPGRADE FAILED" phase
// prefixes never appear, since those are added only by helm's cmd layer. What is
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
	HelmErrorType             compositeerrors.Type = "helm.error"
	HelmImmutableFieldType    compositeerrors.Type = "helm.immutable_field"
	HelmOwnershipConflictType compositeerrors.Type = "helm.ownership_conflict"
	HelmNameInUseType         compositeerrors.Type = "helm.name_in_use"
	HelmNoDeployedReleaseType compositeerrors.Type = "helm.no_deployed_release"
	HelmHookFailedType        compositeerrors.Type = "helm.hook_failed"
	HelmWaitTimeoutType       compositeerrors.Type = "helm.wait_timeout"
	HelmRenderErrorType       compositeerrors.Type = "helm.render_error"
)

const (
	maxHeadline = 240
	maxBody     = 8000
)

var wrappers = []string{"helm release:", "with dry-run:"}

// why: causes are verified helm v4 SDK (pkg/action, pkg/kube) error substrings, used
// as a backup anchor when the captured output does not carry the runner wrapper
// (e.g. only a log line was retained). Every entry is helm-specific; generic
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

type classifier struct {
	typ   compositeerrors.Type
	hints compositeerrors.Hints
	match func(summary string) bool
}

var skipRetry = compositeerrors.NewHints().WithSkipAutoRetry()

var classifiers = []classifier{
	{HelmOwnershipConflictType, skipRetry, contains("exists and cannot be imported into the current release", "invalid ownership metadata")},
	{HelmImmutableFieldType, skipRetry, contains("field is immutable", "Forbidden: updates to")},
	{HelmNameInUseType, skipRetry, contains("cannot reuse a name that is still in use")},
	{HelmNoDeployedReleaseType, skipRetry, contains("has no deployed releases")},
	{HelmRenderErrorType, skipRetry, contains("unable to build kubernetes objects from", "YAML parse error", "template:", "unable to recognize")},
	{HelmHookFailedType, nil, contains("hooks failed", "failed pre-install", "failed post-install")},
	{HelmWaitTimeoutType, nil, contains("timed out waiting for the condition")},
}

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

func signals() []string {
	return append(append([]string{}, wrappers...), causes...)
}

func parseError(ctx *errparse.ParseContext) compositeerrors.CompositeError {
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
	errparse.Register(errparse.NewParser(errparse.LayerTool, parseError,
		errparse.WithTools(errparse.ToolHelm),
		errparse.WithSignals(signals()...),
	))
}

// why: wrapperSummary returns the SDK error that follows the runner's helm wrapper on
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

func cleanedLines(raw string) []string {
	var out []string
	for _, line := range strings.Split(raw, "\n") {
		if t := strings.TrimSpace(line); t != "" {
			out = append(out, t)
		}
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
