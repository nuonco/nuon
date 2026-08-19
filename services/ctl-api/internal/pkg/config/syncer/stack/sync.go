package stack

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/sync"
	appshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/customstacks"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/config/build"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

// Sync creates the app stack configuration via the shared builder in
// internal/pkg/config/build, which the CreateAppStackConfig handler also uses.
func Sync(ctx context.Context, db *gorm.DB, appsHelpers *appshelpers.Helpers, cfg *config.AppConfig, appID, appConfigID string) error {
	if cfg.Stack == nil {
		return nil
	}

	obj, err := build.StackConfig(cfg.Stack, appID, appConfigID)
	if err != nil {
		return sync.SyncErr{
			Resource:    "app-cloudformation-stack",
			Description: err.Error(),
		}
	}

	if res := db.WithContext(ctx).Create(obj); res.Error != nil {
		return sync.SyncInternalErr{
			Description: "unable to create app stack config",
			Err:         res.Error,
		}
	}

	if len(obj.CustomNestedStacks) == 0 {
		return nil
	}

	// Templates upload to S3 asynchronously; until then each stays pending and
	// is skipped at stack generation.
	q, err := appsHelpers.QueueClient().GetQueueByOwner(ctx, appID, "apps")
	if err != nil {
		return sync.SyncInternalErr{
			Description: "unable to get apps queue for custom nested stacks",
			Err:         fmt.Errorf("unable to get apps queue for app %s: %w", appID, err),
		}
	}

	if _, err := appsHelpers.QueueClient().EnqueueSignalInTransaction(ctx, db, &queueclient.EnqueueSignalRequest{
		QueueID: q.ID,
		Signal: &customstacks.Signal{
			AppStackConfigID: obj.ID,
		},
	}); err != nil {
		return sync.SyncInternalErr{
			Description: "unable to enqueue custom stacks sync signal",
			Err:         err,
		}
	}

	return nil
}
