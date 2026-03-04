package policies

import (
	"context"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/sync"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// Sync creates the app policies configuration.
// Duplicates logic from services/ctl-api/internal/app/apps/service/create_app_policies_config.go
func Sync(ctx context.Context, db *gorm.DB, cfg *config.AppConfig, appID, appConfigID string) error {
	if cfg.Policies == nil {
		return nil
	}

	policyConfigs := make([]app.AppPolicyConfig, 0, len(cfg.Policies.Policies))
	for _, policy := range cfg.Policies.Policies {
		policyConfigs = append(policyConfigs, app.AppPolicyConfig{
			AppID:       appID,
			AppConfigID: appConfigID,
			Type:        config.AppPolicyType(policy.Type),
			Engine:      config.AppPolicyEngine(policy.Engine),
			Contents:    policy.Contents,
			Components:  policy.Components,
		})
	}

	obj := app.AppPoliciesConfig{
		AppID:       appID,
		AppConfigID: appConfigID,
		Policies:    policyConfigs,
	}

	res := db.WithContext(ctx).Create(&obj)
	if res.Error != nil {
		return sync.SyncInternalErr{
			Description: "unable to create app policies config",
			Err:         res.Error,
		}
	}

	return nil
}
