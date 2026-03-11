package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nuonco/nuon/pkg/render"
	"github.com/nuonco/nuon/pkg/types/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// InstallPermissionsRoleStatus embeds AppAWSIAMRoleConfig and adds provisioning status fields.
type InstallPermissionsRoleStatus struct {
	app.AppAWSIAMRoleConfig
	Enabled bool   `json:"enabled"`
	ARN     string `json:"arn"`
}

// InstallAppPermissionsConfigResponse is the response type for GetInstallAppPermissionsConfig.
type InstallAppPermissionsConfigResponse struct {
	ProvisionRole   *InstallPermissionsRoleStatus  `json:"provision_role"`
	DeprovisionRole *InstallPermissionsRoleStatus  `json:"deprovision_role"`
	MaintenanceRole *InstallPermissionsRoleStatus  `json:"maintenance_role"`
	BreakGlassRoles []InstallPermissionsRoleStatus `json:"break_glass_roles"`
	CustomRoles     []InstallPermissionsRoleStatus `json:"custom_roles"`
}

// @ID                    GetInstallAppPermissionsConfig
// @Summary               get app permissions config for an install with provisioning status
// @Description.markdown  get_install_app_permissions_config.md
// @Param                 install_id path string true "install ID"
// @Tags                  installs
// @Accept                json
// @Produce               json
// @Security              APIKey
// @Security              OrgID
// @Failure               400 {object} stderr.ErrResponse
// @Failure               401 {object} stderr.ErrResponse
// @Failure               403 {object} stderr.ErrResponse
// @Failure               404 {object} stderr.ErrResponse
// @Failure               500 {object} stderr.ErrResponse
// @Success               200 {object} InstallAppPermissionsConfigResponse
// @Router                /v1/installs/{install_id}/app-permissions-config [GET]
func (s *service) GetInstallAppPermissionsConfig(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	installID := ctx.Param("install_id")

	install, err := s.getInstall(ctx, installID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install: %w", err))
		return
	}

	appCfg, err := s.appsHelpers.GetFullAppConfig(ctx, install.AppConfigID, false)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get app config: %w", err))
		return
	}

	installState, err := s.helpers.GetInstallState(ctx, installID, false, false)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install state: %w", err))
		return
	}

	installStack, err := s.getInstallStack(ctx, installID, org.ID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install stack: %w", err))
		return
	}

	resp, err := s.buildInstallAppPermissionsConfig(appCfg, installStack, installState)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to build permissions config response: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

func (s *service) buildInstallAppPermissionsConfig(appCfg *app.AppConfig, installStack *app.InstallStack, installState *state.State) (*InstallAppPermissionsConfigResponse, error) {
	resp := &InstallAppPermissionsConfigResponse{
		BreakGlassRoles: []InstallPermissionsRoleStatus{},
		CustomRoles:     []InstallPermissionsRoleStatus{},
	}

	// If no stack outputs exist yet, return roles with Enabled: false and empty ARNs.
	if installStack == nil || installStack.InstallStackOutputs.AWSStackOutputs == nil {
		provisionRole := InstallPermissionsRoleStatus{
			AppAWSIAMRoleConfig: appCfg.PermissionsConfig.ProvisionRole,
			Enabled:             false,
			ARN:                 "",
		}
		resp.ProvisionRole = &provisionRole

		deprovisionRole := InstallPermissionsRoleStatus{
			AppAWSIAMRoleConfig: appCfg.PermissionsConfig.DeprovisionRole,
			Enabled:             false,
			ARN:                 "",
		}
		resp.DeprovisionRole = &deprovisionRole

		maintenanceRole := InstallPermissionsRoleStatus{
			AppAWSIAMRoleConfig: appCfg.PermissionsConfig.MaintenanceRole,
			Enabled:             false,
			ARN:                 "",
		}
		resp.MaintenanceRole = &maintenanceRole

		for _, role := range appCfg.BreakGlassConfig.Roles {
			resp.BreakGlassRoles = append(resp.BreakGlassRoles, InstallPermissionsRoleStatus{
				AppAWSIAMRoleConfig: role,
				Enabled:             false,
				ARN:                 "",
			})
		}

		for _, role := range appCfg.PermissionsConfig.CustomRoles {
			resp.CustomRoles = append(resp.CustomRoles, InstallPermissionsRoleStatus{
				AppAWSIAMRoleConfig: role,
				Enabled:             false,
				ARN:                 "",
			})
		}

		return resp, nil
	}

	awsOutputs := installStack.InstallStackOutputs.AWSStackOutputs

	stateMap, err := installState.AsMap()
	if err != nil {
		return nil, fmt.Errorf("unable to get install state map: %w", err)
	}

	// Provision role
	{
		role := appCfg.PermissionsConfig.ProvisionRole
		arn := awsOutputs.ProvisionIAMRoleARN
		resp.ProvisionRole = &InstallPermissionsRoleStatus{
			AppAWSIAMRoleConfig: role,
			Enabled:             arn != "",
			ARN:                 arn,
		}
	}

	// Deprovision role
	{
		role := appCfg.PermissionsConfig.DeprovisionRole
		arn := awsOutputs.DeprovisionIAMRoleARN
		resp.DeprovisionRole = &InstallPermissionsRoleStatus{
			AppAWSIAMRoleConfig: role,
			Enabled:             arn != "",
			ARN:                 arn,
		}
	}

	// Maintenance role
	{
		role := appCfg.PermissionsConfig.MaintenanceRole
		arn := awsOutputs.MaintenanceIAMRoleARN
		resp.MaintenanceRole = &InstallPermissionsRoleStatus{
			AppAWSIAMRoleConfig: role,
			Enabled:             arn != "",
			ARN:                 arn,
		}
	}

	// Break glass roles (array, keyed by rendered name in stack outputs)
	for _, role := range appCfg.BreakGlassConfig.Roles {
		rendered, err := render.RenderV2(role.Name, stateMap)
		if err != nil {
			return nil, fmt.Errorf("unable to render break glass role name: %w", err)
		}
		arn := awsOutputs.BreakGlassRoleARNs[rendered]
		resp.BreakGlassRoles = append(resp.BreakGlassRoles, InstallPermissionsRoleStatus{
			AppAWSIAMRoleConfig: role,
			Enabled:             arn != "",
			ARN:                 arn,
		})
	}

	// Custom roles (array, keyed by rendered name in stack outputs)
	for _, role := range appCfg.PermissionsConfig.CustomRoles {
		rendered, err := render.RenderV2(role.Name, stateMap)
		if err != nil {
			return nil, fmt.Errorf("unable to render custom role name: %w", err)
		}
		arn := awsOutputs.CustomRoleARNs[rendered]
		resp.CustomRoles = append(resp.CustomRoles, InstallPermissionsRoleStatus{
			AppAWSIAMRoleConfig: role,
			Enabled:             arn != "",
			ARN:                 arn,
		})
	}

	return resp, nil
}
