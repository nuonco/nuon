package helm

import (
	"regexp"
	"strings"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/errparse"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
)

// HelmPendingOperationType is the discriminator for a deploy blocked by a release
// that helm left mid-operation. It is split out of the generic helm.error so the
// dashboard can surface targeted remediation (recover the release) and so the
// orchestrator parks the step instead of burning retries: helm writes a pending
// status before it touches the cluster and only clears it when the operation
// finishes, so a release left pending is a rollout whose driver went away and no
// number of retries will move it.
const HelmPendingOperationType compositeerrors.Type = "helm.pending_operation"

const (
	// stuckSignal is the runner's own deterministic message, emitted when it sees
	// a pending release before attempting an apply. It is the preferred signal
	// because it is unambiguous and names the status and revision.
	stuckSignal = "is stuck in pending-"
	// inProgressSignal is helm's SDK error for the same condition, kept as a
	// fallback for jobs that predate the runner check.
	inProgressSignal = "another operation (install/upgrade/rollback) is in progress"
)

// Deliberately NOT a signal: "cannot reuse a name that is still in use". Helm
// emits it both for a pending release and for a genuine name collision with a
// release Nuon does not own, and recovery is wrong advice for the latter. That
// string stays on helm.name_in_use.

// pendingStatusPattern pulls the helm status out of the runner's message so the
// headline can name it. Helm only ever has these three pending statuses.
var pendingStatusPattern = regexp.MustCompile(`pending-(install|upgrade|rollback)`)

// PendingOperationError is the payload for a deploy blocked by a stuck release.
// Status is the helm status when it could be recovered from the output, empty
// otherwise; Output carries the captured context.
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

// parsePendingOperation recognises a deploy blocked by a pending helm release. It
// registers at LayerToolSpecific, independently of the helm catch-all, so the two
// classifiers stay decoupled and the more specific one wins the tie.
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

// containsPendingOperation reports whether any cleaned line carries one of the
// pending-release signals.
func containsPendingOperation(lines []string) bool {
	for _, l := range lines {
		if strings.Contains(l, stuckSignal) || strings.Contains(l, inProgressSignal) {
			return true
		}
	}
	return false
}

// pendingStatus returns the helm status named in the output, or "" when only the
// SDK's status-less "another operation ... in progress" wording is present.
func pendingStatus(lines []string) string {
	for _, l := range lines {
		if m := pendingStatusPattern.FindString(l); m != "" {
			return m
		}
	}
	return ""
}

// pendingOperationRemediation renders the "How to fix" body (markdown). Recovery
// is presented as the whole fix because, unlike a state lock, there is nothing
// for the operator to confirm first: the recovery itself refuses to act unless
// the release really is pending.
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
		"You can also recover from the dashboard: open the component and choose " +
		"**Recover Helm release** under Component controls."
}
