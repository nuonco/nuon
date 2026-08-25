package activities

import (
	"context"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	dbgenerics "github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
	"gorm.io/gorm/clause"
)

type DeleteRequest struct {
	InstallID string `validate:"required"`
}

// @temporal-gen-v2 activity
// @by-field InstallID
func (a *Activities) Delete(ctx context.Context, req DeleteRequest) error {
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
			return dbgenerics.TemporalGormError(res.Error, "unable to delete queue emitters: %w")
		}

		if res := a.db.WithContext(ctx).
			Where("queue_id IN ?", queueIDs).
			Delete(&app.QueueSignal{}); res.Error != nil {
			return dbgenerics.TemporalGormError(res.Error, "unable to delete queue signals: %w")
		}

	// Stack service accounts are only reachable by naming convention while the
	// stack rows still exist; see DeleteInstallStackServiceAccounts.
	if err := a.acctClient.DeleteInstallStackServiceAccounts(ctx, req.InstallID); err != nil {
		return err

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
