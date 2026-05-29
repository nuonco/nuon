package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	pkggenerics "github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
	validatorPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/validator"
)

type CreateAppPermissionsConfigRequest struct {
	ProvisionRole   AppAWSIAMRoleConfig `json:"provision_role" validate:"required"`
	DeprovisionRole AppAWSIAMRoleConfig `json:"deprovision_role" validate:"required"`
	MaintenanceRole AppAWSIAMRoleConfig `json:"maintenance_role" validate:"required"`

	BreakGlassRoles *[]AppAWSIAMRoleConfig `json:"break_glass_roles"`
	CustomRoles     *[]AppAWSIAMRoleConfig `json:"custom_roles"`

	AppConfigID string `json:"app_config_id" validate:"required"`
}

type AppAWSIAMRoleConfig struct {
	Name                string `json:"name" validate:"required"`
	DisplayName         string `json:"display_name" validate:"required"`
	Description         string `json:"description" validate:"required"`
	PermissionsBoundary string `json:"permissions_boundary,omitempty" swaggertype:"string" validate:"optional_json"`
	CloudPlatform       string `json:"cloud_platform,omitempty" validate:"omitempty,oneof=aws gcp azure"`
	EnabledInStack      *bool  `json:"enabled_in_stack" swaggertype:"boolean" extensions:"x-nullable"`

	Policies []AppAWSIAMPolicyConfig `json:"policies" validate:"min=1,dive"`
}

func (a AppAWSIAMRoleConfig) getPolicies(appConfigID string) []app.AppAWSIAMPolicyConfig {
	policies := make([]app.AppAWSIAMPolicyConfig, 0, len(a.Policies))

	for _, policy := range a.Policies {
		policies = append(policies, app.AppAWSIAMPolicyConfig{
			AppConfigID:       appConfigID,
			ManagedPolicyName: policy.ManagedPolicyName,
			Name:              policy.Name,
			Contents:          generics.ToJSON(policy.Contents),
			GCPPermissions:    policy.GCPPermissions,
			GCPPredefinedRole: policy.GCPPredefinedRole,
		})
	}

	return policies
}

type AppAWSIAMPolicyConfig struct {
	ManagedPolicyName string `json:"managed_policy_name"`

	Name              string   `json:"name"`
	Contents          string   `json:"contents" swaggertype:"string" validate:"optional_json"`
	GCPPermissions    []string `json:"gcp_permissions,omitempty"`
	GCPPredefinedRole string   `json:"gcp_predefined_role,omitempty"`
}

func (p *AppAWSIAMPolicyConfig) validateMutualExclusivity(roleName string) error {
	if p.Contents != "" && p.ManagedPolicyName != "" {
		return fmt.Errorf("role %q policy %q: contents and managed_policy_name are mutually exclusive; specify one or the other", roleName, p.Name)
	}

	if len(p.GCPPermissions) > 0 && p.GCPPredefinedRole != "" {
		return fmt.Errorf("role %q policy %q: gcp_permissions and gcp_predefined_role are mutually exclusive; use gcp_permissions for fine-grained custom permissions or gcp_predefined_role for a Google-managed role, not both", roleName, p.Name)
	}

	return nil
}

func (c *CreateAppPermissionsConfigRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return validatorPkg.FormatValidationError(err)
	}

	allRoles := []struct {
		name string
		role AppAWSIAMRoleConfig
	}{
		{"provision_role", c.ProvisionRole},
		{"deprovision_role", c.DeprovisionRole},
		{"maintenance_role", c.MaintenanceRole},
	}
	if c.BreakGlassRoles != nil {
		for _, r := range *c.BreakGlassRoles {
			allRoles = append(allRoles, struct {
				name string
				role AppAWSIAMRoleConfig
			}{r.Name, r})
		}
	}
	if c.CustomRoles != nil {
		for _, r := range *c.CustomRoles {
			allRoles = append(allRoles, struct {
				name string
				role AppAWSIAMRoleConfig
			}{r.Name, r})
		}
	}

	for _, entry := range allRoles {
		for i := range entry.role.Policies {
			if err := entry.role.Policies[i].validateMutualExclusivity(entry.name); err != nil {
				return err
			}
		}
	}

	return nil
}

// @ID						CreateAppPermissionsConfig
// @Description.markdown	create_app_permissions_config.md
// @Tags					apps
// @Accept					json
// @Param					req	body	CreateAppPermissionsConfigRequest	true	"Input"
// @Produce				json
// @Param					app_id	path	string	true	"app ID"
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				409	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{object}	app.AppPermissionsConfig
// @Router /v1/apps/{app_id}/permissions-configs [post]
func (s *service) CreateAppPermissionsConfig(ctx *gin.Context) {
	appID := ctx.Param("app_id")

	var req CreateAppPermissionsConfigRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.ErrInvalidRequest{
			Err: err,
		})
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(err)
		return
	}

	cfg, err := s.createAppPermissionsConfig(ctx, appID, &req)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusCreated, cfg)
}

func (s *service) getCustomRoleConfigs(roles []AppAWSIAMRoleConfig, appConfigID string) []app.AppAWSIAMRoleConfig {
	roleConfigs := make([]app.AppAWSIAMRoleConfig, 0, len(roles))

	for _, role := range roles {
		roleConfig := app.AppAWSIAMRoleConfig{
			AppConfigID:             appConfigID,
			CloudPlatform:           role.CloudPlatform,
			Type:                    app.AWSIAMRoleTypeCustom,
			Name:                    role.Name,
			Description:             role.Description,
			DisplayName:             role.DisplayName,
			PermissionsBoundaryJSON: generics.ToJSON(role.PermissionsBoundary),
			EnabledInStack:          pkggenerics.NewNullBoolFromPtr(role.EnabledInStack),
			Policies:                role.getPolicies(appConfigID),
		}
		roleConfigs = append(roleConfigs, roleConfig)
	}

	return roleConfigs
}

func (s *service) getBreakGlassRoleConfigs(roles []AppAWSIAMRoleConfig, appConfigID string) []app.AppAWSIAMRoleConfig {
	roleConfigs := make([]app.AppAWSIAMRoleConfig, 0, len(roles))

	for _, role := range roles {
		roleConfig := app.AppAWSIAMRoleConfig{
			AppConfigID:             appConfigID,
			CloudPlatform:           role.CloudPlatform,
			Type:                    app.AWSIAMRoleTypeBreakGlass,
			Name:                    role.Name,
			Description:             role.Description,
			DisplayName:             role.DisplayName,
			PermissionsBoundaryJSON: generics.ToJSON(role.PermissionsBoundary),
			EnabledInStack:          pkggenerics.NewNullBoolFromPtr(role.EnabledInStack),
			Policies:                role.getPolicies(appConfigID),
		}
		roleConfigs = append(roleConfigs, roleConfig)
	}

	return roleConfigs
}

func (s *service) createAppPermissionsConfig(ctx context.Context, appID string, req *CreateAppPermissionsConfigRequest) (*app.AppPermissionsConfig, error) {
	obj := app.AppPermissionsConfig{
		AppID:       appID,
		AppConfigID: req.AppConfigID,
		Roles: []app.AppAWSIAMRoleConfig{
			{
				AppConfigID:             req.AppConfigID,
				CloudPlatform:           req.ProvisionRole.CloudPlatform,
				Type:                    app.AWSIAMRoleTypeRunnerProvision,
				Name:                    req.ProvisionRole.Name,
				Description:             req.ProvisionRole.Description,
				DisplayName:             req.ProvisionRole.DisplayName,
				PermissionsBoundaryJSON: generics.ToJSON(req.ProvisionRole.PermissionsBoundary),
				EnabledInStack:          pkggenerics.NewNullBoolFromPtr(req.ProvisionRole.EnabledInStack),
				Policies:                req.ProvisionRole.getPolicies(req.AppConfigID),
			},
			{
				AppConfigID:             req.AppConfigID,
				CloudPlatform:           req.MaintenanceRole.CloudPlatform,
				Type:                    app.AWSIAMRoleTypeRunnerMaintenance,
				Name:                    req.MaintenanceRole.Name,
				Description:             req.MaintenanceRole.Description,
				DisplayName:             req.MaintenanceRole.DisplayName,
				PermissionsBoundaryJSON: generics.ToJSON(req.MaintenanceRole.PermissionsBoundary),
				EnabledInStack:          pkggenerics.NewNullBoolFromPtr(req.MaintenanceRole.EnabledInStack),
				Policies:                req.MaintenanceRole.getPolicies(req.AppConfigID),
			},
			{
				AppConfigID:             req.AppConfigID,
				CloudPlatform:           req.DeprovisionRole.CloudPlatform,
				Type:                    app.AWSIAMRoleTypeRunnerDeprovision,
				Name:                    req.DeprovisionRole.Name,
				Description:             req.DeprovisionRole.Description,
				DisplayName:             req.DeprovisionRole.DisplayName,
				PermissionsBoundaryJSON: generics.ToJSON(req.DeprovisionRole.PermissionsBoundary),
				EnabledInStack:          pkggenerics.NewNullBoolFromPtr(req.DeprovisionRole.EnabledInStack),
				Policies:                req.DeprovisionRole.getPolicies(req.AppConfigID),
			},
		},
	}
	if req.BreakGlassRoles != nil {
		obj.Roles = append(obj.Roles, s.getBreakGlassRoleConfigs(*req.BreakGlassRoles, req.AppConfigID)...)
	}

	if req.CustomRoles != nil {
		obj.Roles = append(obj.Roles, s.getCustomRoleConfigs(*req.CustomRoles, req.AppConfigID)...)
	}

	res := s.db.WithContext(ctx).Create(&obj)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to create app permissions config")
	}

	tx := s.db.WithContext(ctx).Begin()
	if err := s.installsHelpers.MigrateInstallRoles(ctx, tx, appID, obj); err != nil {
		tx.Rollback()
		s.l.Warn("failed to migrate install roles", zap.Error(err))
	} else {
		tx.Commit()
	}

	return &obj, nil
}
