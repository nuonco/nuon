package permissions

import (
	"context"
	"encoding/json"
	"fmt"

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

	if err := validateInlinePolicyContents(obj.Roles); err != nil {
		return sync.SyncErr{
			Resource:    "permissions",
			Description: err.Error(),
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
			GCPPermissions:    policy.GCPPermissions,
			GCPPredefinedRole: policy.GCPPredefinedRole,
		})
	}

	return app.AppAWSIAMRoleConfig{
		AppConfigID:             appConfigID,
		CloudPlatform:           role.CloudPlatform,
		Type:                    roleType,
		Name:                    role.Name,
		Description:             role.Description,
		DisplayName:             role.DisplayName,
		PermissionsBoundaryJSON: generics.ToJSON(role.PermissionsBoundary),
		Policies:                policies,
	}
}

// validateInlinePolicyContents ensures every inline policy on every role parses
// as a well-formed IAM policy document. Without this, malformed inline policies
// would silently render to an empty inline policy on the AWS Terraform install
// path, leaving runner roles under-permissioned at apply time. Catch the
// problem at sync rather than at provision.
func validateInlinePolicyContents(roles []app.AppAWSIAMRoleConfig) error {
	for _, role := range roles {
		for _, policy := range role.Policies {
			if len(policy.Contents) == 0 {
				continue
			}
			if policy.ManagedPolicyName != "" {
				continue
			}
			var doc struct {
				Statement []map[string]json.RawMessage `json:"Statement"`
			}
			if err := json.Unmarshal(policy.Contents, &doc); err != nil {
				return fmt.Errorf("role %q policy %q: inline policy contents must be a JSON IAM policy document: %w", role.Name, policy.Name, err)
			}
			if len(doc.Statement) == 0 {
				return fmt.Errorf("role %q policy %q: inline policy must contain at least one Statement", role.Name, policy.Name)
			}
			for i, stmt := range doc.Statement {
				if _, ok := stmt["Effect"]; !ok {
					return fmt.Errorf("role %q policy %q: Statement[%d] missing required Effect field", role.Name, policy.Name, i)
				}
			}
		}
	}
	return nil
}
