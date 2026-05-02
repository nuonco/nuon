package activities

import (
	"context"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
)

type GetInstallSignalsQueueRequest struct {
	InstallID string `validate:"required"`
}

// @temporal-gen-v2 activity
// @by-field InstallID
// @start-to-close-timeout 1m
func (a *Activities) GetInstallSignalsQueue(ctx context.Context, req GetInstallSignalsQueueRequest) (*app.Queue, error) {
	var queue app.Queue
	res := a.db.WithContext(ctx).
		Where(app.Queue{
			OwnerID: req.InstallID,
			Name:    helpers.InstallSignalsQueueName,
		}).
		First(&queue)
	if res.Error != nil {
		return nil, generics.TemporalGormError(res.Error, "get install signals queue")
	}

	return &queue, nil
}
