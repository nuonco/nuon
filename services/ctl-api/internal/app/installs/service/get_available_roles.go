package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nuonco/nuon/pkg/principal"
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

	installStack, err := s.getInstallStack(ctx, installID, org.ID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install stack: %w", err))
		return
	}

	if installStack == nil || installStack.InstallStackOutputs.AWSStackOutputs == nil {
		ctx.JSON(http.StatusOK, AvailableRolesResponse{Roles: []AvailableRole{}})
		return
	}

	roles := buildAvailableRoles(installStack.InstallStackOutputs.AWSStackOutputs, operationType)

	ctx.JSON(http.StatusOK, AvailableRolesResponse{Roles: roles})
}

func buildAvailableRoles(aws *app.AWSStackOutputs, operationType string) []AvailableRole {
	roles := []AvailableRole{}

	if aws == nil {
		return roles
	}

	for name, arn := range aws.CustomRoleARNs {
		roles = append(roles, AvailableRole{
			Name:     name,
			ARN:      arn,
			RoleType: "custom",
		})
	}

	for name, arn := range aws.BreakGlassRoleARNs {
		roles = append(roles, AvailableRole{
			Name:     name,
			ARN:      arn,
			RoleType: "break_glass",
		})
	}

	switch operationType {
	case "provision", "reprovision":
		if aws.ProvisionIAMRoleARN != "" {
			roles = append(roles, AvailableRole{
				Name:     "provision",
				ARN:      aws.ProvisionIAMRoleARN,
				RoleType: "provision",
			})
		}
	case "deprovision":
		if aws.DeprovisionIAMRoleARN != "" {
			roles = append(roles, AvailableRole{
				Name:     "deprovision",
				ARN:      aws.DeprovisionIAMRoleARN,
				RoleType: "deprovision",
			})
		}
	case "deploy", "teardown":
		if aws.MaintenanceIAMRoleARN != "" {
			roles = append(roles, AvailableRole{
				Name:     "maintenance",
				ARN:      aws.MaintenanceIAMRoleARN,
				RoleType: "maintenance",
			})
		}
	case "trigger":
		if aws.MaintenanceIAMRoleARN != "" {
			roles = append(roles, AvailableRole{
				Name:     "maintenance",
				ARN:      aws.MaintenanceIAMRoleARN,
				RoleType: "maintenance",
			})
		}
	}

	return roles
}

func validatePrincipalType(principalType string) error {
	if principalType == "" {
		return fmt.Errorf("principal_type query parameter is required")
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
		return fmt.Errorf("operation_type query parameter is required")
	}

	for _, validOp := range app.ValidOperations {
		if operationType == string(validOp) {
			return nil
		}
	}

	return fmt.Errorf("operation_type must be one of: provision, reprovision, deprovision, deploy, teardown, trigger (got: %s)", operationType)
}
