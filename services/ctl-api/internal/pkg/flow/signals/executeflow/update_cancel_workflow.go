package executeflow

import (
	"fmt"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/directive"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
	workflowactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

// CancelWorkflowResponse is the response from the "cancel-workflow" update handler.
type CancelWorkflowResponse struct {
	WorkflowID string `json:"workflow_id"`
}

// cancelWorkflowHandler cancels the entire workflow. It actively cancels the
// currently running group signal (which cascades to steps and calls Cancel()
// callbacks), then marks the workflow as cancelled.
func (s *Signal) cancelWorkflowHandler(ctx workflow.Context) (*CancelWorkflowResponse, error) {
	defer s.beginUpdate()()

	s.cancelRequested = true

	if s.Resident {
		steps, err := workflowactivities.AwaitPkgWorkflowsFlowGetFlowSteps(ctx, workflowactivities.GetFlowStepsRequest{
			FlowID: s.WorkflowID,
		})
		if err != nil {
			return nil, fmt.Errorf("unable to load steps for resident workflow cancellation: %w", err)
		}
		var residentStep *app.WorkflowStep
		for i := range steps {
			if directive.Step(steps[i].ResultDirective) == directive.StepAwaitRetry {
				residentStep = &steps[i]
				break
			}
		}
		if residentStep == nil {
			for i := range steps {
				if !isStepTerminal(steps[i].Status.Status) {
					residentStep = &steps[i]
					break
				}
			}
		}
		if residentStep != nil {
			if _, err := workflowactivities.AwaitForwardCancelStepToGroup(ctx, workflowactivities.ForwardCancelStepToGroupRequest{
				StepID:      residentStep.ID,
				StepGroupID: residentStep.WorkflowStepGroupID,
			}); err != nil {
				if l, _ := log.WorkflowLogger(ctx); l != nil {
					l.Warn("cancel-workflow: unable to forward cancel to resident group",
						zap.String("step_id", residentStep.ID),
						zap.Error(err))
				}
			}
			if err := s.cancelResidentStep(ctx, residentStep); err != nil {
				return nil, err
			}
		}
		if err := s.finalizeCancellation(ctx); err != nil {
			return nil, err
		}
	} else {
		// Persist cancel_requested_at in metadata so downstream signals (groups,
		// steps) can detect cancellation even if the in-memory flag hasn't
		// propagated yet. This survives ContinueAsNew and is the durable
		// source of truth for cancellation.
		_ = statusactivities.AwaitPkgStatusUpdateFlowStatus(ctx, statusactivities.UpdateStatusRequest{
			ID: s.WorkflowID,
			Status: app.CompositeStatus{
				Status:                 app.StatusCancelled,
				StatusHumanDescription: "workflow cancelled",
				Metadata: map[string]any{
					"cancel_requested_at": workflow.Now(ctx).Unix(),
				},
			},
		})
	}

	// Cancel the active group signal. This triggers the group's Cancel()
	// method which propagates to step signals and their inner signals.
	if s.activeGroupQueueSignalID != "" {
		client.AwaitCancelSignal(ctx, s.activeGroupQueueSignalID)
	}

	if !s.Resident {
		_ = workflowactivities.AwaitPkgWorkflowsFlowUpdateFlowFinishedAtByID(ctx, s.WorkflowID)
	}

	return &CancelWorkflowResponse{WorkflowID: s.WorkflowID}, nil
}
