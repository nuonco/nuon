package helpers

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

const (
	// DefaultAppBranchName is the branch `nuon apps sync` routes through when the org
	// has default-app-branches on; the CLI holds the same value in sync_branch.go.
	DefaultAppBranchName = "default"

	// DefaultAppBranchInstallGroupName names that branch's single all-installs group.
	DefaultAppBranchInstallGroupName = "all installs"
)

func (h *Helpers) CreateAppBranch(
	ctx context.Context,
	appID string,
	name string,
	opts ...app.AppBranchManagedBy,
) (*app.AppBranch, error) {
	branch, err := h.CreateAppBranchWithDB(ctx, h.db, appID, name, opts...)
	if err != nil {
		return nil, err
	}
	if err := h.EnsureAppBranchQueues(ctx, branch.ID); err != nil {
		return nil, err
	}

	return branch, nil
}

// Queue creation starts Temporal workflows, so a transactional caller must call
// EnsureAppBranchQueues itself once committed.
func (h *Helpers) CreateAppBranchWithDB(
	ctx context.Context,
	db *gorm.DB,
	appID string,
	name string,
	opts ...app.AppBranchManagedBy,
) (*app.AppBranch, error) {
	managedBy := app.AppBranchManagedByManually
	if len(opts) > 0 {
		managedBy = opts[0]
	}

	branch := app.AppBranch{
		AppID:     appID,
		Name:      name,
		ManagedBy: managedBy,
	}

	// Create branch first to get ID
	if err := db.WithContext(ctx).Create(&branch).Error; err != nil {
		return nil, fmt.Errorf("unable to create app branch: %w", err)
	}

	return &branch, nil
}

func (h *Helpers) EnsureAppBranchQueues(ctx context.Context, branchID string) error {
	ownerType := plugins.TableName(h.db, app.AppBranch{})

	// Create default queue for app branch signals
	_, err := h.queueClient.Create(ctx, &queueclient.CreateQueueRequest{
		OwnerID:     branchID,
		OwnerType:   ownerType,
		Namespace:   "apps",
		MaxInFlight: 2,
		MaxDepth:    50,
	})
	if err != nil {
		return fmt.Errorf("unable to create queue: %w", err)
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
		{AppBranchSandboxBuildsQueueName, 2},
	}
	for _, nq := range namedQueues {
		if _, err := h.queueClient.Create(ctx, &queueclient.CreateQueueRequest{
			OwnerID:     branchID,
			OwnerType:   ownerType,
			Namespace:   "apps",
			Name:        nq.name,
			MaxInFlight: nq.maxInFlight,
			MaxDepth:    50,
		}); err != nil {
			return fmt.Errorf("unable to create %s queue: %w", nq.name, err)
		}
	}

	return nil
}
