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

	ownerType := plugins.TableName(h.db, app.AppBranch{})

	// Create default queue for app branch signals
	_, err := h.queueClient.Create(ctx, &queueclient.CreateQueueRequest{
		OwnerID:     branch.ID,
		OwnerType:   ownerType,
		Namespace:   "apps",
		MaxInFlight: 2,
		MaxDepth:    50,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create queue: %w", err)
	}

	// Create named queues for workflow execution pipeline
	namedQueues := []struct {
		name        string
		maxInFlight int
	}{
		{"app-branch-signals", 5},
		{"app-branch-workflow-step-groups", 2},
		{"app-branch-workflow-steps", 5},
		{"app-branch-generate-steps", 2},
	}
	for _, nq := range namedQueues {
		if _, err := h.queueClient.Create(ctx, &queueclient.CreateQueueRequest{
			OwnerID:     branch.ID,
			OwnerType:   ownerType,
			Namespace:   "apps",
			Name:        nq.name,
			MaxInFlight: nq.maxInFlight,
			MaxDepth:    50,
		}); err != nil {
			return nil, fmt.Errorf("unable to create %s queue: %w", nq.name, err)
		}
	}

	return &branch, nil
}
