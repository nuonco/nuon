package arm

import (
	"fmt"
	"regexp"
	"slices"
	"sort"

	"github.com/iancoleman/strcase"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks/cloudformation"
)

var armLogicalIDRegexp = regexp.MustCompile(`[^A-Za-z0-9]`)

func sanitizeDeploymentName(name string) string {
	camel := strcase.ToCamel(name)
	return armLogicalIDRegexp.ReplaceAllString(camel, "")
}

// customDeploymentIdentity holds info about a managed identity declared in a
// custom nested stack so the parent template can create the subscription-level
// role assignment that the identity needs.
type customDeploymentIdentity struct {
	DeploymentName string
}

func (t *Templates) getCustomLinkedDeployments(inp *stacks.TemplateInput) ([]any, map[string]ARMParameter, []customDeploymentIdentity, error) {
	stacks := inp.AppCfg.StackConfig.CustomNestedStacks
	if len(stacks) == 0 {
		return nil, nil, nil, nil
	}

	// Sort by index
	sorted := make([]config.CustomNestedStack, len(stacks))
	copy(sorted, stacks)
	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i].Index < sorted[j].Index
	})

	// Check for duplicate indices
	seenIndices := map[int]string{}
	for _, stack := range sorted {
		if prev, exists := seenIndices[stack.Index]; exists {
			return nil, nil, nil, fmt.Errorf("custom_nested_stacks: duplicate index %d for stacks %q and %q", stack.Index, prev, stack.Name)
		}
		seenIndices[stack.Index] = stack.Name
	}

	var resources []any
	var identities []customDeploymentIdentity
	hoistedParams := map[string]ARMParameter{}
	allParamNames := map[string]string{}
	prevDeploymentName := ""

	for i, stack := range sorted {
		if stack.Name == "" {
			return nil, nil, nil, fmt.Errorf("custom_nested_stacks[%d]: name is required", i)
		}
		if stack.TemplateURL == "" {
			return nil, nil, nil, fmt.Errorf("custom_nested_stacks[%d] (%s): template_url is required", i, stack.Name)
		}

		deploymentName := sanitizeDeploymentName(stack.Name)
		if deploymentName == "" {
			return nil, nil, nil, fmt.Errorf("custom_nested_stacks[%d] (%s): name produces invalid deployment name", i, stack.Name)
		}

		// Resolve template URL (use uploaded S3 URL if contents were uploaded)
		templateURL := stack.TemplateURL
		if stack.ContentsHash != "" && t.cfg.AWSCloudFormationStackTemplateBaseURL != "" {
			templateURL = cloudformation.CustomNestedStackTemplateURL(
				t.cfg.AWSCloudFormationStackTemplateBaseURL,
				inp.Install.OrgID,
				inp.AppCfg.AppID,
				stack.ContentsHash,
				stack.TemplateURL,
			)
		}

		// Fetch template, validate structure, and extract parameters
		armTmpl, err := fetchARMTemplate(templateURL)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("custom_nested_stacks[%d] (%s): %w", i, stack.Name, err)
		}

		if err := validateARMTemplate(armTmpl); err != nil {
			return nil, nil, nil, fmt.Errorf("custom_nested_stacks[%d] (%s): %w", i, stack.Name, err)
		}

		params, defaultParams := extractARMParameters(armTmpl, ReservedParamNames)

		// Track custom nested stacks that declare managed identities so the
		// parent template can create subscription-level role assignments.
		if armTmpl.hasManagedIdentity() {
			identities = append(identities, customDeploymentIdentity{
				DeploymentName: deploymentName,
			})
		}

		// Check for parameter name conflicts
		for paramName := range defaultParams {
			if owner, exists := allParamNames[paramName]; exists {
				return nil, nil, nil, fmt.Errorf("custom_nested_stacks[%d] (%s): parameter %q conflicts with stack %q", i, stack.Name, paramName, owner)
			}
			allParamNames[paramName] = stack.Name
		}

		// Build deployment parameters
		deploymentParams := map[string]any{}

		// Inject Nuon-reserved params if template declares them
		nuonParams := map[string]string{
			"nuonInstallID": inp.Install.ID,
			"nuonOrgID":     inp.Runner.OrgID,
			"nuonAppID":     inp.Install.AppID,
			"location":      "[parameters('location')]",
		}
		for paramName := range params {
			if val, ok := nuonParams[paramName]; ok {
				deploymentParams[paramName] = map[string]any{"value": val}
			}
		}

		// Resolve explicit parameter mappings from install inputs
		for cfnParamName, templateValue := range stack.Parameters {
			inputName, err := config.ParseInstallInputReference(templateValue)
			if err != nil {
				return nil, nil, nil, fmt.Errorf("custom_nested_stacks[%d] (%s): parameter %q: %w", i, stack.Name, cfnParamName, err)
			}

			resolved := ""
			if inp.Install.CurrentInstallInputs != nil {
				if val, ok := inp.Install.CurrentInstallInputs.Values[inputName]; ok && val != nil {
					resolved = *val
				}
			}

			deploymentParams[cfnParamName] = map[string]any{"value": resolved}
			delete(defaultParams, cfnParamName)
		}

		// Wire VNet outputs to matching parameter names
		vnetOutputs := []string{
			"vnetId", "vnetName",
			"publicSubnet1Id", "publicSubnet1Name",
			"publicSubnet2Id", "publicSubnet2Name",
			"publicSubnet3Id", "publicSubnet3Name",
			"privateSubnet1Id", "privateSubnet1Name",
			"privateSubnet2Id", "privateSubnet2Name",
			"privateSubnet3Id", "privateSubnet3Name",
			"runnerSubnetId", "runnerSubnetName",
			"publicSubnetIds", "publicSubnetNames",
			"privateSubnetIds", "privateSubnetNames",
		}
		for paramName := range params {
			if slices.Contains(vnetOutputs, paramName) {
				deploymentParams[paramName] = map[string]any{
					"value": fmt.Sprintf("[reference('vnetDeployment').outputs.%s.value]", paramName),
				}
				delete(defaultParams, paramName)
			}
		}

		// Remaining params are hoisted into parent
		for paramName := range defaultParams {
			if _, alreadySet := deploymentParams[paramName]; !alreadySet {
				deploymentParams[paramName] = map[string]any{"value": fmt.Sprintf("[parameters('%s')]", paramName)}
			}
		}

		// Merge hoisted params
		for k, v := range defaultParams {
			hoistedParams[k] = v
		}

		// Build dependsOn chain
		var dependsOn []string
		if prevDeploymentName != "" {
			dependsOn = append(dependsOn, prevDeploymentName)
		} else {
			dependsOn = append(dependsOn, "vnetDeployment")
			if !t.cfg.UseLocalRunners {
				dependsOn = append(dependsOn, "runnerDeployment")
			}
		}

		deployment := map[string]any{
			"type":       "Microsoft.Resources/deployments",
			"apiVersion": "2022-09-01",
			"name":       deploymentName,
			"dependsOn":  dependsOn,
			"properties": map[string]any{
				"mode": "Incremental",
				"templateLink": map[string]any{
					"uri": templateURL,
				},
				"parameters": deploymentParams,
			},
		}

		resources = append(resources, deployment)
		prevDeploymentName = deploymentName
	}

	return resources, hoistedParams, identities, nil
}
