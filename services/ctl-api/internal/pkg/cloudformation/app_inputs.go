package cloudformation

import (
	"github.com/awslabs/goformation/v7/cloudformation"
	"github.com/iancoleman/strcase"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

// getInstallInputParameters returns CloudFormation parameters for inputs with source "install_stack"
func (t *Templates) getInstallInputParameters(appInputs []app.AppInput, inputGroup string) map[string]cloudformation.Parameter {
	params := make(map[string]cloudformation.Parameter)

	// Get app input config to find install_stack sourced inputs
	for _, input := range appInputs {
		if input.Source != app.AppInputSourceCustomer {
			continue
		}

		params[input.CloudFormationStackParamName] = cloudformation.Parameter{
			Type:        getCloudFormationTypeForInput(input.Type),
			Default:     input.Default,
			Description: &input.Description,
		}
	}

	return params
}

func (a *Templates) getInstallInputsParamLabels(appInputs []app.AppInput, inputGroup string) map[string]any {
	paramLabels := make(map[string]any, 0)
	for _, input := range appInputs {
		if input.Source != app.AppInputSourceCustomer {
			continue
		}

		displayName := input.DisplayName
		if displayName == "" {
			displayName = input.Name
		}

		paramLabels[input.CloudFormationStackParamName] = displayName
	}

	return paramLabels
}

func (t *Templates) getInstallInputGroupParameters(inp *TemplateInput) map[string]map[string]cloudformation.Parameter {
	groupIDAppInputs := make(map[string][]app.AppInput)
	for _, inputGroup := range inp.AppCfg.InputConfig.AppInputGroups {
		groupIDAppInputs[inputGroup.ID] = make([]app.AppInput, 0)
	}
	for _, appInput := range inp.AppCfg.InputConfig.AppInputs {
		groupIDAppInputs[appInput.AppInputGroupID] = append(groupIDAppInputs[appInput.AppInputGroupID], appInput)
	}

	installGroupParameters := make(map[string]map[string]cloudformation.Parameter)
	for _, inputGroup := range inp.AppCfg.InputConfig.AppInputGroups {
		installInputGroupParams := t.getInstallInputParameters(
			groupIDAppInputs[inputGroup.ID],
			strcase.ToCamel(inputGroup.Name),
		)
		if len(installInputGroupParams) == 0 {
			continue
		}
		installGroupParameters[inputGroup.Name] = installInputGroupParams
	}
	return installGroupParameters
}

func (t *Templates) getInstallInputGroupParamLable(inp *TemplateInput) map[string]map[string]any {
	groupIDAppInputs := make(map[string][]app.AppInput)
	for _, inputGroup := range inp.AppCfg.InputConfig.AppInputGroups {
		groupIDAppInputs[inputGroup.ID] = make([]app.AppInput, 0)
	}
	for _, appInput := range inp.AppCfg.InputConfig.AppInputs {
		groupIDAppInputs[appInput.AppInputGroupID] = append(groupIDAppInputs[appInput.AppInputGroupID], appInput)
	}

	installGroupInputParamLables := make(map[string]map[string]any)
	for _, inputGroup := range inp.AppCfg.InputConfig.AppInputGroups {
		installInputParamLabels := t.getInstallInputsParamLabels(
			groupIDAppInputs[inputGroup.ID],
			strcase.ToCamel(inputGroup.Name),
		)
		if len(installInputParamLabels) == 0 {
			continue
		}
		installInputParamLabels[inputGroup.Name] = installInputParamLabels
	}
	return installGroupInputParamLables
}

// getCloudFormationTypeForInput converts app input type to CloudFormation parameter type
func getCloudFormationTypeForInput(inputType app.AppInputType) string {
	switch inputType {
	case app.AppInputTypeNumber:
		return "Number"
	case app.AppInputTypeBool:
		return "String" // CloudFormation doesn't have boolean parameters
	case app.AppInputTypeList:
		return "CommaDelimitedList"
	case app.AppInputTypeJSON, app.AppInputTypeString:
		fallthrough
	default:
		return "String"
	}
}
