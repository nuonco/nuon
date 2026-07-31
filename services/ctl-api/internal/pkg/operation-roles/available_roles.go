package operationroles

import (
	"fmt"

	"github.com/nuonco/nuon/pkg/types/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// AvailableRoles returns every role name an install can assume, mapped to its
// cloud identifier, with templated names rendered against install state. Roles
// the stack did not emit are dropped, since a name with no identifier cannot be
// assumed.
func AvailableRoles(
	appCfg *app.AppConfig,
	installStackOutputs *app.InstallStackOutputs,
	installState *state.State,
) (map[string]string, error) {
	if installStackOutputs == nil {
		return nil, fmt.Errorf("stack outputs are required")
	}

	var stackOutput app.StackOutput
	switch {
	case installStackOutputs.AzureStackOutputs != nil:
		stackOutput = installStackOutputs.AzureStackOutputs
	case installStackOutputs.AWSStackOutputs != nil:
		stackOutput = installStackOutputs.AWSStackOutputs
	case installStackOutputs.GCPStackOutputs != nil:
		stackOutput = installStackOutputs.GCPStackOutputs
	default:
		return nil, fmt.Errorf("stack outputs must have either AWS, Azure, or GCP outputs")
	}

	roles, err := getRoleMap(appCfg, stackOutput, installState)
	if err != nil {
		return nil, err
	}

	for name, id := range roles {
		if name == "" || id == "" {
			delete(roles, name)
		}
	}

	return roles, nil
}

// MaintenanceRoleName returns the install's rendered maintenance role name, the
// default identity for day-2 operations.
func MaintenanceRoleName(appCfg *app.AppConfig, installState *state.State) (string, error) {
	if appCfg == nil {
		return "", fmt.Errorf("app config is required")
	}
	return renderRoleName(appCfg.PermissionsConfig.MaintenanceRole.Name, installState)
}
