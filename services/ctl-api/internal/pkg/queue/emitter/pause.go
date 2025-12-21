package emitter

import (
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/emitter/activities"
)

type PauseRequest struct{}

type PauseResponse struct {
	Paused bool
}

func (e *emitterWorkflow) pauseHandler(ctx workflow.Context, req *PauseRequest) (*PauseResponse, error) {
	if _, err := activities.AwaitUpdateEmitterStatus(ctx, &activities.UpdateEmitterStatusRequest{
		EmitterID: e.emitterID,
		Status:    app.StatusCancelled,
	}); err != nil {
		return nil, err
	}

	return &PauseResponse{Paused: true}, nil
}

type ResumeRequest struct{}

type ResumeResponse struct {
	Paused bool
}

func (e *emitterWorkflow) resumeHandler(ctx workflow.Context, req *ResumeRequest) (*ResumeResponse, error) {
	if _, err := activities.AwaitUpdateEmitterStatus(ctx, &activities.UpdateEmitterStatusRequest{
		EmitterID: e.emitterID,
		Status:    app.StatusInProgress,
	}); err != nil {
		return nil, err
	}

	return &ResumeResponse{Paused: false}, nil
}
