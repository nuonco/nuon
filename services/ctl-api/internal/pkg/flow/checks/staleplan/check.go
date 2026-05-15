package staleplan

import (
	"fmt"
	"math"
	"time"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/directive"
)

const ThresholdHours = 72

// Check implements directive.ApprovalResponseCheck for stale plan detection.
// It auto-retries when an approval response arrives more than 72 hours after
// the plan was created, preventing stale plans from being applied.
type Check struct {
	SetResultDirective func(ctx workflow.Context, stepID string, d directive.Step) error
}

func New(setDirective func(workflow.Context, string, directive.Step) error) directive.ApprovalResponseCheck {
	return &Check{SetResultDirective: setDirective}
}

func (c *Check) Name() string { return "stale-plan" }

func (c *Check) ShouldRun(step *app.WorkflowStep, flw *app.Workflow, resp *app.WorkflowStepApprovalResponse) bool {
	return resp.Type == app.WorkflowStepApprovalResponseTypeApprove
}

func (c *Check) Run(ctx workflow.Context, step *app.WorkflowStep, flw *app.Workflow, resp *app.WorkflowStepApprovalResponse) (directive.CheckResult, error) {
	if step.Approval == nil {
		return directive.Pass(), nil
	}

	approvalCreatedAt := step.Approval.CreatedAt
	responseCreatedAt := resp.CreatedAt

	if responseCreatedAt.IsZero() {
		var now time.Time
		_ = workflow.SideEffect(ctx, func(workflow.Context) interface{} {
			return time.Now()
		}).Get(&now)
		responseCreatedAt = now
	}

	age := responseCreatedAt.Sub(approvalCreatedAt)
	threshold := time.Duration(ThresholdHours) * time.Hour

	if age <= threshold {
		return directive.Pass(), nil
	}

	ageHours := int(math.Round(age.Hours()))

	if err := c.SetResultDirective(ctx, step.ID, directive.StepRetryGroup); err != nil {
		return directive.Pass(), fmt.Errorf("unable to set retry-group directive for stale plan: %w", err)
	}

	return directive.CheckResult{
		Directive: directive.StepRetryGroup,
		Reason: directive.CheckReason{
			Check:   "stale-plan",
			Summary: "Plan is stale, auto-retrying",
			Detail:  fmt.Sprintf("Approval was submitted %d hours after plan creation (threshold: %dh)", ageHours, ThresholdHours),
			Labels: map[string]string{
				"stale":           "true",
				"age_hours":       fmt.Sprintf("%d", ageHours),
				"threshold_hours": fmt.Sprintf("%d", ThresholdHours),
			},
		},
	}, nil
}
