package emitter

import (
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/emitter/activities"
)

type EnsureRunningRequest struct{}

type EnsureRunningResponse struct {
	Running   bool
	EmitterID string
	Mode      string
	EmitCount int64
	Stopped   bool
	Fired     bool
}

func (e *emitterWorkflow) ensureRunningHandler(ctx workflow.Context, req *EnsureRunningRequest) (*EnsureRunningResponse, error) {
	emitter, err := activities.AwaitGetEmitter(ctx, &activities.GetEmitterRequest{
		EmitterID: e.emitterID,
	})
	if err != nil {
		return nil, err
	}

	return &EnsureRunningResponse{
		Running:   true,
		EmitterID: e.emitterID,
		Mode:      string(emitter.Mode),
		EmitCount: e.state.EmitCount,
		Stopped:   e.stopped,
		Fired:     emitter.Fired,
	}, nil
}
