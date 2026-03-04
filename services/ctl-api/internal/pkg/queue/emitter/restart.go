package emitter

import (
	"go.temporal.io/sdk/workflow"
)

type RestartRequest struct{}

type RestartResponse struct{}

func (e *emitterWorkflow) restartHandler(ctx workflow.Context, req *RestartRequest) (*RestartResponse, error) {
	e.restarted = true
	return &RestartResponse{}, nil
}
