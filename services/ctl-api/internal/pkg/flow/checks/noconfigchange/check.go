package noconfigchange

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/directive"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

// Check auto-skips an install-group plan step when the run has no config
// changes. The plan signal skips such runs without dispatching an approval
// request, so without this the step parks in AwaitingApproval forever.
type Check struct {
	sig signal.Signal

	// SetResultDirective writes the directive to the step's ResultDirective column.
	SetResultDirective func(ctx workflow.Context, stepID string, d directive.Step) error
}

func New(sig signal.Signal, setDirective func(ctx workflow.Context, stepID string, d directive.Step) error) directive.ApprovalCreateCheck {
	return &Check{sig: sig, SetResultDirective: setDirective}
}

func (c *Check) Name() string { return "no_config_changes" }

func (c *Check) ShouldRun(step *app.WorkflowStep, flw *app.Workflow) bool {
	_, ok := c.sig.(signal.SignalWithNoConfigChangesCheck)
	return ok
}

func (c *Check) Run(ctx workflow.Context, step *app.WorkflowStep, flw *app.Workflow) (directive.CheckResult, error) {
	l, _ := log.WorkflowLogger(ctx)

	nc := c.sig.(signal.SignalWithNoConfigChangesCheck)
	noChanges, err := nc.HasNoConfigChanges(ctx)
	if err != nil {
		return directive.Pass(), errors.Wrap(err, "failed to check for config changes")
	}

	if !noChanges {
		return directive.Pass(), nil
	}

	l.Debug("run has no config changes, auto-skipping plan and deploy",
		zap.String("step_id", step.ID),
		zap.String("workflow_id", flw.ID))

	// Skip the paired deploy step: it's in a separate step group the skip-group
	// directive won't reach, so OnSkip marks it explicitly. applyCheckResult
	// marks the plan step itself.
	if sk, ok := c.sig.(signal.SignalWithOnSkip); ok {
		if err := sk.OnSkip(ctx); err != nil {
			return directive.Pass(), errors.Wrap(err, "unable to skip deploy step for run with no config changes")
		}
	}

	if err := c.SetResultDirective(ctx, step.ID, directive.StepSkipGroup); err != nil {
		return directive.Pass(), errors.Wrap(err, "unable to set skip-group directive for run with no config changes")
	}

	return directive.CheckResult{
		Directive: directive.StepSkipGroup,
		Status:    app.StatusAutoSkipped,
		Reason: directive.CheckReason{
			Check:   "no_config_changes",
			Summary: "Run has no config changes, automatically skipped",
		},
	}, nil
}
