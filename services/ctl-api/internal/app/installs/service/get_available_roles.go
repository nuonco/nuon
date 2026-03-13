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

	switch {
	case outputs.AWSStackOutputs != nil:
		roles, err = buildAvailableRolesAWS(outputs.AWSStackOutputs, appCfg, state, operationType)
	case outputs.GCPStackOutputs != nil:
		roles, err = buildAvailableRolesGCP(outputs.GCPStackOutputs, appCfg, state, operationType)
	default:
		ctx.JSON(http.StatusOK, AvailableRolesResponse{Roles: []AvailableRole{}})
		return
	}

	if err != nil {
		ctx.Error(fmt.Errorf("unable to build available roles: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, AvailableRolesResponse{Roles: roles})
}

func buildAvailableRolesAWS(installStackOutput *app.AWSStackOutputs, appCfg *app.AppConfig, state *state.State, operationType string) ([]AvailableRole, error) {
	roles := []AvailableRole{}

	if installStackOutput == nil {
		return roles, nil
	}

	stateMap, err := state.AsMap()
	if err != nil {
		return nil, fmt.Errorf("unable to get install state map: %w", err)
	}

	for name, arn := range installStackOutput.CustomRoleARNs {
		roles = append(roles, AvailableRole{
			Name:     name,
			ARN:      arn,
			RoleType: "custom",
		})
	}

	for name, arn := range installStackOutput.BreakGlassRoleARNs {
		roles = append(roles, AvailableRole{
			Name:     name,
			ARN:      arn,
			RoleType: "break_glass",
		})
	}

	if installStackOutput.ProvisionIAMRoleARN != "" {
		rendered, err := render.RenderV2(appCfg.PermissionsConfig.ProvisionRole.Name, stateMap)
		if err != nil {
			return nil, fmt.Errorf("unable to render provision role name: %w", err)
		}
		roles = append(roles, AvailableRole{
			Name:     rendered,
			ARN:      installStackOutput.ProvisionIAMRoleARN,
			RoleType: "provision",
		})
	}

	if installStackOutput.DeprovisionIAMRoleARN != "" {
		rendered, err := render.RenderV2(appCfg.PermissionsConfig.DeprovisionRole.Name, stateMap)
		if err != nil {
			return nil, fmt.Errorf("unable to render deprovision role name: %w", err)
		}
		roles = append(roles, AvailableRole{
			Name:     rendered,
			ARN:      installStackOutput.DeprovisionIAMRoleARN,
			RoleType: "deprovision",
		})
	}

	if installStackOutput.MaintenanceIAMRoleARN != "" {
		rendered, err := render.RenderV2(appCfg.PermissionsConfig.MaintenanceRole.Name, stateMap)
		if err != nil {
			return nil, fmt.Errorf("unable to render maintenance role name: %s", err)
		}
		roles = append(roles, AvailableRole{
			Name:     rendered,
			ARN:      installStackOutput.MaintenanceIAMRoleARN,
			RoleType: "maintenance",
		})
	}

	return roles, nil
}

func buildAvailableRolesGCP(installStackOutput *app.GCPStackOutputs, appCfg *app.AppConfig, state *state.State, operationType string) ([]AvailableRole, error) {
	roles := []AvailableRole{}

	if installStackOutput == nil {
		return roles, nil
	}

	stateMap, err := state.AsMap()
	if err != nil {
		return nil, fmt.Errorf("unable to get install state map: %w", err)
	}

	for name, email := range installStackOutput.CustomSAEmails {
		roles = append(roles, AvailableRole{
			Name:     name,
			ARN:      email,
			RoleType: "custom",
		})
	}

	for name, email := range installStackOutput.BreakGlassSAEmails {
		roles = append(roles, AvailableRole{
			Name:     name,
			ARN:      email,
			RoleType: "break_glass",
		})
	}

	if installStackOutput.ProvisionSAEmail != "" {
		rendered, err := render.RenderV2(appCfg.PermissionsConfig.ProvisionRole.Name, stateMap)
		if err != nil {
			return nil, fmt.Errorf("unable to render provision role name: %w", err)
		}
		roles = append(roles, AvailableRole{
			Name:     rendered,
			ARN:      installStackOutput.ProvisionSAEmail,
			RoleType: "provision",
		})
	}

	if installStackOutput.DeprovisionSAEmail != "" {
		rendered, err := render.RenderV2(appCfg.PermissionsConfig.DeprovisionRole.Name, stateMap)
		if err != nil {
			return nil, fmt.Errorf("unable to render deprovision role name: %w", err)
		}
		roles = append(roles, AvailableRole{
			Name:     rendered,
			ARN:      installStackOutput.DeprovisionSAEmail,
			RoleType: "deprovision",
		})
	}

	if installStackOutput.MaintenanceSAEmail != "" {
		rendered, err := render.RenderV2(appCfg.PermissionsConfig.MaintenanceRole.Name, stateMap)
		if err != nil {
			return nil, fmt.Errorf("unable to render maintenance role name: %w", err)
		}
		roles = append(roles, AvailableRole{
			Name:     rendered,
			ARN:      installStackOutput.MaintenanceSAEmail,
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
