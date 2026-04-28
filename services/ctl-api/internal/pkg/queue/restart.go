package queue

import (
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
)

const RestartUpdateName string = "restart"

type RestartRequest struct{}

type RestartResponse struct{}

func (q *queue) restartUpdateHandler(ctx workflow.Context, req *RestartRequest) (*RestartResponse, error) {
	l, _ := log.WorkflowLogger(ctx)
	if l != nil {
		q.setStatus(ctx, l, QueueStatusRestartPending)
	}
	q.restarted = true
	return &RestartResponse{}, nil
}
