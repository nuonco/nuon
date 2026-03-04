package permissions

import (
	"context"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/sync"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
)

// Sync creates the app permissions configuration.
// Duplicates logic from services/ctl-api/internal/app/apps/service/create_app_permissions_config.go
func Sync(ctx context.Context, db *gorm.DB, cfg *config.AppConfig, appID, appConfigID string) error {
	if cfg.Permissions == nil {
		return nil
	}

	obj := app.AppPermissionsConfig{
		AppID:       appID,
		AppConfigID: appConfigID,
		Roles:       []app.AppAWSIAMRoleConfig{},
	}

	// Add provision role
	if cfg.Permissions.ProvisionRole != nil {
		obj.Roles = append(obj.Roles, convertIAMRole(cfg.Permissions.ProvisionRole, appConfigID, app.AWSIAMRoleTypeRunnerProvision))
	}

	// Add maintenance role
	if cfg.Permissions.MaintenanceRole != nil {
		obj.Roles = append(obj.Roles, convertIAMRole(cfg.Permissions.MaintenanceRole, appConfigID, app.AWSIAMRoleTypeRunnerMaintenance))
	}

	// Add deprovision role
	if cfg.Permissions.DeprovisionRole != nil {
		obj.Roles = append(obj.Roles, convertIAMRole(cfg.Permissions.DeprovisionRole, appConfigID, app.AWSIAMRoleTypeRunnerDeprovision))
	}

	// Add break-glass roles if they exist
	if cfg.BreakGlass != nil && len(cfg.BreakGlass.Roles) > 0 {
		for _, role := range cfg.BreakGlass.Roles {
			obj.Roles = append(obj.Roles, convertIAMRole(role, appConfigID, app.AWSIAMRoleTypeBreakGlass))
		}
	}

	res := db.WithContext(ctx).Create(&obj)
	if res.Error != nil {
		return sync.SyncInternalErr{
			Description: "unable to create app permissions config",
			Err:         res.Error,
		}
	}

	return nil
}

func convertIAMRole(role *config.AppAWSIAMRole, appConfigID string, roleType app.AWSIAMRoleType) app.AppAWSIAMRoleConfig {
	policies := make([]app.AppAWSIAMPolicyConfig, 0, len(role.Policies))
	for _, policy := range role.Policies {
		policies = append(policies, app.AppAWSIAMPolicyConfig{
			AppConfigID:       appConfigID,
			ManagedPolicyName: policy.ManagedPolicyName,
			Name:              policy.Name,
			Contents:          generics.ToJSON(policy.Contents),
		})
	}

	return app.AppAWSIAMRoleConfig{
		AppConfigID:             appConfigID,
		Type:                    roleType,
		Name:                    role.Name,
		Description:             role.Description,
		DisplayName:             role.DisplayName,
		PermissionsBoundaryJSON: generics.ToJSON(role.PermissionsBoundary),
		Policies:                policies,
	}
}
