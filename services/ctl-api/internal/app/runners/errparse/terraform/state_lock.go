package terraform

import (
	"strings"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/errparse"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
)

// TerraformStateLockType is the discriminator for a terraform run that could
// not acquire the state lock. It is split out of the generic terraform.error so
// the dashboard can surface targeted remediation (force-unlock) and so the
// orchestrator parks the step instead of burning retries: Nuon serialises
// operations per install/sandbox, so a lock is almost always stale (left by a
// crashed or cancelled run) rather than a run genuinely in progress, and a
// blind retry can never clear it.
const TerraformStateLockType compositeerrors.Type = "terraform.state_lock"

// stateLockSignal is the substring terraform emits when it cannot take the DynamoDB/
// backend lock. It gates the parser so it is only a candidate for lock failures.
const stateLockSignal = "acquiring the state lock"

// StateLockError is the payload for a terraform state-lock failure. The detailed
// lock info (ID, who, path) is emitted to the log stream, not the captured error
// output, so it is not available here — Output carries whatever context was
// captured. target ("sandbox"/"component"/"") tailors the force-unlock command
// in the remediation.
type StateLockError struct {
	Output string `json:"output,omitempty"`

	target string
}

var (
	_ compositeerrors.CompositeError = (*StateLockError)(nil)
	_ compositeerrors.HintsProvider  = (*StateLockError)(nil)
)

func (e *StateLockError) Error() string              { return "Terraform could not acquire the state lock" }
func (e *StateLockError) Type() compositeerrors.Type { return TerraformStateLockType }
func (e *StateLockError) Severity() compositeerrors.Severity {
	return compositeerrors.SeverityError
}

func (e *StateLockError) Hints() compositeerrors.Hints {
	return compositeerrors.Hints{compositeerrors.HintSkipAutoRetry: "true"}
}

func (e *StateLockError) Sections() []compositeerrors.Section {
	sections := []compositeerrors.Section{{
		Heading: "How to fix",
		Body:    stateLockRemediation(e.target),
	}}
	if e.Output != "" {
		sections = append(sections, compositeerrors.Section{
			Heading: "Output",
			Body:    "```\n" + e.Output + "\n```",
		})
	}
	return sections
}

// stateLockParser recognises a terraform state-lock failure and yields a
// dedicated composite error. It registers independently of the terraform
// catch-all so the two classifiers stay decoupled.
type stateLockParser struct{}

func (stateLockParser) Layer() errparse.Layer                  { return errparse.LayerToolSpecific }
func (stateLockParser) Tools() []errparse.Tool                 { return []errparse.Tool{errparse.ToolTerraform} }
func (stateLockParser) Signals() []string                      { return []string{stateLockSignal} }
func (stateLockParser) Applicable(*errparse.ParseContext) bool { return true }

func (stateLockParser) Parse(ctx *errparse.ParseContext) compositeerrors.CompositeError {
	lines := cleanedLines(ctx.Raw)
	if !containsStateLock(lines) {
		return nil
	}
	return &StateLockError{
		Output: truncate(strings.Join(lines, "\n"), maxBody),
		target: stateLockTarget(ctx),
	}
}

func init() {
	errparse.Register(stateLockParser{})
}

// containsStateLock reports whether any cleaned line carries the state-lock
// signal. Gating on the cleaned lines (not the raw blob) keeps it robust to
// terraform's "│" box-drawing prefixes.
func containsStateLock(lines []string) bool {
	for _, l := range lines {
		if strings.Contains(l, stateLockSignal) {
			return true
		}
	}
	return false
}

// stateLockRemediation renders the "How to fix" body (markdown). It gates
// force-unlock on confirming nothing is running, since force-unlocking an
// active operation can corrupt state, and tailors the CLI command to the
// workspace kind.
func stateLockRemediation(target string) string {
	var cmd string
	switch target {
	case "sandbox":
		cmd = "nuon terraform sandbox force-unlock <LOCK_ID>"
	case "component":
		cmd = "nuon terraform component force-unlock <LOCK_ID> --name <component>"
	default:
		cmd = "nuon terraform sandbox force-unlock <LOCK_ID>\n" +
			"# or, for a component workspace:\n" +
			"nuon terraform component force-unlock <LOCK_ID> --name <component>"
	}

	return "The Terraform state is locked, most likely by a previous run that did not release it. " +
		"Nuon runs one operation at a time per install, so this is almost always a stale lock " +
		"rather than a run in progress.\n\n" +
		"To resolve:\n\n" +
		"1. Confirm no other run for this install or sandbox is currently in progress.\n" +
		"2. Copy the lock ID from the run logs (view details of the `Error aquiring the state lock` block for `ID: ...`).\n" +
		"3. Force-unlock the workspace with the Nuon terraform CLI extension, then retry:\n\n" +
		"```\n" + cmd + "\n```\n\n" +
		"You can also unlock from the dashboard: open the Terraform state panel and choose **Unlock Terraform state**."
}

// stateLockTarget maps the runner job's owner to the terraform extension
// workspace kind used in the remediation command.
func stateLockTarget(ctx *errparse.ParseContext) string {
	switch ctx.Owner.Type {
	case "install_sandbox_runs":
		return "sandbox"
	case "install_deploys":
		return "component"
	default:
		return ""
	}
}
