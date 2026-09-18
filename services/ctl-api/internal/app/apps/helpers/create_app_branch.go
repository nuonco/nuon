package helpers

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/queuenames"
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

type EnsureAppBranchQueuesOptions struct {
	SkipRestartHint bool
}

func (h *Helpers) EnsureAppBranchQueues(ctx context.Context, branchID string, opts ...EnsureAppBranchQueuesOptions) error {
	ownerType := plugins.TableName(h.db, app.AppBranch{})
	var options EnsureAppBranchQueuesOptions
	if len(opts) > 0 {
		options = opts[0]
	}

	specs, ok := queuenames.Specs(queuenames.OwnerAppBranches)
	if !ok {
		return fmt.Errorf("app branch queue specs are not registered")
	}

	for _, spec := range specs {
		_, err := h.queueClient.Create(ctx, &queueclient.CreateQueueRequest{
			OwnerID:         branchID,
			OwnerType:       ownerType,
			Namespace:       "apps",
			Name:            spec.Name,
			MaxInFlight:     spec.MaxInFlight,
			MaxDepth:        spec.MaxDepth,
			SkipRestartHint: options.SkipRestartHint,
		})
		if err != nil {
			return fmt.Errorf("unable to create %s queue: %w", spec.Name, err)
		}
	}

	return nil
}
