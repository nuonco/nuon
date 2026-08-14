package helm

import (
	"regexp"
	"strings"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/errparse"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
)

// HelmPendingOperationType is split out of the generic helm.error so the step
// parks with targeted remediation instead of burning retries a pending release
// can never clear.
const HelmPendingOperationType compositeerrors.Type = "helm.pending_operation"

const (
	// The runner's own message, which also names the status and revision.
	stuckSignal = "is stuck in pending-"
	// Helm's SDK wording, for jobs that predate the runner check.
	inProgressSignal = "another operation (install/upgrade/rollback) is in progress"
)

// "cannot reuse a name that is still in use" is deliberately NOT a signal: helm
// emits it for a genuine name collision too, where recovery is wrong advice.

var pendingStatusPattern = regexp.MustCompile(`pending-(install|upgrade|rollback)`)

// PendingOperationError is the payload for a deploy blocked by a stuck release.
type PendingOperationError struct {
	Status string `json:"status,omitempty"`
	Output string `json:"output,omitempty"`
}

var (
	_ compositeerrors.CompositeError = (*PendingOperationError)(nil)
	_ compositeerrors.HintsProvider  = (*PendingOperationError)(nil)
)

func (e *PendingOperationError) Error() string {
	if e.Status != "" {
		return "Helm release is stuck in " + e.Status + " from an operation that never finished"
	}
	return "Helm release is stuck from an operation that never finished"
}

func (e *PendingOperationError) Type() compositeerrors.Type { return HelmPendingOperationType }

func (e *PendingOperationError) Severity() compositeerrors.Severity {
	return compositeerrors.SeverityError
}

func (e *PendingOperationError) Hints() compositeerrors.Hints {
	return compositeerrors.NewHints().WithSkipAutoRetry()
}

func (e *PendingOperationError) Sections() []compositeerrors.Section {
	sections := []compositeerrors.Section{
		compositeerrors.MarkdownSection("How to fix", pendingOperationRemediation(e.Status)),
	}
	if e.Output != "" {
		sections = append(sections, compositeerrors.CodeSection("Output", e.Output))
	}
	return sections
}

// Registers at LayerToolSpecific so it wins the tie against the helm catch-all.
func parsePendingOperation(ctx *errparse.ParseContext) compositeerrors.CompositeError {
	lines := cleanedLines(ctx.Raw)
	if !containsPendingOperation(lines) {
		return nil
	}

	return &PendingOperationError{
		Status: pendingStatus(lines),
		Output: truncate(strings.Join(lines, "\n"), maxBody),
	}
}

func init() {
	errparse.Register(errparse.NewParser(errparse.LayerToolSpecific, parsePendingOperation,
		errparse.WithTools(errparse.ToolHelm),
		errparse.WithSignals(stuckSignal, inProgressSignal),
	))
}

func containsPendingOperation(lines []string) bool {
	for _, l := range lines {
		if strings.Contains(l, stuckSignal) || strings.Contains(l, inProgressSignal) {
			return true
		}
	}
	return false
}

// Empty when only helm's status-less SDK wording is present.
func pendingStatus(lines []string) string {
	for _, l := range lines {
		if m := pendingStatusPattern.FindString(l); m != "" {
			return m
		}
	}
	return ""
}

// No "confirm nothing is running" step, unlike a state lock: the recovery itself
// refuses to act unless the release really is pending.
func pendingOperationRemediation(status string) string {
	var intro string
	if status != "" {
		intro = "The Helm release for this component is stuck in `" + status + "`. "
	} else {
		intro = "The Helm release for this component is stuck part-way through an operation. "
	}

	return intro +
		"Helm records a pending status before it starts changing the cluster and clears it when the " +
		"operation finishes, so a release left this way is a rollout whose runner went away — a crash, " +
		"a cancelled workflow, or a job that timed out. Helm refuses every further operation on the " +
		"release until it is recovered, and retrying the deploy cannot clear it.\n\n" +
		"To resolve:\n\n" +
		"1. Recover the release. Nuon rolls it back to the last revision that finished a rollout, or " +
		"removes the release when it never rolled out at all:\n\n" +
		"```\nnuon installs components recover-helm-release --component <component>\n```\n\n" +
		"2. Deploy the component again.\n\n" +
		"You can also recover from the dashboard: open the component and use " +
		"**Recover Helm release** on the stuck-release banner."
}
