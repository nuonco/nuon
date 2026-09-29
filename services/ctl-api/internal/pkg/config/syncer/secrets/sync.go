package secrets

import (
	"context"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/sync"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/config/build"
)

func Sync(ctx context.Context, db *gorm.DB, cfg *config.AppConfig, appID, appConfigID string) error {
	if cfg.Secrets == nil {
		return nil
	}

	obj, err := build.SecretsConfig(build.SecretInputsFromConfig(cfg.Secrets), appID, appConfigID)
	if err != nil {
		return sync.SyncErr{
			Resource:    "app-secrets",
			Description: err.Error(),
			Err:         err,
		}
	}

	if res := db.WithContext(ctx).Create(obj); res.Error != nil {
		return sync.SyncInternalErr{
			Description: "unable to create app secrets config",
			Err:         res.Error,
		}
	}

	return nil
}
