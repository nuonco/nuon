package queue

import (
	"go.temporal.io/sdk/workflow"
)

const StopUpdateName string = "stop"

type StopRequest struct{}

type StopResponse struct{}

// EnqueueSignal adds the signal to the queue and returns the db id and workflow id
func (q *queue) stopUpdateHandler(ctx workflow.Context, req *StopRequest) (*StopResponse, error) {
	// stop children
	q.stopped = true
	return &StopResponse{}, nil
}
