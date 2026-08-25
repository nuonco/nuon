package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
	"gorm.io/gorm/clause"
)

type DeleteRequest struct {
	AppID string `validate:"required"`
}

// @temporal-gen-v2 activity
// @by-field AppID
func (a *Activities) Delete(ctx context.Context, req DeleteRequest) error {
	var queueIDs []string
	if res := a.db.WithContext(ctx).
		Model(&app.Queue{}).
		Where(app.Queue{
			OwnerID:   req.AppID,
			OwnerType: plugins.TableName(a.db, app.App{}),
		}).
		Pluck("id", &queueIDs); res.Error != nil {
		return fmt.Errorf("unable to list app queues: %w", res.Error)
	}

	if len(queueIDs) > 0 {
		if res := a.db.WithContext(ctx).
			Where("queue_id IN ?", queueIDs).
			Delete(&app.QueueEmitter{}); res.Error != nil {
			return fmt.Errorf("unable to delete queue emitters: %w", res.Error)
		}

		if res := a.db.WithContext(ctx).
			Where("queue_id IN ?", queueIDs).
			Delete(&app.QueueSignal{}); res.Error != nil {
			return fmt.Errorf("unable to delete queue signals: %w", res.Error)
		}

		if res := a.db.WithContext(ctx).
			Where("id IN ?", queueIDs).
			Delete(&app.Queue{}); res.Error != nil {
			return fmt.Errorf("unable to delete queues: %w", res.Error)
		}
	}

	res := a.db.WithContext(ctx).
		Select(clause.Associations).
		Delete(&app.App{
			ID: req.AppID,
		})
	if res.Error != nil {
		return fmt.Errorf("unable to delete app: %w", res.Error)
	}

	return nil
}
