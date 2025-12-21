package queue

import (
	"go.temporal.io/sdk/workflow"
)

const UnpauseUpdateName string = "unpause"

type UnpauseRequest struct{}

type UnpauseResponse struct{}

func (q *queue) unpauseUpdateHandler(ctx workflow.Context, req *UnpauseRequest) (*UnpauseResponse, error) {
	q.paused = false
	return &UnpauseResponse{}, nil
}
