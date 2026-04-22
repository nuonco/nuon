package executeflow

import (
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
	workflowactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

// CancelGroupRequest is the input for the "cancel-group" update handler.
type CancelGroupRequest struct {
	// StepID is any step in the group to cancel. The handler resolves the
	// group from this step's GroupIdx and cancels all in-flight steps.
	StepID string `json:"step_id"`
}

// CancelGroupResponse is the response from the "cancel-group" update handler.
type CancelGroupResponse struct {
	WorkflowID string `json:"workflow_id"`
}

// cancelGroupHandler cancels the currently executing group signal and marks
// the workflow as cancelled. The group signal's Cancel() method handles
// propagating cancellation to in-flight steps.
func (s *Signal) cancelGroupHandler(ctx workflow.Context, req CancelGroupRequest) (*CancelGroupResponse, error) {
	s.cancelRequested = true

	// Cancel the active group signal.
	if s.activeGroupQueueSignalID != "" {
		client.AwaitCancelSignal(ctx, s.activeGroupQueueSignalID)
	}

	// Mark the workflow as cancelled.
	_ = workflowactivities.AwaitPkgWorkflowsFlowUpdateFlowFinishedAtByID(ctx, s.WorkflowID)
	_ = statusactivities.AwaitPkgStatusUpdateFlowStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: s.WorkflowID,
		Status: app.CompositeStatus{
			Status:                 app.StatusCancelled,
			StatusHumanDescription: "workflow cancelled",
		},
	})

	return &CancelGroupResponse{WorkflowID: s.WorkflowID}, nil
}
