package stack

import (
	"context"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/sync"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// Sync creates the app CloudFormation stack configuration.
// Duplicates logic from services/ctl-api/internal/app/apps/service/create_app_stack_config.go
func Sync(ctx context.Context, db *gorm.DB, cfg *config.AppConfig, appID, appConfigID string) error {
	if cfg.Stack == nil {
		return nil
	}

	appCloudFormationStackConfig := app.AppStackConfig{
		Type:                    app.StackType(cfg.Stack.Type),
		AppConfigID:             appConfigID,
		AppID:                   appID,
		Name:                    cfg.Stack.Name,
		Description:             cfg.Stack.Description,
		VPCNestedTemplateURL:    cfg.Stack.VPCNestedTemplateURL,
		RunnerNestedTemplateURL: cfg.Stack.RunnerNestedTemplateURL,
	}

	res := db.WithContext(ctx).Create(&appCloudFormationStackConfig)
	if res.Error != nil {
		return sync.SyncInternalErr{
			Description: "unable to create app stack config",
			Err:         res.Error,
		}
	}

	return nil
}
