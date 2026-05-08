package executeworkflowstep

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
	activities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

// runNoopCheck checks if the plan has no changes and auto-skips if so.
// Returns done=true when the step was auto-skipped and the workflow is not plan-only.
func (s *Signal) runNoopCheck(ctx workflow.Context, l *zap.Logger, step *app.WorkflowStep, flw *app.Workflow, noopPlan *bool) (bool, error) {
	var err error
	*noopPlan, err = activities.AwaitCheckNoopPlan(ctx, &activities.CheckNoopPlanRequest{
		StepTargetID: step.StepTargetID,
	})
	if err != nil {
		return false, errors.Wrap(err, "failed to check for noop plan")
	}
	if *noopPlan {
		l.Debug("approval plan contents empty",
			zap.String("step_id", step.ID),
			zap.String("workflow_id", flw.ID))
		if err := s.handleNoopDeployPlan(ctx, step, flw); err != nil {
			return false, errors.Wrap(err, "failed to handle noop plan")
		}
		if !flw.PlanOnly {
			// Write skip-group directive so the rest of the group (e.g. apply) is skipped.
			if err := setResultDirective(ctx, step.ID, DirectiveSkipGroup); err != nil {
				return false, errors.Wrap(err, "unable to set skip-group directive for noop plan")
			}
			return true, nil
		}
	}
	return false, nil
}

// handleNoopDeployPlan handles the case where the plan has no changes.
func (s *Signal) handleNoopDeployPlan(ctx workflow.Context, step *app.WorkflowStep, flw *app.Workflow) error {
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

	currentStepIndex := getStepIndex(step.ID, flw.Steps)
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
