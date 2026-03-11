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

// renderRole renders the template fields (Name, DisplayName, Description) on a role config using the install state.
func renderRole(role app.AppAWSIAMRoleConfig, stateMap map[string]any) (app.AppAWSIAMRoleConfig, error) {
	rendered, err := render.RenderV2(role.Name, stateMap)
	if err != nil {
		return role, fmt.Errorf("unable to render role name: %w", err)
	}
	role.Name = rendered

	rendered, err = render.RenderV2(role.DisplayName, stateMap)
	if err != nil {
		return role, fmt.Errorf("unable to render role display name: %w", err)
	}
	role.DisplayName = rendered

	rendered, err = render.RenderV2(role.Description, stateMap)
	if err != nil {
		return role, fmt.Errorf("unable to render role description: %w", err)
	}
	role.Description = rendered

	return role, nil
}

func (s *service) buildInstallAppPermissionsConfig(appCfg *app.AppConfig, installStack *app.InstallStack, installState *state.State) (*InstallAppPermissionsConfigResponse, error) {
	resp := &InstallAppPermissionsConfigResponse{
		BreakGlassRoles: []InstallPermissionsRoleStatus{},
		CustomRoles:     []InstallPermissionsRoleStatus{},
	}

	// Resolve the StackOutput interface from whichever cloud provider is present.
	var stackOutput app.StackOutput
	if installStack != nil {
		outputs := installStack.InstallStackOutputs
		switch {
		case outputs.AWSStackOutputs != nil:
			stackOutput = outputs.AWSStackOutputs
		case outputs.GCPStackOutputs != nil:
			stackOutput = outputs.GCPStackOutputs
		case outputs.AzureStackOutputs != nil:
			stackOutput = outputs.AzureStackOutputs
		}
	}

	stateMap, err := installState.AsMap()
	if err != nil {
		return nil, fmt.Errorf("unable to get install state map: %w", err)
	}

	// If no stack outputs exist yet, return roles with rendered names but Enabled: false.
	if stackOutput == nil {
		provisionRole, err := renderRole(appCfg.PermissionsConfig.ProvisionRole, stateMap)
		if err != nil {
			return nil, fmt.Errorf("provision role: %w", err)
		}
		resp.ProvisionRole = &InstallPermissionsRoleStatus{
			AppAWSIAMRoleConfig: provisionRole,
		}

		deprovisionRole, err := renderRole(appCfg.PermissionsConfig.DeprovisionRole, stateMap)
		if err != nil {
			return nil, fmt.Errorf("deprovision role: %w", err)
		}
		resp.DeprovisionRole = &InstallPermissionsRoleStatus{
			AppAWSIAMRoleConfig: deprovisionRole,
		}

		maintenanceRole, err := renderRole(appCfg.PermissionsConfig.MaintenanceRole, stateMap)
		if err != nil {
			return nil, fmt.Errorf("maintenance role: %w", err)
		}
		resp.MaintenanceRole = &InstallPermissionsRoleStatus{
			AppAWSIAMRoleConfig: maintenanceRole,
		}

		for _, role := range appCfg.BreakGlassConfig.Roles {
			rendered, err := renderRole(role, stateMap)
			if err != nil {
				return nil, fmt.Errorf("break glass role: %w", err)
			}
			resp.BreakGlassRoles = append(resp.BreakGlassRoles, InstallPermissionsRoleStatus{
				AppAWSIAMRoleConfig: rendered,
			})
		}

		for _, role := range appCfg.PermissionsConfig.CustomRoles {
			rendered, err := renderRole(role, stateMap)
			if err != nil {
				return nil, fmt.Errorf("custom role: %w", err)
			}
			resp.CustomRoles = append(resp.CustomRoles, InstallPermissionsRoleStatus{
				AppAWSIAMRoleConfig: rendered,
			})
		}

		return resp, nil
	}

	// Provision role
	{
		role, err := renderRole(appCfg.PermissionsConfig.ProvisionRole, stateMap)
		if err != nil {
			return nil, fmt.Errorf("provision role: %w", err)
		}
		arn, _ := stackOutput.ProvisionRoleID()
		resp.ProvisionRole = &InstallPermissionsRoleStatus{
			AppAWSIAMRoleConfig: role,
			Enabled:             arn != "",
			ARN:                 arn,
		}
	}

	// Deprovision role
	{
		role, err := renderRole(appCfg.PermissionsConfig.DeprovisionRole, stateMap)
		if err != nil {
			return nil, fmt.Errorf("deprovision role: %w", err)
		}
		arn, _ := stackOutput.DeprovisionRoleID()
		resp.DeprovisionRole = &InstallPermissionsRoleStatus{
			AppAWSIAMRoleConfig: role,
			Enabled:             arn != "",
			ARN:                 arn,
		}
	}

	// Maintenance role
	{
		role, err := renderRole(appCfg.PermissionsConfig.MaintenanceRole, stateMap)
		if err != nil {
			return nil, fmt.Errorf("maintenance role: %w", err)
		}
		arn, _ := stackOutput.MaintenanceRoleID()
		resp.MaintenanceRole = &InstallPermissionsRoleStatus{
			AppAWSIAMRoleConfig: role,
			Enabled:             arn != "",
			ARN:                 arn,
		}
	}

	// Break glass roles (array, keyed by rendered name in stack outputs)
	for _, role := range appCfg.BreakGlassConfig.Roles {
		rendered, err := renderRole(role, stateMap)
		if err != nil {
			return nil, fmt.Errorf("break glass role: %w", err)
		}
		arn, _ := stackOutput.BreakGlassRoleID(rendered.Name)
		resp.BreakGlassRoles = append(resp.BreakGlassRoles, InstallPermissionsRoleStatus{
			AppAWSIAMRoleConfig: rendered,
			Enabled:             arn != "",
			ARN:                 arn,
		})
	}

	// Custom roles (array, keyed by rendered name in stack outputs)
	for _, role := range appCfg.PermissionsConfig.CustomRoles {
		rendered, err := renderRole(role, stateMap)
		if err != nil {
			return nil, fmt.Errorf("custom role: %w", err)
		}
		arn, _ := stackOutput.CustomRoleID(rendered.Name)
		resp.CustomRoles = append(resp.CustomRoles, InstallPermissionsRoleStatus{
			AppAWSIAMRoleConfig: rendered,
			Enabled:             arn != "",
			ARN:                 arn,
		})
	}

	return resp, nil
}
