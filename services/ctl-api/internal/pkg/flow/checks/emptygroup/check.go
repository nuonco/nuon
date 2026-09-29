package emptygroup

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/directive"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

type Check struct {
	sig signal.Signal

	SetResultDirective func(ctx workflow.Context, stepID string, d directive.Step) error
}

func New(sig signal.Signal, setDirective func(ctx workflow.Context, stepID string, d directive.Step) error) directive.ApprovalCreateCheck {
	return &Check{sig: sig, SetResultDirective: setDirective}
}

func (c *Check) Name() string { return "empty_group" }

func (c *Check) ShouldRun(step *app.WorkflowStep, flw *app.Workflow) bool {
	_, ok := c.sig.(signal.SignalWithEmptyGroupCheck)
	return ok
}

func (c *Check) Run(ctx workflow.Context, step *app.WorkflowStep, flw *app.Workflow) (directive.CheckResult, error) {
	l, _ := log.WorkflowLogger(ctx)

	ec := c.sig.(signal.SignalWithEmptyGroupCheck)
	empty, err := ec.IsEmptyInstallGroup(ctx)
	if err != nil {
		return directive.Pass(), errors.Wrap(err, "failed to check for empty install group")
	}

	if !empty {
		return directive.Pass(), nil
	}

	l.Debug("install group has no installs, auto-skipping plan and deploy",
		zap.String("step_id", step.ID),
		zap.String("workflow_id", flw.ID))

	if sk, ok := c.sig.(signal.SignalWithOnSkip); ok {
		if err := sk.OnSkip(ctx); err != nil {
			return directive.Pass(), errors.Wrap(err, "unable to skip deploy step for empty install group")
		}
	}

	if err := c.SetResultDirective(ctx, step.ID, directive.StepSkipGroup); err != nil {
		return directive.Pass(), errors.Wrap(err, "unable to set skip-group directive for empty install group")
	}

	return directive.CheckResult{
		Directive: directive.StepSkipGroup,
		Status:    app.StatusAutoSkipped,
		Reason: directive.CheckReason{
			Check:   "empty_group",
			Summary: "Install group has no installs, automatically skipped",
		},
	}, nil
}
