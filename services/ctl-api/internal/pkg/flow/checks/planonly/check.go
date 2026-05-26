package planonly

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/driftdetected"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/directive"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
	activities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

// Check implements directive.ApprovalCreateCheck for plan-only mode.
type Check struct {
	OwnerID  string
	checkCtx *directive.CheckContext
}

func New(ownerID string, checkCtx *directive.CheckContext) directive.ApprovalCreateCheck {
	return &Check{OwnerID: ownerID, checkCtx: checkCtx}
}

func (c *Check) Name() string { return "plan-only" }

func (c *Check) ShouldRun(step *app.WorkflowStep, flw *app.Workflow) bool {
	return flw.PlanOnly
}

func (c *Check) Run(ctx workflow.Context, step *app.WorkflowStep, flw *app.Workflow) (directive.CheckResult, error) {
	noopPlan := c.checkCtx.NoopPlan

	if err := handlePlanOnlyApproval(ctx, c.OwnerID, step, noopPlan); err != nil {
		return directive.Pass(), errors.Wrap(err, "failed to handle plan-only auto-approval")
	}

	summary := "Drift detected"
	if noopPlan {
		summary = "No drift detected"
	}

	return directive.CheckResult{
		Directive: directive.StepContinue,
		// handlePlanOnlyApproval already wrote the step's authoritative status
		// (Approved) and the step-target status (Drifted / NoDrift). Set Status
		// here so applyCheckResult does not default to StatusError and overwrite
		// the row we just persisted.
		Status: app.WorkflowStepApprovalStatusApproved,
		Reason: directive.CheckReason{
			Check:   "plan-only",
			Summary: summary,
			Labels: map[string]string{
				"plan_only": "true",
				"noop":      boolStr(noopPlan),
			},
		},
	}, nil
}

func boolStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

func handlePlanOnlyApproval(ctx workflow.Context, ownerID string, step *app.WorkflowStep, noopPlan bool) error {
	statusDescription := "Auto-approved in plan-only mode."
	targetStatus := app.WorkflowStepApprovalStatusApproved

	if noopPlan {
		statusDescription = "No drift detected"
		targetStatus = app.WorkflowStepNoDrift
	} else {
		statusDescription = "Drift detected"
		targetStatus = app.WorkflowStepDrifted
	}

	if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: step.ID,
		Status: app.CompositeStatus{
			Status:                 app.WorkflowStepApprovalStatusApproved,
			StatusHumanDescription: "auto-approved (plan-only mode) " + step.Name,
			Metadata: map[string]any{
				"step_idx":  step.Idx,
				"status":    "auto-approved",
				"plan_only": true,
				"no_op":     noopPlan,
			},
		},
	}); err != nil {
		return errors.Wrap(err, "unable to update step to auto-approved status")
	}

	if err := activities.AwaitPkgWorkflowsFlowUpdateFlowStepTargetStatus(ctx, activities.UpdateFlowStepTargetStatusRequest{
		StepID:            step.ID,
		Status:            targetStatus,
		StatusDescription: statusDescription,
	}); err != nil {
		return errors.Wrap(err, "unable to update step target status")
	}

	if !noopPlan {
		if err := driftdetected.Dispatch(ctx, &driftdetected.Signal{
			InstallID:         ownerID,
			InstallWorkflowID: step.InstallWorkflowID,
			WorkflowStepID:    step.ID,
			OwnerID:           step.StepTargetID,
			OwnerType:         string(step.StepTargetType),
		}); err != nil {
			workflow.GetLogger(ctx).Warn("unable to dispatch drift-detected signal",
				"step_id", step.ID,
				"error", err)
		}
	}

	return nil
}
