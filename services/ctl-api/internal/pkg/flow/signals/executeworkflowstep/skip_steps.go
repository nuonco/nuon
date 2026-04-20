package executeworkflowstep

import (
	"fmt"
	"slices"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
	activities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

func getStepIndex(stepID string, steps []app.WorkflowStep) int {
	for i, s := range steps {
		if s.ID == stepID {
			return i
		}
	}
	return -1
}

// markWorkflowApprovalPlanDenied marks the approval step and its group siblings as denied/skipped.
func (s *Signal) markWorkflowApprovalPlanDenied(ctx workflow.Context, flw *app.Workflow, approvalStep *app.WorkflowStep) error {
	var groupSteps []app.WorkflowStep
	for _, step := range flw.Steps {
		if step.GroupIdx == approvalStep.GroupIdx {
			groupSteps = append(groupSteps, step)
		}
	}
	if len(groupSteps) == 0 {
		return fmt.Errorf("workflow steps for groupIdx %d not found", approvalStep.GroupIdx)
	}

	if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: approvalStep.ID,
		Status: app.CompositeStatus{
			Status:                 app.WorkflowStepApprovalStatusApprovalDenied,
			StatusHumanDescription: "Plan changes denied, skipping current step group",
		},
	}); err != nil {
		return errors.Wrap(err, "unable to update step to success status")
	}

	if err := activities.AwaitPkgWorkflowsFlowUpdateFlowStepTargetStatus(ctx, activities.UpdateFlowStepTargetStatusRequest{
		StepID:            approvalStep.ID,
		Status:            app.Status(app.InstallDeployApprovalDenied),
		StatusDescription: "Approval denied",
	}); err != nil {
		return errors.Wrap(err, "unable to update step target status")
	}

	for _, step := range groupSteps {
		if step.ID == approvalStep.ID {
			continue
		}

		if !slices.Contains([]app.Status{
			app.StatusPending,
			app.AwaitingApproval,
			app.StatusNotAttempted,
			app.WorkflowStepApprovalStatusApprovalRetryPlan,
		}, step.Status.Status) {
			continue
		}

		if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
			ID: step.ID,
			Status: app.CompositeStatus{
				Status:                 app.StatusUserSkipped,
				StatusHumanDescription: "Plan denied and skipped by the user.",
			},
		}); err != nil {
			return errors.Wrap(err, "unable to update step to success status")
		}
	}

	return nil
}

// markDependentStepsAsSkipped marks the approval step as denied and skips dependent steps.
func (s *Signal) markDependentStepsAsSkipped(ctx workflow.Context, flw *app.Workflow, step *app.WorkflowStep) error {
	if err := s.markWorkflowApprovalPlanDenied(ctx, flw, step); err != nil {
		return errors.Wrap(err, "unable to mark workflow steps approval denied")
	}

	switch app.WorkflowStepTargetType(step.StepTargetType) {
	case app.WorkflowStepTargetTypeInstallSandboxRun, app.WorkflowStepTargetTypeInstallSandboxRuns:
		if err := s.markAllComponentDeployStepsSkipped(ctx, flw); err != nil {
			return errors.Wrap(err, "unable to update step to retry plan status")
		}
	case app.WorkflowStepTargetTypeInstallDeploy, app.WorkflowStepTargetTypeInstallDeploys:
		// Future: skip dependent components
	}
	return nil
}

// markAllComponentDeployStepsSkipped marks all component deploy steps as skipped.
func (s *Signal) markAllComponentDeployStepsSkipped(ctx workflow.Context, flw *app.Workflow) error {
	var groupsToSkip []int
	for _, step := range flw.Steps {
		if app.WorkflowStepTargetType(step.StepTargetType) == app.WorkflowStepTargetTypeInstallDeploy || app.WorkflowStepTargetType(step.StepTargetType) == app.WorkflowStepTargetTypeInstallDeploys {
			groupsToSkip = append(groupsToSkip, step.GroupIdx)
		}
	}

	for _, step := range flw.Steps {
		if slices.Contains(groupsToSkip, step.GroupIdx) {
			if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
				ID: step.ID,
				Status: app.CompositeStatus{
					Status:                 app.StatusUserSkipped,
					StatusHumanDescription: "Plan denied and skipped by the user.",
				},
			}); err != nil {
				return errors.Wrap(err, "unable to update step to success status")
			}
		}
	}

	return nil
}
