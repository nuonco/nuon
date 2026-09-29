package stack

import (
	"context"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/sync"
	appshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/config/build"
)

func Sync(ctx context.Context, db *gorm.DB, appsHelpers *appshelpers.Helpers, cfg *config.AppConfig, appID, appConfigID string) error {
	if cfg.Stack == nil {
		return nil
	}

	obj, err := build.StackConfig(cfg.Stack, appID, appConfigID)
	if err != nil {
		return sync.SyncErr{
			Resource:    "app-cloudformation-stack",
			Description: err.Error(),
			Err:         err,
		}
	}

	if res := db.WithContext(ctx).Create(obj); res.Error != nil {
		return sync.SyncInternalErr{
			Description: "unable to create app stack config",
			Err:         res.Error,
		}
	}

	if err := appsHelpers.UploadCustomNestedStackTemplates(ctx, db, obj); err != nil {
		return sync.SyncInternalErr{
			Description: "unable to upload custom nested stack templates",
			Err:         err,
		}
	}

	return nil
}
