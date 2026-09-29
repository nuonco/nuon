package executeflow

import (
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
	workflowactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

type CancelGroupRequest struct {
	StepID string `json:"step_id"`
}

type CancelGroupResponse struct {
	WorkflowID string `json:"workflow_id"`
}

func (s *Signal) cancelGroupHandler(ctx workflow.Context, req CancelGroupRequest) (*CancelGroupResponse, error) {
	defer s.beginUpdate()()

	s.cancelRequested = true

	if s.activeGroupQueueSignalID != "" {
		client.AwaitCancelSignal(ctx, s.activeGroupQueueSignalID)
	}

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
