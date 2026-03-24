package service

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nuonco/nuon/pkg/principal"
	"github.com/nuonco/nuon/pkg/render"
	"github.com/nuonco/nuon/pkg/types/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type AvailableRole struct {
	Name     string `json:"name"`
	ARN      string `json:"arn"`
	RoleType string `json:"role_type"`
}

type AvailableRolesResponse struct {
	Roles []AvailableRole `json:"roles"`
}

// @ID						GetAvailableRoles
// @Summary				get available IAM roles for a specific operation
// @Description.markdown	get_available_roles.md
// @Param					install_id					path	string	true	"install ID"
// @Param					principal_type				query	principal.Type	true	"principal type: component, sandbox, action"
// @Param					app.operationType				query	string	true	"operation type: provision, reprovision, deprovision, deploy, teardown, trigger"
// @Tags					installs
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	AvailableRolesResponse
// @Router					/v1/installs/{install_id}/available-roles [GET]
func (s *service) GetAvailableRoles(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	installID := ctx.Param("install_id")
	principalType := ctx.Query("principal_type")
	operationType := ctx.Query("operation_type")

	if err := validatePrincipalType(principalType); err != nil {
		ctx.Error(stderr.ErrUser{
			Err:         err,
			Description: err.Error(),
		})
		return
	}

	if err := validateOperationType(operationType); err != nil {
		ctx.Error(stderr.ErrUser{
			Err:         err,
			Description: err.Error(),
		})
		return
	}

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

	state, err := s.helpers.GetInstallState(ctx, installID, false, false)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install state: %w", err))
		return
	}

	installStack, err := s.getInstallStack(ctx, installID, org.ID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install stack: %w", err))
		return
	}

	if installStack == nil {
		ctx.JSON(http.StatusOK, AvailableRolesResponse{Roles: []AvailableRole{}})
		return
	}

	outputs := installStack.InstallStackOutputs
	var roles []AvailableRole

	var stackOutput app.StackOutput
	switch {
	case outputs.AWSStackOutputs != nil:
		stackOutput = outputs.AWSStackOutputs
	case outputs.GCPStackOutputs != nil:
		stackOutput = outputs.GCPStackOutputs
	default:
		ctx.JSON(http.StatusOK, AvailableRolesResponse{Roles: []AvailableRole{}})
		return
	}

	roles, err = buildAvailableRoles(stackOutput, appCfg, state, operationType)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to build available roles: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, AvailableRolesResponse{Roles: roles})
}

func buildAvailableRoles(stackOutputs app.StackOutput, appCfg *app.AppConfig, state *state.State, operationType string) ([]AvailableRole, error) {
	roles := []AvailableRole{}

	if stackOutputs == nil {
		return roles, nil
	}

	stateMap, err := state.AsMap()
	if err != nil {
		return nil, fmt.Errorf("unable to get install state map: %w", err)
	}

	customRoles, err := stackOutputs.CustomRoles()
	if err != nil {
		return nil, fmt.Errorf("unable to fetch custom roles %w", err)
	}
	for name, arn := range customRoles {
		rendered, err := render.RenderV2(name, stateMap)
		if err != nil {
			return nil, fmt.Errorf("unable to render provision role name: %w", err)
		}
		roles = append(roles, AvailableRole{
			Name:     rendered,
			ARN:      arn,
			RoleType: "custom",
		})
	}

	breakGlassRoles, err := stackOutputs.BreakGlassRoles()
	if err != nil {
		return nil, fmt.Errorf("unable to fetch break glass roles %w", err)
	}
	for name, arn := range breakGlassRoles {
		rendered, err := render.RenderV2(name, stateMap)
		if err != nil {
			return nil, fmt.Errorf("unable to render provision role name: %w", err)
		}
		roles = append(roles, AvailableRole{
			Name:     rendered,
			ARN:      arn,
			RoleType: "break_glass",
		})
	}

	provisionRoleID, err := stackOutputs.ProvisionRoleID()
	if err != nil {
		return nil, fmt.Errorf("unable to fetch provision role %w", err)
	}
	if provisionRoleID != "" {
		rendered, err := render.RenderV2(appCfg.PermissionsConfig.ProvisionRole.Name, stateMap)
		if err != nil {
			return nil, fmt.Errorf("unable to render provision role name: %w", err)
		}
		roles = append(roles, AvailableRole{
			Name:     rendered,
			ARN:      provisionRoleID,
			RoleType: "provision",
		})
	}

	deprovisionRoleID, err := stackOutputs.DeprovisionRoleID()
	if err != nil {
		return nil, fmt.Errorf("unable to fetch deprovision role %w", err)
	}
	if deprovisionRoleID != "" {
		rendered, err := render.RenderV2(appCfg.PermissionsConfig.DeprovisionRole.Name, stateMap)
		if err != nil {
			return nil, fmt.Errorf("unable to render deprovision role name: %w", err)
		}
		roles = append(roles, AvailableRole{
			Name:     rendered,
			ARN:      deprovisionRoleID,
			RoleType: "deprovision",
		})
	}

	maintenanceRoleID, err := stackOutputs.MaintenanceRoleID()
	if err != nil {
		return nil, fmt.Errorf("unable to fetch maintenance role %w", err)
	}
	if maintenanceRoleID != "" {
		rendered, err := render.RenderV2(appCfg.PermissionsConfig.MaintenanceRole.Name, stateMap)
		if err != nil {
			return nil, fmt.Errorf("unable to render maintenance role name: %s", err)
		}
		roles = append(roles, AvailableRole{
			Name:     rendered,
			ARN:      maintenanceRoleID,
			RoleType: "maintenance",
		})
	}

	return roles, nil
}

func validatePrincipalType(principalType string) error {
	if principalType == "" {
		return errors.New("principal_type query parameter is required")
	}

	for _, validType := range principal.ValidTypes {
		if principalType == string(validType) {
			return nil
		}
	}

	return fmt.Errorf("principal_type must be one of: component, sandbox, action (got: %s)", principalType)
}

func validateOperationType(operationType string) error {
	if operationType == "" {
		return errors.New("operation_type query parameter is required")
	}

	for _, validOp := range app.ValidOperations {
		if operationType == string(validOp) {
			return nil
		}
	}

	return fmt.Errorf("operation_type must be one of: provision, reprovision, deprovision, deploy, teardown, trigger (got: %s)", operationType)
}
