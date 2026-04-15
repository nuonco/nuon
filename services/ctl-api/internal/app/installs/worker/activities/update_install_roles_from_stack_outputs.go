package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/pkg/render"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
)

type UpdateInstallRolesFromStackOutputs struct {
	InstallID string `validate:"required"`
}

// @temporal-gen-v2 activity
func (a *Activities) UpdateInstallRolesFromStackOutputs(ctx context.Context, req UpdateInstallRolesFromStackOutputs) error {
	var install app.Install
	res := a.db.WithContext(ctx).
		Preload("InstallRoles").
		Preload("InstallRoles.AppRoleConfig").
		Preload("InstallStack").
		Preload("InstallStack.InstallStackOutputs").
		Where("id = ?", req.InstallID).
		First(&install)
	if res.Error != nil {
		return generics.TemporalGormError(res.Error)
	}

	if install.InstallStack == nil || install.InstallStack.InstallStackOutputs.ID == "" {
		return nil
	}

	outputs := install.InstallStack.InstallStackOutputs
	var stackOutputs app.StackOutput
	switch {
	case outputs.AWSStackOutputs != nil:
		stackOutputs = outputs.AWSStackOutputs
	case outputs.GCPStackOutputs != nil:
		stackOutputs = outputs.GCPStackOutputs
	case outputs.AzureStackOutputs != nil:
		stackOutputs = outputs.AzureStackOutputs
	default:
		return nil
	}

	installState, err := a.helpers.GetInstallState(ctx, req.InstallID, false, false)
	if err != nil {
		return fmt.Errorf("unable to get install state: %w", err)
	}
	stateData, err := installState.AsMap()
	if err != nil {
		return fmt.Errorf("unable to convert install state to map: %w", err)
	}

	for _, ir := range install.InstallRoles {
		roleCfg := ir.AppRoleConfig
		if err := render.RenderStruct(&roleCfg, stateData); err != nil {
			continue
		}

		roleID, err := resolveRoleID(stackOutputs, roleCfg)
		if err != nil {
			roleID = ""
		}

		provisioned := roleID != ""

		res := a.db.WithContext(ctx).
			Model(&app.InstallRoles{}).
			Where("id = ?", ir.ID).
			Updates(map[string]interface{}{
				"provisioned": provisioned,
				"role_id":     roleID,
			})
		if res.Error != nil {
			return generics.TemporalGormError(res.Error)
		}
	}

	return nil
}

func resolveRoleID(outputs app.StackOutput, roleCfg app.AppAWSIAMRoleConfig) (string, error) {
	switch roleCfg.Type {
	case app.AWSIAMRoleTypeRunnerProvision:
		return outputs.ProvisionRoleID()
	case app.AWSIAMRoleTypeRunnerDeprovision:
		return outputs.DeprovisionRoleID()
	case app.AWSIAMRoleTypeRunnerMaintenance:
		return outputs.MaintenanceRoleID()
	case app.AWSIAMRoleTypeCustom:
		return outputs.CustomRoleID(roleCfg.Name)
	case app.AWSIAMRoleTypeBreakGlass:
		return outputs.BreakGlassRoleID(roleCfg.Name)
	default:
		return "", fmt.Errorf("unknown role type: %s", roleCfg.Type)
	}
}
