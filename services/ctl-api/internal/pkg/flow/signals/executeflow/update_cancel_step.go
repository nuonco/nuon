package executeflow

import (
	"fmt"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/directive"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
	workflowactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

// CancelStepRequest is the input for the "cancel-step" update handler.
type CancelStepRequest struct {
	StepID string `json:"step_id"`
}

// CancelStepResponse is the response from the "cancel-step" update handler.
type CancelStepResponse struct {
	WorkflowID string `json:"workflow_id"`
}

// cancelStepHandler sets cancelRequested immediately and propagates the
// cancellation to a live group. Resident flows also persist the equivalent
// cancellation directly so the same update works after descendants unwind.
func (s *Signal) cancelStepHandler(ctx workflow.Context, req CancelStepRequest) (*CancelStepResponse, error) {
	s.updatesInFlight++
	defer func() { s.updatesInFlight-- }()

	s.cancelRequested = true

	l, _ := log.WorkflowLogger(ctx)

	step, err := workflowactivities.AwaitPkgWorkflowsFlowGetFlowsStepByFlowStepID(ctx, req.StepID)
	if err != nil {
		if l != nil {
			l.Warn("cancel-step: unable to get step",
				zap.String("step_id", req.StepID),
				zap.Error(err))
		}
		return &CancelStepResponse{WorkflowID: s.WorkflowID}, nil
	}
	if _, err := workflowactivities.AwaitForwardCancelStepToGroup(ctx, workflowactivities.ForwardCancelStepToGroupRequest{
		StepID:      req.StepID,
		StepGroupID: step.WorkflowStepGroupID,
	}); err != nil {
		if l != nil {
			l.Warn("cancel-step: unable to forward cancel to group",
				zap.String("step_id", req.StepID),
				zap.Error(err))
		}
	}

	if s.Resident {
		if err := s.cancelResidentStep(ctx, step); err != nil {
			return nil, err
		}
		if err := s.finishResidentCancellation(ctx); err != nil {
			return nil, err
		}
	}

	return &CancelStepResponse{WorkflowID: s.WorkflowID}, nil
}

func (s *Signal) cancelResidentStep(ctx workflow.Context, step *app.WorkflowStep) error {
	if err := workflowactivities.AwaitPkgWorkflowsFlowUpdateFlowStepResultDirective(ctx, workflowactivities.UpdateFlowStepResultDirectiveRequest{
		StepID:    step.ID,
		Directive: string(directive.StepStop),
	}); err != nil {
		return fmt.Errorf("unable to write cancel directive for step %s: %w", step.ID, err)
	}
	if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: step.ID,
		Status: app.CompositeStatus{
			Status:                 app.StatusCancelled,
			StatusHumanDescription: "Step was cancelled by the user.",
		},
	}); err != nil {
		return fmt.Errorf("unable to mark step %s cancelled: %w", step.ID, err)
	}

	if err := workflowactivities.AwaitPkgWorkflowsFlowUpdateFlowStepTargetStatus(ctx, workflowactivities.UpdateFlowStepTargetStatusRequest{
		StepID:            step.ID,
		Status:            app.StatusCancelled,
		StatusDescription: "Cancelled",
	}); err != nil {
		return fmt.Errorf("unable to mark target for step %s cancelled: %w", step.ID, err)
	}

	steps, err := workflowactivities.AwaitPkgWorkflowsFlowGetFlowSteps(ctx, workflowactivities.GetFlowStepsRequest{
		FlowID: s.WorkflowID,
	})
	if err != nil {
		return fmt.Errorf("unable to load steps during resident cancellation: %w", err)
	}
	for _, candidate := range steps {
		if candidate.ID == step.ID || isStepTerminal(candidate.Status.Status) {
			continue
		}

		status := app.StatusNotAttempted
		description := "Step was not attempted because the workflow was cancelled."
		sameGroup := candidate.GroupIdx == step.GroupIdx
		if step.WorkflowStepGroupID != "" {
			sameGroup = candidate.WorkflowStepGroupID == step.WorkflowStepGroupID
		}
		if sameGroup {
			status = app.StatusCancelled
			description = "Step was cancelled with its group."
		} else if candidate.GroupIdx < step.GroupIdx {
			continue
		}

		if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
			ID: candidate.ID,
			Status: app.CompositeStatus{
				Status:                 status,
				StatusHumanDescription: description,
			},
		}); err != nil {
			return fmt.Errorf("unable to update step %s during resident cancellation: %w", candidate.ID, err)
		}
	}

	groups, err := workflowactivities.AwaitPkgWorkflowsFlowGetFlowStepGroups(ctx, s.WorkflowID)
	if err != nil {
		return fmt.Errorf("unable to load groups during resident cancellation: %w", err)
	}
	for _, group := range groups {
		if group.GroupIdx < step.GroupIdx {
			continue
		}

		targetGroup := group.ID == step.WorkflowStepGroupID || (step.WorkflowStepGroupID == "" && group.GroupIdx == step.GroupIdx)
		if !targetGroup && isStepTerminal(group.Status.Status) {
			continue
		}

		groupStatus := app.StatusDiscarded
		groupDirective := directive.GroupStop
		description := "group discarded because the workflow was cancelled"
		if targetGroup {
			groupStatus = app.StatusCancelled
			description = "group cancelled"
		}
		if err := workflowactivities.AwaitPkgWorkflowsFlowUpdateFlowStepGroupResultDirective(ctx, workflowactivities.UpdateFlowStepGroupResultDirectiveRequest{
			StepGroupID: group.ID,
			Directive:   string(groupDirective),
		}); err != nil {
			return fmt.Errorf("unable to update group %s cancel directive: %w", group.ID, err)
		}
		if err := statusactivities.AwaitPkgStatusUpdateFlowStepGroupStatus(ctx, statusactivities.UpdateStatusRequest{
			ID: group.ID,
			Status: app.CompositeStatus{
				Status:                 groupStatus,
				StatusHumanDescription: description,
			},
		}); err != nil {
			return fmt.Errorf("unable to update group %s during resident cancellation: %w", group.ID, err)
		}
	}

	return nil
}

func (s *Signal) finishResidentCancellation(ctx workflow.Context) error {
	if err := workflowactivities.AwaitPkgWorkflowsFlowUpdateFlowFinishedAtByID(ctx, s.WorkflowID); err != nil {
		return fmt.Errorf("unable to mark resident workflow finished: %w", err)
	}
	if err := statusactivities.AwaitPkgStatusUpdateFlowStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: s.WorkflowID,
		Status: app.CompositeStatus{
			Status:                 app.StatusCancelled,
			StatusHumanDescription: "workflow cancelled",
			Metadata: map[string]any{
				"cancel_requested_at": workflow.Now(ctx).Unix(),
			},
		},
	}); err != nil {
		return fmt.Errorf("unable to mark resident workflow cancelled: %w", err)
	}
	return nil
}
