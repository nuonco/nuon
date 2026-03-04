package emitter

import (
	"go.temporal.io/sdk/workflow"
)

type StopRequest struct{}

type StopResponse struct{}

func (e *emitterWorkflow) stopHandler(ctx workflow.Context, req *StopRequest) (*StopResponse, error) {
	e.stopped = true
	return &StopResponse{}, nil
}
