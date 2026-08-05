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
	installOwnerType := plugins.TableName(a.db, app.Install{})

	// must run before the cascade delete below soft-deletes the queues
	var queueIDs []string
	if res := a.db.WithContext(ctx).
		Model(&app.Queue{}).
		Where(app.Queue{
			OwnerID:   req.InstallID,
			OwnerType: installOwnerType,
		}).
		Pluck("id", &queueIDs); res.Error != nil {
		return dbgenerics.TemporalGormError(res.Error, "unable to list install queues: %w")
	}

	var runnerGroupIDs []string
	if res := a.db.WithContext(ctx).
		Model(&app.RunnerGroup{}).
		Where(app.RunnerGroup{
			OwnerID:   req.InstallID,
			OwnerType: installOwnerType,
		}).
		Pluck("id", &runnerGroupIDs); res.Error != nil {
		return dbgenerics.TemporalGormError(res.Error, "unable to list install runner groups: %w")
	}

	var runnerIDs []string
	if len(runnerGroupIDs) > 0 {
		if res := a.db.WithContext(ctx).
			Model(&app.Runner{}).
			Where("runner_group_id IN ?", runnerGroupIDs).
			Pluck("id", &runnerIDs); res.Error != nil {
			return dbgenerics.TemporalGormError(res.Error, "unable to list install runners: %w")
		}
	}

	var runnerQueueIDs []string
	if len(runnerIDs) > 0 {
		if res := a.db.WithContext(ctx).
			Model(&app.Queue{}).
			Where(app.Queue{OwnerType: plugins.TableName(a.db, app.Runner{})}).
			Where("owner_id IN ?", runnerIDs).
			Pluck("id", &runnerQueueIDs); res.Error != nil {
			return dbgenerics.TemporalGormError(res.Error, "unable to list runner queues: %w")
		}
	}

	allQueueIDs := append(queueIDs, runnerQueueIDs...)
	if len(allQueueIDs) > 0 {
		if res := a.db.WithContext(ctx).
			Where("queue_id IN ?", allQueueIDs).
			Delete(&app.QueueEmitter{}); res.Error != nil {
			return dbgenerics.TemporalGormError(res.Error, "unable to delete queue emitters: %w")
		}
	}

	if len(runnerQueueIDs) > 0 {
		if res := a.db.WithContext(ctx).
			Where("id IN ?", runnerQueueIDs).
			Delete(&app.Queue{}); res.Error != nil {
			return dbgenerics.TemporalGormError(res.Error, "unable to delete runner queues: %w")
		}
	}

	if len(runnerIDs) > 0 {
		if res := a.db.WithContext(ctx).
			Where("id IN ?", runnerIDs).
			Delete(&app.Runner{}); res.Error != nil {
			return dbgenerics.TemporalGormError(res.Error, "unable to delete runners: %w")
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
