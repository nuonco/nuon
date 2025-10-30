package service

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/stderr"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/generics"
	validatorPkg "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/validator"
)

type CreateAppPermissionsConfigRequest struct {
	ProvisionRole   AppAWSIAMRoleConfig `json:"provision_role" validate:"required"`
	DeprovisionRole AppAWSIAMRoleConfig `json:"deprovision_role" validate:"required"`
	MaintenanceRole AppAWSIAMRoleConfig `json:"maintenance_role" validate:"required"`

	BreakGlassRoles *[]AppAWSIAMRoleConfig `json:"break_glass_roles"`

	AppConfigID string `json:"app_config_id" validate:"required"`
}

type AppAWSIAMRoleConfig struct {
	Name                string `json:"name" validate:"required"`
	DisplayName         string `json:"display_name" validate:"required"`
	Description         string `json:"description" validate:"required"`
	PermissionsBoundary string `json:"permissions_boundary,omitempty" swaggertype:"string" validate:"optional_json"`

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
		})
	}

	return policies
}

type AppAWSIAMPolicyConfig struct {
	ManagedPolicyName string `json:"managed_policy_name"`

	Name     string `json:"name"`
	Contents string `json:"contents" swaggertype:"string" validate:"optional_json"`
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
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{object}	app.AppPermissionsConfig
// @Router /v1/apps/{app_id}/permissions-configs [post]
func (s *service) CreateAppPermissionsConfig(ctx *gin.Context) {
	appID := ctx.Param("app_id")

	var req CreateAppPermissionsConfigRequest
	if err := ctx.BindJSON(&req); err != nil {
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

func (s *service) getBreakGlassRoleConfigs(roles []AppAWSIAMRoleConfig, appConfigID string) []app.AppAWSIAMRoleConfig {
	roleConfigs := make([]app.AppAWSIAMRoleConfig, 0, len(roles))

	for _, role := range roles {
		roleConfig := app.AppAWSIAMRoleConfig{
			AppConfigID:             appConfigID,
			Type:                    app.AWSIAMRoleTypeBreakGlass,
			Name:                    role.Name,
			Description:             role.Description,
			DisplayName:             role.DisplayName,
			PermissionsBoundaryJSON: generics.ToJSON(role.PermissionsBoundary),
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
				Type:                    app.AWSIAMRoleTypeRunnerProvision,
				Name:                    req.ProvisionRole.Name,
				Description:             req.ProvisionRole.Description,
				DisplayName:             req.ProvisionRole.DisplayName,
				PermissionsBoundaryJSON: generics.ToJSON(req.ProvisionRole.PermissionsBoundary),
				Policies:                req.ProvisionRole.getPolicies(req.AppConfigID),
			},
			{
				AppConfigID:             req.AppConfigID,
				Type:                    app.AWSIAMRoleTypeRunnerMaintenance,
				Name:                    req.MaintenanceRole.Name,
				Description:             req.MaintenanceRole.Description,
				DisplayName:             req.MaintenanceRole.DisplayName,
				PermissionsBoundaryJSON: generics.ToJSON(req.MaintenanceRole.PermissionsBoundary),
				Policies:                req.MaintenanceRole.getPolicies(req.AppConfigID),
			},
			{
				AppConfigID:             req.AppConfigID,
				Type:                    app.AWSIAMRoleTypeRunnerDeprovision,
				Name:                    req.DeprovisionRole.Name,
				Description:             req.DeprovisionRole.Description,
				DisplayName:             req.DeprovisionRole.DisplayName,
				PermissionsBoundaryJSON: generics.ToJSON(req.DeprovisionRole.PermissionsBoundary),
				Policies:                req.DeprovisionRole.getPolicies(req.AppConfigID),
			},
		},
	}
	if req.BreakGlassRoles != nil {
		obj.Roles = append(obj.Roles, s.getBreakGlassRoleConfigs(*req.BreakGlassRoles, req.AppConfigID)...)
	}

	res := s.db.WithContext(ctx).Create(&obj)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to create app permissions config")
	}

	return &obj, nil
}
