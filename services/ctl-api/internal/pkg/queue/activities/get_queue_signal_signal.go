package activities

import (
	"context"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/catalog"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	"go.temporal.io/sdk/temporal"
)

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
// @as-wrapper
// @wrapper-prefix QueueInternal
// @by-field queueSignalID
func (a *Activities) getQueueSignalSignal(ctx context.Context, queueSignalID string) (signal.Signal, error) {
	queueSignal, err := a.getQueueSignal(ctx, queueSignalID)
	if err != nil {
		return nil, generics.TemporalGormError(err, "unable to get queue signal")
	}

	if queueSignal.Signal.Signal == nil {
		return nil, temporal.NewNonRetryableApplicationError(
			"signal type not registered in catalog",
			catalog.SignalTypeNotRegisteredErrorType,
			nil,
		)
	}
	return queueSignal.Signal.Signal, nil
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
// @as-wrapper
// @wrapper-prefix QueueInternal
// @by-field queueSignalID
func (a *Activities) getQueueSignal(ctx context.Context, queueSignalID string) (*app.QueueSignal, error) {
	var qs app.QueueSignal

	if res := a.db.WithContext(ctx).
		Where(app.QueueSignal{
			ID: queueSignalID,
		}).
		First(&qs); res.Error != nil {
		return nil, generics.TemporalGormError(res.Error, "unable to get queue signal")
	}

	return &qs, nil
}
