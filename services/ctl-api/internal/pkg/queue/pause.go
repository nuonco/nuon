package queue

import (
	"go.temporal.io/sdk/workflow"
)

const PauseUpdateName string = "pause"

type PauseRequest struct{}

type PauseResponse struct{}

func (q *queue) pauseUpdateHandler(ctx workflow.Context, req *PauseRequest) (*PauseResponse, error) {
	q.paused = true
	return &PauseResponse{}, nil
}
