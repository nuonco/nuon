package activities

import (
	"context"

	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	dbgenerics "github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
)

type ForgetInstallRequest struct {
	InstallID string `validate:"required"`
}

// @temporal-gen-v2 activity
// @by-field InstallID
func (a *Activities) ForgetInstall(ctx context.Context, req ForgetInstallRequest) error {
	// must run before the cascade delete below soft-deletes the queues
	var queueIDs []string
	if res := a.db.WithContext(ctx).
		Model(&app.Queue{}).
		Where(app.Queue{
			OwnerID:   req.InstallID,
			OwnerType: plugins.TableName(a.db, app.Install{}),
		}).
		Pluck("id", &queueIDs); res.Error != nil {
		return dbgenerics.TemporalGormError(res.Error, "unable to list install queues: %w")
	}

	if len(queueIDs) > 0 {
		if res := a.db.WithContext(ctx).
			Where("queue_id IN ?", queueIDs).
			Delete(&app.QueueEmitter{}); res.Error != nil {
			return dbgenerics.TemporalGormError(res.Error, "unable to delete install queue emitters: %w")
		}
	}

	res := a.db.WithContext(ctx).
		Select(clause.Associations).
		Delete(&app.Install{
			ID: req.InstallID,
		})
	if res.Error != nil {
		return dbgenerics.TemporalGormError(res.Error, "unable to delete install: %w")
	}

	return nil
}
