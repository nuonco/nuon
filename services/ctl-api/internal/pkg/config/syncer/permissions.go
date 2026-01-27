package syncer

import (
	"context"

	"github.com/nuonco/nuon/pkg/config/sync"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
)

// syncAppPermissions creates the app permissions configuration.
// Duplicates logic from services/ctl-api/internal/app/apps/service/create_app_permissions_config.go
func (s *syncer) syncAppPermissions(ctx context.Context) error {
	if s.cfg.Permissions == nil {
		return nil
	}

	obj := app.AppPermissionsConfig{
		AppID:       s.appID,
		AppConfigID: s.appConfigID,
		Roles:       []app.AppAWSIAMRoleConfig{},
	}

	// Add provision role
	if s.cfg.Permissions.ProvisionRole != nil {
		obj.Roles = append(obj.Roles, s.convertIAMRole(s.cfg.Permissions.ProvisionRole, app.AWSIAMRoleTypeRunnerProvision))
	}

	// Add maintenance role
	if s.cfg.Permissions.MaintenanceRole != nil {
		obj.Roles = append(obj.Roles, s.convertIAMRole(s.cfg.Permissions.MaintenanceRole, app.AWSIAMRoleTypeRunnerMaintenance))
	}

	// Add deprovision role
	if s.cfg.Permissions.DeprovisionRole != nil {
		obj.Roles = append(obj.Roles, s.convertIAMRole(s.cfg.Permissions.DeprovisionRole, app.AWSIAMRoleTypeRunnerDeprovision))
	}

	// Add break-glass roles if they exist
	if s.cfg.BreakGlass != nil && len(s.cfg.BreakGlass.Roles) > 0 {
		for _, role := range s.cfg.BreakGlass.Roles {
			obj.Roles = append(obj.Roles, s.convertIAMRole(role, app.AWSIAMRoleTypeBreakGlass))
		}
	}

	res := s.db.WithContext(ctx).Create(&obj)
	if res.Error != nil {
		return sync.SyncInternalErr{
			Description: "unable to create app permissions config",
			Err:         res.Error,
		}
	}

	return nil
}

func (s *syncer) convertIAMRole(role *config.AppAWSIAMRole, roleType app.AWSIAMRoleType) app.AppAWSIAMRoleConfig {
	policies := make([]app.AppAWSIAMPolicyConfig, 0, len(role.Policies))
	for _, policy := range role.Policies {
		policies = append(policies, app.AppAWSIAMPolicyConfig{
			AppConfigID:       s.appConfigID,
			ManagedPolicyName: policy.ManagedPolicyName,
			Name:              policy.Name,
			Contents:          generics.ToJSON(policy.Contents),
		})
	}

	return app.AppAWSIAMRoleConfig{
		AppConfigID:             s.appConfigID,
		Type:                    roleType,
		Name:                    role.Name,
		Description:             role.Description,
		DisplayName:             role.DisplayName,
		PermissionsBoundaryJSON: generics.ToJSON(role.PermissionsBoundary),
		Policies:                policies,
	}
}
