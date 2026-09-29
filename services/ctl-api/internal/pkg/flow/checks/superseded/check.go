package superseded

import (
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/directive"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

type Check struct {
	stepSignalFn func(step *app.WorkflowStep) signal.Signal

	SetResultDirective func(ctx workflow.Context, stepID string, d directive.Step) error
}

func New(
	stepSignalFn func(step *app.WorkflowStep) signal.Signal,
	setDirective func(workflow.Context, string, directive.Step) error,
) directive.ApprovalResponseCheck {
	return &Check{
		stepSignalFn:       stepSignalFn,
		SetResultDirective: setDirective,
	}
}

func (c *Check) Name() string { return "superseded" }

func (c *Check) ShouldRun(step *app.WorkflowStep, flw *app.Workflow, resp *app.WorkflowStepApprovalResponse) bool {
	if resp.Type != app.WorkflowStepApprovalResponseTypeApprove {
		return false
	}
	sig := c.stepSignalFn(step)
	_, ok := sig.(signal.SignalWithApprovalValidation)
	return ok
}

func (c *Check) Run(ctx workflow.Context, step *app.WorkflowStep, flw *app.Workflow, resp *app.WorkflowStepApprovalResponse) (directive.CheckResult, error) {
	sig := c.stepSignalFn(step)
	av, ok := sig.(signal.SignalWithApprovalValidation)
	if !ok {
		return directive.Pass(), nil
	}

	err := av.ValidateApproval(ctx)
	if err == nil {
		return directive.Pass(), nil
	}

	if dirErr := c.SetResultDirective(ctx, step.ID, directive.StepRetryGroup); dirErr != nil {
		return directive.Pass(), dirErr
	}

	return directive.CheckResult{
		Directive: directive.StepRetry,
		Reason: directive.CheckReason{
			Check:   "superseded",
			Summary: "Plan superseded, auto-retrying",
			Detail:  err.Error(),
			Labels: map[string]string{
				"superseded": "true",
			},
		},
	}, nil
}
