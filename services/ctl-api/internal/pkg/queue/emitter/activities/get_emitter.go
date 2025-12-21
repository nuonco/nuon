package activities

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type GetEmitterRequest struct {
	EmitterID string `validate:"required"`
}

func AwaitGetEmitter(ctx workflow.Context, req *GetEmitterRequest) (*app.QueueEmitter, error) {
	ao := workflow.ActivityOptions{
		ScheduleToCloseTimeout: 30 * time.Second,
		StartToCloseTimeout:    5 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	fut := workflow.ExecuteActivity(ctx, (&Activities{}).GetEmitter, req)
	var ret app.QueueEmitter
	if err := fut.Get(ctx, &ret); err != nil {
		return nil, err
	}
	return &ret, nil
}

// @temporal-gen activity
// @by-id EmitterID
func (a *Activities) GetEmitter(ctx context.Context, req *GetEmitterRequest) (*app.QueueEmitter, error) {
	var emitter app.QueueEmitter

	if res := a.db.WithContext(ctx).
		Where("id = ?", req.EmitterID).
		First(&emitter); res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get emitter")
	}

	return &emitter, nil
}
