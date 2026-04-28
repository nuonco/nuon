package queue

import (
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
)

const ForceRestartUpdateName string = "force-restart"

type ForceRestartRequest struct{}

type ForceRestartResponse struct{}

// @temporal-gen-v2 update
// @id force-restart
func (q *queue) forceRestartHandler(ctx workflow.Context, req *ForceRestartRequest) (*ForceRestartResponse, error) {
	l, _ := log.WorkflowLogger(ctx)
	if l != nil {
		q.setStatus(ctx, l, QueueStatusForceRestarted)
	}
	q.forceRestarted = true
	q.restarted = true
	return &ForceRestartResponse{}, nil
}
