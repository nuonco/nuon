package activities

import (
	"context"
	"fmt"

	dbgenerics "github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type StopProcessQueueRequest struct {
	RunnerID  string `validate:"required"`
	ProcessID string `validate:"required"`
}

// @temporal-gen-v2 activity
// @by-field RunnerID
// @local
func (a *Activities) StopProcessQueue(ctx context.Context, req StopProcessQueueRequest) error {
	queueName := fmt.Sprintf("runner-process-%s", req.ProcessID)

	var q app.Queue
	if res := a.db.WithContext(ctx).
		Where(&app.Queue{
			OwnerID:   req.RunnerID,
			OwnerType: "runners",
			Name:      queueName,
		}).First(&q); res.Error != nil {
		return dbgenerics.TemporalGormError(res.Error, fmt.Sprintf("unable to find process queue %s", queueName))
	}

	if err := a.helpers.StopProcessQueue(ctx, q.ID); err != nil {
		return fmt.Errorf("unable to stop process queue: %w", err)
	}

	return nil
}
