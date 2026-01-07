package helpers

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

func (h *Helpers) CreateAppBranch(ctx context.Context, orgID, appID, name string, connectedGithubVCSConfigID string) (*app.AppBranch, error) {
	// Start transaction
	tx := h.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("unable to start transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	branch := app.AppBranch{
		OrgID:                      orgID,
		AppID:                      appID,
		Name:                       name,
		ConnectedGithubVCSConfigID: connectedGithubVCSConfigID,
	}

	if err := tx.Create(&branch).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("unable to create app branch: %w", err)
	}

	// Create queue for app branch
	queue, err := h.queueClient.Create(ctx, &queueclient.CreateQueueRequest{
		OwnerID:     branch.ID,
		OwnerType:   "AppBranch",
		Namespace:   "apps",
		MaxInFlight: 3,
		MaxDepth:    50,
	})
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("unable to create queue: %w", err)
	}

	// Update branch with queue ID
	branch.QueueID = queue.ID
	if err := tx.Save(&branch).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("unable to update app branch with queue: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("unable to commit transaction: %w", err)
	}

	return &branch, nil
}
