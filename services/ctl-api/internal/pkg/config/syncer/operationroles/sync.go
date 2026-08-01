package operationroles

import (
	"context"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/sync"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/config/build"
)

// Sync creates the app operation role configuration via the shared builders in
// internal/pkg/config/build, which the CreateAppOperationRoleConfig handler also
// uses.
func Sync(ctx context.Context, db *gorm.DB, cfg *config.AppConfig, appID, appConfigID string) error {
	if cfg.OperationRoles == nil || len(cfg.OperationRoles.RuleMatrix) == 0 {
		return nil
	}

	opRoleConfig := build.OperationRoleConfig(appID, appConfigID)
	if err := db.WithContext(ctx).Create(opRoleConfig).Error; err != nil {
		return sync.SyncInternalErr{
			Description: "unable to create operation role config",
			Err:         err,
		}
	}

	rules, err := build.OperationRoleRules(build.OperationRoleRuleInputsFromConfig(cfg.OperationRoles), opRoleConfig.ID)
	if err != nil {
		return sync.SyncErr{
			Resource:    "app-operations-roles",
			Description: err.Error(),
		}
	}

	if err := db.WithContext(ctx).Create(&rules).Error; err != nil {
		return sync.SyncInternalErr{
			Description: "unable to create operation role rules",
			Err:         err,
		}
	}

	return nil
}
