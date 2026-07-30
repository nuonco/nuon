package activities

import (
	"context"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
)

type GetInstallDriftCronSignalsQueueRequest struct {
	InstallID string `validate:"required"`
}

// @temporal-gen-v2 activity
// @by-field InstallID
// @start-to-close-timeout 1m
func (a *Activities) GetInstallDriftCronSignalsQueue(ctx context.Context, req GetInstallDriftCronSignalsQueueRequest) (*app.Queue, error) {
	var queue app.Queue
	res := a.db.WithContext(ctx).
		Where(app.Queue{
			OwnerID: req.InstallID,
			Name:    helpers.InstallDriftCronSignalsQueueName,
		}).
		First(&queue)
	if res.Error != nil {
		return nil, generics.TemporalGormError(res.Error, "get install drift cron signals queue")
	}

	return &queue, nil
}
