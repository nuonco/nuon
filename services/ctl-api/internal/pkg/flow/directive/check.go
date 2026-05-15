package directive

import (
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// StepUnknown means the check has no opinion — pass through to the next check.
const StepUnknown Step = ""

// CheckResult is the outcome of an approval check: a directive and a reason.
// StepUnknown (empty) means "no opinion, continue to next check."
// Any other directive short-circuits the pipeline.
type CheckResult struct {
	Directive Step
	Reason    CheckReason

	// Status overrides the step status written by applyCheckResult.
	// When empty, defaults to StatusError for backward compatibility.
	Status app.Status
}

// Pass returns a result with no opinion.
func Pass() CheckResult { return CheckResult{} }

// CheckReason carries structured metadata about why a check produced its result.
type CheckReason struct {
	// Check is the identifier of the check (e.g. "noop", "policy", "stale-plan").
	Check string

	// Summary is a human-readable one-liner shown in the dashboard.
	Summary string

	// Detail is a longer explanation shown in step details.
	Detail string

	// Labels are machine-readable key-value pairs for filtering/routing.
	Labels map[string]string
}

// Metadata returns the reason as a map suitable for CompositeStatus.Metadata.
func (r CheckReason) Metadata() map[string]any {
	m := map[string]any{
		"check":   r.Check,
		"summary": r.Summary,
	}
	if r.Detail != "" {
		m["detail"] = r.Detail
	}
	for k, v := range r.Labels {
		m["check_label_"+k] = v
		// Promote auto_approved to top level for UI badge detection.
		if k == "auto_approved" {
			m["auto_approved"] = v == "true"
		}
	}
	return m
}

// CheckContext carries state shared between checks in a single pipeline run.
type CheckContext struct {
	NoopPlan bool
}

// ApprovalCreateCheck runs BEFORE the approval is presented to the user.
// Return StepUnknown to pass through; any other directive short-circuits.
type ApprovalCreateCheck interface {
	Name() string
	ShouldRun(step *app.WorkflowStep, flw *app.Workflow) bool
	Run(ctx workflow.Context, step *app.WorkflowStep, flw *app.Workflow) (CheckResult, error)
}

// ApprovalResponseCheck runs AFTER an approval response is received.
// Return StepUnknown to let the normal response handler run; any other directive overrides it.
type ApprovalResponseCheck interface {
	Name() string
	ShouldRun(step *app.WorkflowStep, flw *app.Workflow, resp *app.WorkflowStepApprovalResponse) bool
	Run(ctx workflow.Context, step *app.WorkflowStep, flw *app.Workflow, resp *app.WorkflowStepApprovalResponse) (CheckResult, error)
}
