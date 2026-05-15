package staleplan

import (
	"fmt"
	"math"
	"time"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/directive"
)

const DefaultThreshold = 72 * time.Hour

// Check implements directive.ApprovalResponseCheck for stale plan detection.
// It auto-retries when an approval response arrives after the threshold
// duration since the plan was created, preventing stale plans from being applied.
type Check struct {
	// Threshold is the max age before a plan is considered stale.
	// Defaults to DefaultThreshold (72h) when zero.
	Threshold time.Duration

	SetResultDirective func(ctx workflow.Context, stepID string, d directive.Step) error
}

func New(threshold time.Duration, setDirective func(workflow.Context, string, directive.Step) error) directive.ApprovalResponseCheck {
	if threshold == 0 {
		threshold = DefaultThreshold
	}
	return &Check{Threshold: threshold, SetResultDirective: setDirective}
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

	if age <= c.Threshold {
		return directive.Pass(), nil
	}

	thresholdMinutes := int(math.Round(c.Threshold.Minutes()))
	ageMinutes := int(math.Round(age.Minutes()))

	if err := c.SetResultDirective(ctx, step.ID, directive.StepRetryGroup); err != nil {
		return directive.Pass(), fmt.Errorf("unable to set retry-group directive for stale plan: %w", err)
	}

	return directive.CheckResult{
		Directive: directive.StepRetry,
		Reason: directive.CheckReason{
			Check:   "stale-plan",
			Summary: "Plan is stale, auto-retrying",
			Detail:  fmt.Sprintf("Approval was submitted %dm after plan creation (threshold: %dm)", ageMinutes, thresholdMinutes),
			Labels: map[string]string{
				"stale":             "true",
				"age_minutes":       fmt.Sprintf("%d", ageMinutes),
				"threshold_minutes": fmt.Sprintf("%d", thresholdMinutes),
			},
		},
	}, nil
}
