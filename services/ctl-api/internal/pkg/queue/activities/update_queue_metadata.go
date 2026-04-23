package activities

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
)

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
// @as-wrapper
// @wrapper-prefix QueueInternal
func (a *Activities) updateQueueMetadata(ctx context.Context, queueID string, metadata map[string]*string) error {
	var queue app.Queue
	if res := a.db.WithContext(ctx).Where(app.Queue{ID: queueID}).First(&queue); res.Error != nil {
		return generics.TemporalGormError(res.Error, "unable to get queue for metadata update")
	}

	if queue.Metadata == nil {
		queue.Metadata = pgtype.Hstore{}
	}

	for k, v := range metadata {
		if v == nil {
			delete(queue.Metadata, k)
		} else {
			queue.Metadata[k] = v
		}
	}

	if res := a.db.WithContext(ctx).Model(&app.Queue{}).Where(app.Queue{ID: queueID}).Update("metadata", queue.Metadata); res.Error != nil {
		return generics.TemporalGormError(res.Error, "unable to update queue metadata")
	}

	return nil
}
