package activities

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/queue/signal"
)

// NOTE(jm): signal.Signal is not compatible with temporal-gen
func AwaitGetQueueSignalSignal(ctx workflow.Context, req *GetQueueSignalSignalRequest) (signal.Signal, error) {
	_ = (&Activities{}).GetQueueSignalSignal

	ao := workflow.ActivityOptions{
		ScheduleToCloseTimeout: 30 * time.Minute,
		StartToCloseTimeout:    5 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	fut := workflow.ExecuteActivity(ctx, (&Activities{}).GetQueueSignalSignal, req)

	var sig signal.Signal
	if err := fut.Get(ctx, &sig); err != nil {
		return nil, err
	}

	return sig, nil
}

func AwaitGetQueueSignalSignalByQueueSignalID(ctx workflow.Context, queueSignalID string) (signal.Signal, error) {
	return AwaitGetQueueSignalSignal(ctx, &GetQueueSignalSignalRequest{QueueSignalID: queueSignalID})
}

type GetQueueSignalSignalRequest struct {
	QueueSignalID string `validate:"required"`
}

func (a *Activities) GetQueueSignalSignal(ctx context.Context, req *GetQueueSignalSignalRequest) (signal.Signal, error) {
	queueSignal, err := a.getQueueSignal(ctx, req.QueueSignalID)
	if err != nil {
		return nil, generics.TemporalGormError(err, "unable to get queue signal")
	}

	return queueSignal.Signal.Signal, nil
}

func (a *Activities) getQueueSignal(ctx context.Context, queueID string) (*app.QueueSignal, error) {
	var qs app.QueueSignal

	if res := a.db.WithContext(ctx).
		Where(app.QueueSignal{
			ID: queueID,
		}).
		First(&qs); res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get queue signal")
	}

	return &qs, nil
}
