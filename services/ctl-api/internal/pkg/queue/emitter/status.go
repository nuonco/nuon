package emitter

import (
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/emitter/activities"
)

type StatusRequest struct{}

type StatusResponse struct {
	EmitterID string
	QueueID   string
	EmitCount int64
	Stopped   bool
}

func (e *emitterWorkflow) statusHandler(ctx workflow.Context, req *StatusRequest) (*StatusResponse, error) {
	emitter, err := activities.AwaitGetEmitter(ctx, &activities.GetEmitterRequest{
		EmitterID: e.emitterID,
	})
	if err != nil {
		return nil, err
	}

	return &StatusResponse{
		EmitterID: e.emitterID,
		QueueID:   emitter.QueueID,
		EmitCount: e.state.EmitCount,
		Stopped:   e.stopped,
	}, nil
}
