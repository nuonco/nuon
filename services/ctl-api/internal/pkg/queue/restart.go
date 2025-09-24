package queue

import "go.temporal.io/sdk/workflow"

const RestartUpdateName string = "restart"

type RestartRequest struct{}

type RestartResponse struct{}

func (q *queue) restartUpdateHandler(ctx workflow.Context, req *RestartRequest) (*RestartResponse, error) {
	q.restarted = true
	return &RestartResponse{}, nil
}
