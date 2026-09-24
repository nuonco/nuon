package arm

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks"
)

const customStacksResourceGroupParameter = "nuonResourceGroupName"

func (t *Templates) getAzureCustomStacksOnlyTemplate(inp *stacks.TemplateInput) (*ARMTemplate, map[string]map[string]string, map[string]map[string]string, error) {
	scope := armScope{subscription: true}
	tmpl := &ARMTemplate{
		Schema:         subscriptionTemplateSchema,
		ContentVersion: "1.0.0.0",
		Parameters:     make(map[string]ARMParameter),
		Variables: map[string]any{
			installRGVarName: "[parameters('" + customStacksResourceGroupParameter + "')]",
			locationVarName:  "[parameters('location')]",
		},
		Resources: []any{},
		Outputs:   make(map[string]ARMOutput),
	}

	contractParameters := append([]string{}, vnetContractOutputs...)
	contractParameters = append(contractParameters, customStacksResourceGroupParameter, "location")
	for _, name := range contractParameters {
		tmpl.Parameters[name] = ARMParameter{Type: "string"}
	}

	resources, params, identities, outputs, err := t.getCustomLinkedDeployments(inp)
	if err != nil {
		return nil, nil, nil, err
	}
	if len(resources) == 0 {
		return nil, nil, nil, fmt.Errorf("custom stacks only template requested but app config has no custom_nested_stacks configured")
	}

	tmpl.Resources = append(tmpl.Resources, resources...)
	rootParameterNames := make(map[string]string, len(tmpl.Parameters))
	for name := range tmpl.Parameters {
		rootParameterNames[strings.ToLower(name)] = name
	}
	for name, param := range params {
		if existing, exists := rootParameterNames[strings.ToLower(name)]; exists {
			return nil, nil, nil, fmt.Errorf("custom stacks only template: parameter %q conflicts with the Nuon module contract parameter %q", name, existing)
		}
		tmpl.Parameters[name] = param
		rootParameterNames[strings.ToLower(name)] = name
	}
	inputParametersMap, err := liftCustomStackInstallInputs(tmpl, inp, resources, outputs)
	if err != nil {
		return nil, nil, nil, err
	}
	for _, identity := range identities {
		tmpl.Resources = append(tmpl.Resources, t.getCustomDeploymentRoleAssignment(identity, inp.Install.ID, scope))
	}

	outputMap := make(map[string]map[string]string, len(outputs))
	seen := make(map[string]string)
	for _, stackOutput := range outputs {
		stackMap := make(map[string]string, len(stackOutput.OutputKeys))
		for _, outputKey := range stackOutput.OutputKeys {
			flatName := sanitizeDeploymentName(stackOutput.StackName) + sanitizeDeploymentName(outputKey)
			if flatName == "" {
				return nil, nil, nil, fmt.Errorf("custom stacks only template: stack %q output %q produces an empty output name", stackOutput.StackName, outputKey)
			}
			canonicalName := strings.ToLower(flatName)
			if owner, exists := seen[canonicalName]; exists {
				return nil, nil, nil, fmt.Errorf("custom stacks only template: output name %q collides between %s and %s.%s", flatName, owner, stackOutput.StackName, outputKey)
			}
			seen[canonicalName] = stackOutput.StackName + "." + outputKey
			stackMap[outputKey] = flatName
			tmpl.Outputs[flatName] = ARMOutput{
				Type:  "string",
				Value: fmt.Sprintf("[string(reference('%s').outputs.%s.value)]", stackOutput.DeploymentName, outputKey),
			}
		}
		outputMap[stackOutput.StackName] = stackMap
	}

	return tmpl, outputMap, inputParametersMap, nil
}

func liftCustomStackInstallInputs(
	tmpl *ARMTemplate,
	inp *stacks.TemplateInput,
	resources []any,
	outputs []customDeploymentOutputs,
) (map[string]map[string]string, error) {
	deploymentNames := make(map[string]string, len(outputs))
	for _, output := range outputs {
		deploymentNames[output.StackName] = output.DeploymentName
	}
	deployments := make(map[string]map[string]any, len(resources))
	for _, resource := range resources {
		deployment := resource.(map[string]any)
		deployments[deployment["name"].(string)] = deployment
	}

	result := make(map[string]map[string]string)
	owners := make(map[string]string)
	rootParameterNames := make(map[string]struct{}, len(tmpl.Parameters))
	for name := range tmpl.Parameters {
		rootParameterNames[strings.ToLower(name)] = struct{}{}
	}
	customerInputNames := make(map[string]struct{})
	for _, input := range inp.AppCfg.InputConfig.AppInputs {
		if input.Source == app.AppInputSourceCustomer {
			customerInputNames[input.Name] = struct{}{}
		}
	}
	for _, stack := range inp.AppCfg.StackConfig.CustomNestedStacks {
		deployment := deployments[deploymentNames[stack.Name]]
		if deployment == nil {
			continue
		}
		properties := deployment["properties"].(map[string]any)
		parameters := properties["parameters"].(map[string]any)

		for _, parameterName := range slices.Sorted(maps.Keys(inp.UnrenderedCustomStackParameters[stack.Name])) {
			inputName, err := config.ParseInstallInputReference(inp.UnrenderedCustomStackParameters[stack.Name][parameterName])
			if err != nil {
				continue
			}
			if _, ok := customerInputNames[inputName]; !ok {
				continue
			}
			if _, exists := parameters[parameterName]; !exists {
				continue
			}

			topLevelName := sanitizeDeploymentName(stack.Name) + sanitizeDeploymentName(parameterName)
			canonicalName := strings.ToLower(topLevelName)
			if owner, exists := owners[canonicalName]; exists {
				return nil, fmt.Errorf("custom stacks only template: input parameter name %q collides between %s and %s.%s", topLevelName, owner, stack.Name, parameterName)
			}
			if _, exists := rootParameterNames[canonicalName]; exists {
				return nil, fmt.Errorf("custom stacks only template: input parameter %s.%s conflicts with root parameter %q", stack.Name, parameterName, topLevelName)
			}

			owners[canonicalName] = stack.Name + "." + parameterName
			rootParameterNames[canonicalName] = struct{}{}
			tmpl.Parameters[topLevelName] = ARMParameter{Type: "string"}
			parameters[parameterName] = map[string]any{"value": "[parameters('" + topLevelName + "')]"}
			if result[stack.Name] == nil {
				result[stack.Name] = make(map[string]string)
			}
			result[stack.Name][topLevelName] = inputName
		}
	}

	return result, nil
}
