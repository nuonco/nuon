package queue

import (
	"go.temporal.io/sdk/workflow"
)

const StopUpdateName string = "stop"

type StopRequest struct{}

type StopResponse struct{}

func (q *queue) stopUpdateHandler(ctx workflow.Context, req *StopRequest) (*StopResponse, error) {
	q.stopped = true
	return &StopResponse{}, nil
}
