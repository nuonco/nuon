package executeworkflowstep

import (
	"strconv"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
	activities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

// runPlanOnlyCheck auto-approves steps in plan-only mode.
// Always returns done=true (plan-only mode terminates the plan check flow).
func (s *Signal) runPlanOnlyCheck(ctx workflow.Context, step *app.WorkflowStep, noopPlan bool) (bool, error) {
	if err := s.handlePlanOnlyApproval(ctx, step, noopPlan); err != nil {
		return false, errors.Wrap(err, "failed to handle plan-only auto-approval")
	}
	return true, nil
}

// handlePlanOnlyApproval auto-approves steps in plan-only mode.
func (s *Signal) handlePlanOnlyApproval(ctx workflow.Context, step *app.WorkflowStep, noopPlan bool) error {
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
			StatusHumanDescription: "auto-approved (plan-only mode) " + strconv.Itoa(step.Idx+1),
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

	return nil
}
