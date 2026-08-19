package service

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/config/build"
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

type AppAWSIAMPolicyConfig struct {
	ManagedPolicyName string `json:"managed_policy_name"`

	Name              string   `json:"name"`
	Contents          string   `json:"contents" swaggertype:"string" validate:"optional_json"`
	GCPPermissions    []string `json:"gcp_permissions,omitempty"`
	GCPPredefinedRole string   `json:"gcp_predefined_role,omitempty"`
	AzureActions      []string `json:"azure_actions,omitempty"`
	AzureBuiltInRoles []string `json:"azure_built_in_roles,omitempty"`
}

func (a AppAWSIAMRoleConfig) toConfig() *config.AppAWSIAMRole {
	policies := make([]config.AppAWSIAMPolicy, 0, len(a.Policies))
	for _, policy := range a.Policies {
		policies = append(policies, config.AppAWSIAMPolicy{
			ManagedPolicyName: policy.ManagedPolicyName,
			Name:              policy.Name,
			Contents:          policy.Contents,
			GCPPermissions:    policy.GCPPermissions,
			GCPPredefinedRole: policy.GCPPredefinedRole,
			AzureActions:      policy.AzureActions,
			AzureBuiltInRoles: policy.AzureBuiltInRoles,
		})
	}

	return &config.AppAWSIAMRole{
		CloudPlatform:       a.CloudPlatform,
		Name:                a.Name,
		Description:         a.Description,
		DisplayName:         a.DisplayName,
		PermissionsBoundary: a.PermissionsBoundary,
		EnabledInStack:      a.EnabledInStack,
		Policies:            policies,
	}
}

func toConfigRoles(roles *[]AppAWSIAMRoleConfig) []*config.AppAWSIAMRole {
	if roles == nil {
		return nil
	}
	out := make([]*config.AppAWSIAMRole, 0, len(*roles))
	for _, role := range *roles {
		out = append(out, role.toConfig())
	}
	return out
}

func (c *CreateAppPermissionsConfigRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return validatorPkg.FormatValidationError(err)
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

func (s *service) createAppPermissionsConfig(ctx context.Context, appID string, req *CreateAppPermissionsConfigRequest) (*app.AppPermissionsConfig, error) {
	obj, err := build.PermissionsConfig(build.PermissionsInput{
		AppID:       appID,
		AppConfigID: req.AppConfigID,
		Permissions: &config.PermissionsConfig{
			ProvisionRole:   req.ProvisionRole.toConfig(),
			DeprovisionRole: req.DeprovisionRole.toConfig(),
			MaintenanceRole: req.MaintenanceRole.toConfig(),
			CustomRoles:     toConfigRoles(req.CustomRoles),
		},
		BreakGlassRoles: toConfigRoles(req.BreakGlassRoles),
	})
	if err != nil {
		return nil, stderr.NewInvalidRequest(err)
	}

	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := s.db.WithContext(ctx).Create(obj)
		if res.Error != nil {
			return errors.Wrap(res.Error, "unable to create app permissions config")
		}
		if err := s.installsHelpers.MigrateInstallRoles(ctx, tx, appID, *obj); err != nil {
			return errors.Wrap(err, "failed to migrate install roles")
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return obj, nil
}
