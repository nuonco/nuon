package build

import (
	"errors"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type PermissionsInput struct {
	AppID       string
	AppConfigID string

	Permissions *config.PermissionsConfig

	// Also written onto the permissions config so the rows are reachable from
	// both owners, matching the CLI sync path.
	BreakGlassRoles []*config.AppAWSIAMRole
}

func PermissionsConfig(in PermissionsInput) (*app.AppPermissionsConfig, error) {
	if in.Permissions == nil {
		return nil, errors.New("permissions config is required")
	}

	obj := &app.AppPermissionsConfig{
		AppID:       in.AppID,
		AppConfigID: in.AppConfigID,
		Roles:       make([]app.AppAWSIAMRoleConfig, 0),
	}

	standard := []struct {
		role     *config.AppAWSIAMRole
		roleType app.AWSIAMRoleType
	}{
		{in.Permissions.ProvisionRole, app.AWSIAMRoleTypeRunnerProvision},
		{in.Permissions.MaintenanceRole, app.AWSIAMRoleTypeRunnerMaintenance},
		{in.Permissions.DeprovisionRole, app.AWSIAMRoleTypeRunnerDeprovision},
	}
	for _, entry := range standard {
		if entry.role == nil {
			continue
		}
		obj.Roles = append(obj.Roles, IAMRole(entry.role, in.AppConfigID, entry.roleType))
	}

	obj.Roles = append(obj.Roles, IAMRoles(in.BreakGlassRoles, in.AppConfigID, app.AWSIAMRoleTypeBreakGlass)...)
	obj.Roles = append(obj.Roles, IAMRoles(in.Permissions.CustomRoles, in.AppConfigID, app.AWSIAMRoleTypeCustom)...)

	if err := validatePermissionRoles(in); err != nil {
		return nil, err
	}
	if err := ValidateInlinePolicyContents(obj.Roles); err != nil {
		return nil, err
	}

	return obj, nil
}

func BreakGlassConfig(appID, appConfigID string, roles []*config.AppAWSIAMRole) (*app.AppBreakGlassConfig, error) {
	obj := &app.AppBreakGlassConfig{
		AppID:       appID,
		AppConfigID: appConfigID,
		Roles:       IAMRoles(roles, appConfigID, app.AWSIAMRoleTypeBreakGlass),
	}

	if err := ValidateInlinePolicyContents(obj.Roles); err != nil {
		return nil, err
	}

	return obj, nil
}

func validatePermissionRoles(in PermissionsInput) error {
	named := []struct {
		name string
		role *config.AppAWSIAMRole
	}{
		{"provision_role", in.Permissions.ProvisionRole},
		{"deprovision_role", in.Permissions.DeprovisionRole},
		{"maintenance_role", in.Permissions.MaintenanceRole},
	}
	for _, role := range in.BreakGlassRoles {
		if role != nil {
			named = append(named, struct {
				name string
				role *config.AppAWSIAMRole
			}{role.Name, role})
		}
	}
	for _, role := range in.Permissions.CustomRoles {
		if role != nil {
			named = append(named, struct {
				name string
				role *config.AppAWSIAMRole
			}{role.Name, role})
		}
	}

	for _, entry := range named {
		if entry.role == nil {
			continue
		}
		if err := ValidateAzureBuiltInRoles(entry.name, entry.role.Policies); err != nil {
			return err
		}
		if err := ValidatePolicyMutualExclusivity(entry.name, entry.role.Policies); err != nil {
			return err
		}
	}

	return nil
}
