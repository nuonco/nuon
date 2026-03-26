package cloudformation

import (
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strings"

	"github.com/awslabs/goformation/v7/cloudformation"
	nestedcloudformation "github.com/awslabs/goformation/v7/cloudformation/cloudformation"
	"github.com/iancoleman/strcase"

	"github.com/nuonco/nuon/pkg/config"
	pkggenerics "github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks"
)

var logicalIDRegexp = regexp.MustCompile(`[^A-Za-z0-9]`)

type customNestedStackResult struct {
	resources   map[string]*nestedcloudformation.Stack
	params      map[string]cloudformation.Parameter
	paramGroups []map[string]any
}

func sanitizeLogicalID(name string) string {
	camel := strcase.ToCamel(name)
	return logicalIDRegexp.ReplaceAllString(camel, "")
}

func (tpl *Templates) getCustomNestedStacks(inp *stacks.TemplateInput, t tagBuilder, existingResourceKeys map[string]bool) (*customNestedStackResult, error) {
	if len(inp.AppCfg.StackConfig.CustomNestedStacks) == 0 {
		return &customNestedStackResult{
			resources:   map[string]*nestedcloudformation.Stack{},
			params:      map[string]cloudformation.Parameter{},
			paramGroups: nil,
		}, nil
	}

	sorted := make([]config.CustomNestedStack, len(inp.AppCfg.StackConfig.CustomNestedStacks))
	copy(sorted, inp.AppCfg.StackConfig.CustomNestedStacks)
	slices.SortStableFunc(sorted, func(a, b config.CustomNestedStack) int {
		return a.Index - b.Index
	})

	seenIndices := map[int]string{}
	for _, stack := range sorted {
		if prev, exists := seenIndices[stack.Index]; exists {
			return nil, fmt.Errorf("custom_nested_stacks: duplicate index %d for stacks %q and %q", stack.Index, prev, stack.Name)
		}
		seenIndices[stack.Index] = stack.Name
	}

	firstClassStacks := map[string]string{
		"VPC": inp.AppCfg.StackConfig.VPCNestedTemplateURL,
	}
	if _, hasASG := existingResourceKeys["RunnerAutoScalingGroup"]; hasASG {
		firstClassStacks["RunnerAutoScalingGroup"] = inp.AppCfg.StackConfig.RunnerNestedTemplateURL
	}
	fcOutputs := tpl.extractFirstClassOutputs(firstClassStacks)

	result := &customNestedStackResult{
		resources: map[string]*nestedcloudformation.Stack{},
		params:    map[string]cloudformation.Parameter{},
	}

	allParamNames := map[string]string{}
	prevLogicalID := ""

	for i, stack := range sorted {
		if stack.Name == "" {
			return nil, fmt.Errorf("custom_nested_stacks[%d]: name is required", i)
		}
		if stack.TemplateURL == "" {
			return nil, fmt.Errorf("custom_nested_stacks[%d] (%s): template_url is required", i, stack.Name)
		}

		logicalID := sanitizeLogicalID(stack.Name)
		if logicalID == "" {
			return nil, fmt.Errorf("custom_nested_stacks[%d] (%s): name produces invalid CloudFormation logical ID", i, stack.Name)
		}
		if existingResourceKeys[logicalID] {
			return nil, fmt.Errorf("custom_nested_stacks[%d] (%s): logical ID %q conflicts with existing resource", i, stack.Name, logicalID)
		}
		if _, exists := result.resources[logicalID]; exists {
			return nil, fmt.Errorf("custom_nested_stacks[%d] (%s): duplicate logical ID %q", i, stack.Name, logicalID)
		}

		nestedStack, defaultParams, templateOutputs, err := tpl.buildCustomNestedStack(inp, stack, t, logicalID, prevLogicalID, fcOutputs)
		if err != nil {
			return nil, fmt.Errorf("custom_nested_stacks[%d] (%s): %w", i, stack.Name, err)
		}

		for paramName := range defaultParams {
			if owner, exists := allParamNames[paramName]; exists {
				return nil, fmt.Errorf("custom_nested_stacks[%d] (%s): parameter %q conflicts with stack %q", i, stack.Name, paramName, owner)
			}
			allParamNames[paramName] = stack.Name
		}

		result.resources[logicalID] = nestedStack
		maps.Copy(result.params, defaultParams)

		if len(defaultParams) > 0 {
			result.paramGroups = append(result.paramGroups, map[string]any{
				"Label": map[string]any{
					"default": stack.Name,
				},
				"Parameters": pkggenerics.MapToKeys(defaultParams),
			})
		}

		for outputName := range templateOutputs {
			if _, isFirstClass := fcOutputs[outputName]; !isFirstClass {
				fcOutputs[outputName] = firstClassOutput{
					resource:   logicalID,
					outputName: outputName,
				}
			}
		}

		prevLogicalID = logicalID
	}

	return result, nil
}

func (tpl *Templates) buildCustomNestedStack(inp *stacks.TemplateInput, stack config.CustomNestedStack, t tagBuilder, logicalID string, prevLogicalID string, fcOutputs map[string]firstClassOutput) (*nestedcloudformation.Stack, map[string]cloudformation.Parameter, map[string]struct{}, error) {
	// Build role param lookup so we can treat them as reserved during extraction.
	type roleRef struct {
		paramValue string
		resource   string
	}
	roleParams := make(map[string]roleRef)
	for _, role := range inp.AppCfg.PermissionsConfig.Roles {
		roleParams[role.CloudFormationStackParamName] = roleRef{
			paramValue: cloudformation.Ref(role.CloudFormationStackParamName),
			resource:   role.CloudFormationStackName,
		}
	}
	for _, role := range inp.AppCfg.BreakGlassConfig.Roles {
		roleParams[role.CloudFormationStackParamName] = roleRef{
			paramValue: cloudformation.Ref(role.CloudFormationStackParamName),
			resource:   role.CloudFormationStackName,
		}
	}

	roleParamNames := make([]string, 0, len(roleParams))
	for k := range roleParams {
		roleParamNames = append(roleParamNames, k)
	}

	// Use S3 URL when template has been uploaded (ContentsHash is set).
	templateURL := stack.TemplateURL
	if stack.ContentsHash != "" {
		templateURL = CustomNestedStackTemplateURL(
			tpl.cfg.AWSCloudFormationStackTemplateBaseURL,
			inp.Install.OrgID, inp.AppCfg.AppID,
			stack.ContentsHash, stack.TemplateURL,
		)
	}

	parameters, defaultParameters, reservedInTemplate, templateOutputs, err := tpl.extractNestedStackParameters(templateURL, roleParamNames...)
	if err != nil {
		return nil, nil, nil, err
	}

	// Inject reserved Nuon params only if the template declares them
	nuonParams := map[string]string{
		"NuonInstallID": inp.Install.ID,
		"NuonAppID":     inp.Install.AppID,
		"NuonOrgID":     inp.Install.OrgID,
	}
	for k, v := range nuonParams {
		if reservedInTemplate[k] {
			parameters[k] = v
		}
	}

	// Inject role enable params only if the template declares them.
	// NOTE: we intentionally do NOT add roleDeps to DependsOn because
	// the role resources are conditional (they have Condition:
	// Enable<RoleType>). A hard DependsOn on a conditional resource
	// that doesn't exist causes "Unresolved resource dependencies".
	// The nested stack only receives the Enable* boolean parameter
	// (via Ref), which doesn't require a resource dependency.
	for paramName, ref := range roleParams {
		if reservedInTemplate[paramName] {
			parameters[paramName] = ref.paramValue
		}
	}

	// Resolve explicit parameter mappings from install inputs
	for cfnParamName, templateValue := range stack.Parameters {
		inputName, err := config.ParseInstallInputReference(templateValue)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("parameter %q: %w", cfnParamName, err)
		}

		resolved := ""
		if inp.Install.CurrentInstallInputs != nil {
			if val, ok := inp.Install.CurrentInstallInputs.Values[inputName]; ok && val != nil {
				resolved = *val
			}
		}

		parameters[cfnParamName] = resolved
		delete(defaultParameters, cfnParamName)
	}

	// Wire first-class stack outputs (VPC, Runner) to matching parameter names
	for paramName := range parameters {
		if fc, ok := fcOutputs[paramName]; ok {
			parameters[paramName] = cloudformation.GetAtt(fc.resource, "Outputs."+fc.outputName)
			delete(defaultParameters, paramName)
		}
	}

	nestedStack := &nestedcloudformation.Stack{
		Parameters: parameters,
		TemplateURL: cloudformation.Join("", []any{
			templateURL,
		}),
		Tags: t.apply(nil, strings.ToLower(logicalID)),
	}

	var dependsOn []string
	if prevLogicalID != "" {
		dependsOn = append(dependsOn, prevLogicalID)
	} else {
		dependsOn = append(dependsOn, "VPC")
		if !tpl.cfg.UseLocalRunners {
			dependsOn = append(dependsOn, "RunnerAutoScalingGroup")
		}
	}
	nestedStack.AWSCloudFormationDependsOn = dependsOn

	return nestedStack, defaultParameters, templateOutputs, nil
}
