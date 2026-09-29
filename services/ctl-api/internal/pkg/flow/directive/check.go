package directive

import (
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

const StepUnknown Step = ""

type CheckResult struct {
	Directive Step
	Reason    CheckReason

	Status app.Status
}

func Pass() CheckResult { return CheckResult{} }

type CheckReason struct {
	Check string

	Summary string

	Detail string

	Labels map[string]string
}

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
		if k == "auto_approved" {
			m["auto_approved"] = v == "true"
		}
	}
	return m
}

type CheckContext struct {
	NoopPlan bool
}

type ApprovalCreateCheck interface {
	Name() string
	ShouldRun(step *app.WorkflowStep, flw *app.Workflow) bool
	Run(ctx workflow.Context, step *app.WorkflowStep, flw *app.Workflow) (CheckResult, error)
}

type ApprovalResponseCheck interface {
	Name() string
	ShouldRun(step *app.WorkflowStep, flw *app.Workflow, resp *app.WorkflowStepApprovalResponse) bool
	Run(ctx workflow.Context, step *app.WorkflowStep, flw *app.Workflow, resp *app.WorkflowStepApprovalResponse) (CheckResult, error)
}
