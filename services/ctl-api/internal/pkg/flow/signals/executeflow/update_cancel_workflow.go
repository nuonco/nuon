package executeflow

import (
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
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
	s.cancelRequested = true

	// Cancel the active group signal. This triggers the group's Cancel()
	// method which propagates to step signals and their inner signals.
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

	return &CancelWorkflowResponse{WorkflowID: s.WorkflowID}, nil
}
