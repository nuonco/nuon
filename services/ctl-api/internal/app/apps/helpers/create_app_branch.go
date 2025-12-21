package helpers

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

func (h *Helpers) CreateAppBranch(
	ctx context.Context,
	appID string,
	name string,
) (*app.AppBranch, error) {
	branch := app.AppBranch{
		AppID: appID,
		Name:  name,
	}

	// Create branch first to get ID
	if err := h.db.WithContext(ctx).Create(&branch).Error; err != nil {
		return nil, fmt.Errorf("unable to create app branch: %w", err)
	}

	// Create queue for app branch
	_, err := h.queueClient.Create(ctx, &queueclient.CreateQueueRequest{
		OwnerID:     branch.ID,
		OwnerType:   plugins.TableName(h.db, app.AppBranch{}),
		Namespace:   "apps",
		MaxInFlight: 1,
		MaxDepth:    50,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create queue: %w", err)
	}

	return &branch, nil
}
