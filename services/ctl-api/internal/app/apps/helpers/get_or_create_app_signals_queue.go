package helpers

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

// GetOrCreateAppSignalsQueue returns the app-signals queue for an app, creating
// it on miss. It looks the queue up first so the common (already-exists) path
// never calls queueClient.Create — which would hint-restart the running queue
// workflow and cause restart storms when many signals are enqueued in a burst.
func (h *Helpers) GetOrCreateAppSignalsQueue(ctx context.Context, appID string) (*app.Queue, error) {
	ownerType := plugins.TableName(h.db, app.App{})

	var q app.Queue
	res := h.db.WithContext(ctx).
		Where(app.Queue{OwnerID: appID, OwnerType: ownerType, Name: AppSignalsQueueName}).
		First(&q)
	if res.Error == nil {
		return &q, nil
	}
	if !errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("unable to get app-signals queue for app %s: %w", appID, res.Error)
	}

	created, err := h.queueClient.Create(ctx, &queueclient.CreateQueueRequest{
		OwnerID:     appID,
		OwnerType:   ownerType,
		Namespace:   "apps",
		Name:        AppSignalsQueueName,
		MaxInFlight: 20,
		MaxDepth:    50,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create app-signals queue for app %s: %w", appID, err)
	}

	return created, nil
}
