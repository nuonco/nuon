package operationroles

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/sync"
	"github.com/nuonco/nuon/pkg/principal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func Sync(ctx context.Context, db *gorm.DB, cfg *config.AppConfig, appID, appConfigID string) error {
	if cfg.OperationRoles == nil || len(cfg.OperationRoles.RuleMatrix) == 0 {
		return nil
	}

	opRoleConfig := app.AppOperationRoleConfig{
		AppID:       appID,
		AppConfigID: appConfigID,
	}

	if err := db.WithContext(ctx).Create(&opRoleConfig).Error; err != nil {
		return sync.SyncInternalErr{
			Description: "unable to create operation role config",
			Err:         err,
		}
	}

	rules := make([]app.AppOperationRoleRule, 0, len(cfg.OperationRoles.RuleMatrix))
	for _, rule := range cfg.OperationRoles.RuleMatrix {
		p, err := principal.ParsePrincipal(rule.Principal)
		if err != nil {
			return sync.SyncErr{
				Resource:    "app-operations-roles",
				Description: fmt.Sprintf("invalid principal %q: %v", rule.Principal, err),
			}
		}

		rules = append(rules, app.AppOperationRoleRule{
			AppOperationRoleConfigID: opRoleConfig.ID,
			PrincipalType:            p.Type,
			PrincipalName:            p.Name,
			Operation:                app.OperationType(rule.Operation),
			Role:                     rule.RoleName,
		})
	}

	if err := db.WithContext(ctx).Create(&rules).Error; err != nil {
		return sync.SyncInternalErr{
			Description: "unable to create operation role rules",
			Err:         err,
		}
	}

	return nil
}
