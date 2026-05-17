package noop

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/directive"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
	activities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

// Check implements directive.ApprovalCreateCheck for noop plan detection.
type Check struct {
	sig      signal.Signal
	checkCtx *directive.CheckContext

	// OrgAutoSkipNoop is true when the org has the auto-skip-noop feature flag enabled.
	// Passed in by the caller to avoid import cycles with install activities.
	OrgAutoSkipNoop bool

	// SetResultDirective writes the directive to the step's ResultDirective column.
	SetResultDirective func(ctx workflow.Context, stepID string, d directive.Step) error
}

func New(sig signal.Signal, checkCtx *directive.CheckContext, orgAutoSkipNoop bool, setDirective func(ctx workflow.Context, stepID string, d directive.Step) error) directive.ApprovalCreateCheck {
	return &Check{sig: sig, checkCtx: checkCtx, OrgAutoSkipNoop: orgAutoSkipNoop, SetResultDirective: setDirective}
}

func (c *Check) Name() string { return "noop" }

func (c *Check) ShouldRun(step *app.WorkflowStep, flw *app.Workflow) bool {
	nc, ok := c.sig.(signal.SignalWithNoOpCheck)
	return ok && nc.IsNoOpCheckable()
}

func (c *Check) Run(ctx workflow.Context, step *app.WorkflowStep, flw *app.Workflow) (directive.CheckResult, error) {
	l, _ := log.WorkflowLogger(ctx)

	isNoop, err := activities.AwaitCheckNoopPlan(ctx, &activities.CheckNoopPlanRequest{
		StepTargetID: step.StepTargetID,
	})
	if err != nil {
		return directive.Pass(), errors.Wrap(err, "failed to check for noop plan")
	}

	c.checkCtx.NoopPlan = isNoop

	if !isNoop {
		return directive.Pass(), nil
	}

	// Determine whether noop plans should be auto-skipped. Only skip when
	// explicitly enabled at the org level OR the component level.
	shouldSkip := c.OrgAutoSkipNoop
	if !shouldSkip {
		if sn, ok := c.sig.(signal.SignalWithSkipNoops); ok {
			shouldSkip = sn.SkipNoops(ctx)
		}
	}

	if !shouldSkip {
		l.Debug("noop plan detected but skip_noops not enabled, proceeding to approval",
			zap.String("step_id", step.ID))

		// Decorate the step with a noop label so the approval UI can display it,
		// then pass through to let the approval pipeline handle it normally.
		_ = statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
			ID: step.ID,
			Status: app.CompositeStatus{
				Status:                 app.StatusCheckPlan,
				StatusHumanDescription: "noop plan detected, awaiting approval",
				Metadata: map[string]any{
					"noop": true,
				},
			},
		})

		return directive.Pass(), nil
	}

	l.Debug("noop plan detected and skip_noops enabled",
		zap.String("step_id", step.ID),
		zap.Bool("org_auto_skip", c.OrgAutoSkipNoop))

	l.Debug("approval plan contents empty",
		zap.String("step_id", step.ID),
		zap.String("workflow_id", flw.ID))

	if err := handleNoopDeployPlan(ctx, step, flw); err != nil {
		return directive.Pass(), errors.Wrap(err, "failed to handle noop plan")
	}

	if flw.PlanOnly {
		return directive.Pass(), nil
	}

	if err := c.SetResultDirective(ctx, step.ID, directive.StepSkipGroup); err != nil {
		return directive.Pass(), errors.Wrap(err, "unable to set skip-group directive for noop plan")
	}

	return directive.CheckResult{
		Directive: directive.StepSkipGroup,
		Status:    app.StatusAutoSkipped,
		Reason: directive.CheckReason{
			Check:   "noop",
			Summary: "Noop plan, automatically skipped",
		},
	}, nil
}

func handleNoopDeployPlan(ctx workflow.Context, step *app.WorkflowStep, flw *app.Workflow) error {
	if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: step.ID,
		Status: app.CompositeStatus{
			Status:                 app.StatusAutoSkipped,
			StatusHumanDescription: "Noop Plan, automatically skipped " + step.Name,
			Metadata: map[string]any{
				"step_idx": step.Idx,
				"status":   "auto-skipped",
			},
		},
	}); err != nil {
		return errors.Wrap(err, "unable to update step to success status")
	}

	currentStepIndex := -1
	for i, s := range flw.Steps {
		if s.ID == step.ID {
			currentStepIndex = i
			break
		}
	}
	if currentStepIndex == -1 {
		return errors.Errorf("step index not found for step id: %s", step.ID)
	}

	nextStepIndex := currentStepIndex + 1
	if nextStepIndex >= len(flw.Steps) {
		return nil
	}

	nextStep := flw.Steps[nextStepIndex]

	if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: nextStep.ID,
		Status: app.CompositeStatus{
			Status:                 app.StatusAutoSkipped,
			StatusHumanDescription: "Noop Plan, automatically skipped " + nextStep.Name,
			Metadata: map[string]any{
				"step_idx": nextStep.Idx,
				"status":   "auto-skipped",
			},
		},
	}); err != nil {
		return errors.Wrap(err, "unable to update step to success status")
	}

	if err := activities.AwaitPkgWorkflowsFlowUpdateFlowStepTargetStatus(ctx, activities.UpdateFlowStepTargetStatusRequest{
		StepID:            step.ID,
		Status:            app.StatusAutoSkipped,
		StatusDescription: "No changes found in plan, skipping deployment.",
	}); err != nil {
		return errors.Wrap(err, "unable to update step target status")
	}

	if err := activities.AwaitSyncNoopDeployOutputs(ctx, &activities.SyncNoopDeployOutputsRequest{
		StepID: step.ID,
	}); err != nil {
		l, _ := log.WorkflowLogger(ctx)
		if l != nil {
			l.Warn("unable to sync noop deploy outputs", zap.Error(err))
		}
	}

	return nil
}
