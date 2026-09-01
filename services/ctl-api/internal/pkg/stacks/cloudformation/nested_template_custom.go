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
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks"
)

var logicalIDRegexp = regexp.MustCompile(`[^A-Za-z0-9]`)

// AWSCustomStacksOnlyContractParams are the frozen top-level parameter names
// the custom-stacks-only template always declares, standing in for the
// VPC/runner nested stack outputs that don't exist in this mode.
var AWSCustomStacksOnlyContractParams = []string{"VPC", "RunnerSubnet", "PublicSubnets", "PrivateSubnets"}

type customNestedStackOutput struct {
	Name       string
	OutputKeys []string
}

type customNestedStackResult struct {
	resources              map[string]*nestedcloudformation.Stack
	params                 map[string]cloudformation.Parameter
	paramGroups            []map[string]any
	stackOutputs           map[string]customNestedStackOutput
	lastLogicalID          string
	inputParameters        map[string]map[string]string
	nameMatchedInputParams map[string]bool
}

func sanitizeLogicalID(name string) string {
	camel := strcase.ToCamel(name)
	return logicalIDRegexp.ReplaceAllString(camel, "")
}

func (tpl *Templates) getCustomNestedStacks(inp *stacks.TemplateInput, t tagBuilder, existingResourceKeys map[string]bool) (*customNestedStackResult, error) {
	if len(inp.AppCfg.StackConfig.CustomNestedStacks) == 0 {
		return &customNestedStackResult{
			resources:    map[string]*nestedcloudformation.Stack{},
			params:       map[string]cloudformation.Parameter{},
			paramGroups:  nil,
			stackOutputs: map[string]customNestedStackOutput{},
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

	// CustomStacksOnly mode has no VPC/runner nested stack to fetch outputs from,
	// so first-class seeding is skipped in favor of the contract-parameter wiring
	// in buildCustomNestedStack.
	fcOutputs := map[string]firstClassOutput{}
	if !inp.CustomStacksOnly {
		firstClassStacks := map[string]string{
			"VPC": inp.AppCfg.StackConfig.VPCNestedTemplateURL,
		}
		if _, hasASG := existingResourceKeys["RunnerAutoScalingGroup"]; hasASG {
			firstClassStacks["RunnerAutoScalingGroup"] = inp.AppCfg.StackConfig.RunnerNestedTemplateURL
		}
		fcOutputs = tpl.extractFirstClassOutputs(firstClassStacks)
	}

	result := &customNestedStackResult{
		resources:              map[string]*nestedcloudformation.Stack{},
		params:                 map[string]cloudformation.Parameter{},
		stackOutputs:           map[string]customNestedStackOutput{},
		inputParameters:        map[string]map[string]string{},
		nameMatchedInputParams: map[string]bool{},
	}

	allParamNames := map[string]string{}
	prevLogicalID := ""

	for i, stack := range sorted {
		if stack.Name == "" {
			return nil, fmt.Errorf("custom_nested_stacks[%d]: name is required", i)
		}
		// A nested stack is generated from a pre-hosted remote template_url, or
		// from contents that have been uploaded to S3 (ContentsHash set). A
		// local-path template_url with no uploaded contents means the
		// sync_custom_stacks upload has not finished yet (status pending);
		// generating now would fall back to a path CloudFormation cannot
		// resolve.
		isRemoteURL := strings.HasPrefix(stack.TemplateURL, "http://") || strings.HasPrefix(stack.TemplateURL, "https://")
		if stack.ContentsHash == "" && !isRemoteURL {
			return nil, fmt.Errorf("custom_nested_stacks[%d] (%s): template not ready (status %q); contents have not been uploaded to S3 yet", i, stack.Name, stack.Status)
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

		nestedStack, defaultParams, templateOutputs, hoistedParams, installInputParams, err := tpl.buildCustomNestedStack(inp, stack, t, logicalID, prevLogicalID, fcOutputs, allParamNames)
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

		// hoistedParams (explicit config) need a fresh top-level String parameter
		// declared here, tracked in allParamNames since two stacks hoisting the
		// same name is a collision.
		if len(hoistedParams) > 0 || len(installInputParams) > 0 {
			merged := make(map[string]string, len(hoistedParams)+len(installInputParams))
			maps.Copy(merged, hoistedParams)
			maps.Copy(merged, installInputParams)
			result.inputParameters[stack.Name] = merged
			for topLevelParamName := range installInputParams {
				result.nameMatchedInputParams[topLevelParamName] = true
			}
		}
		for topLevelParamName := range hoistedParams {
			allParamNames[topLevelParamName] = stack.Name
			result.params[topLevelParamName] = cloudformation.Parameter{Type: "String"}
		}

		if len(defaultParams) > 0 {
			result.paramGroups = append(result.paramGroups, map[string]any{
				"Label": map[string]any{
					"default": stack.Name,
				},
				"Parameters": pkggenerics.MapToKeys(defaultParams),
			})
		}

		outputKeys := make([]string, 0, len(templateOutputs))
		for outputName := range templateOutputs {
			outputKeys = append(outputKeys, outputName)
			if _, isFirstClass := fcOutputs[outputName]; !isFirstClass {
				fcOutputs[outputName] = firstClassOutput{
					resource:   logicalID,
					outputName: outputName,
				}
			}
		}
		if len(outputKeys) > 0 {
			result.stackOutputs[logicalID] = customNestedStackOutput{
				Name:       stack.Name,
				OutputKeys: outputKeys,
			}
		}

		prevLogicalID = logicalID
	}

	result.lastLogicalID = prevLogicalID

	return result, nil
}

func (tpl *Templates) buildCustomNestedStack(inp *stacks.TemplateInput, stack config.CustomNestedStack, t tagBuilder, logicalID string, prevLogicalID string, fcOutputs map[string]firstClassOutput, reservedParamNames map[string]string) (*nestedcloudformation.Stack, map[string]cloudformation.Parameter, map[string]struct{}, map[string]string, map[string]string, error) {
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
		return nil, nil, nil, nil, nil, err
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

	// Explicit parameter values. These arrive already rendered -- see
	// config.RenderCustomNestedStackParameters, called when the install stack
	// version is generated -- so they are used verbatim. The install-input
	// reference form is still resolved here as a fallback for callers that read the
	// config without rendering it first.
	hoistedParams := map[string]string{}
	installInputParams := map[string]string{}
	explicitlyConfigured := map[string]bool{}

	for cfnParamName, templateValue := range stack.Parameters {
		explicitlyConfigured[cfnParamName] = true
		if inp.CustomStacksOnly {
			if unrendered, ok := inp.UnrenderedCustomStackParameters[stack.Name][cfnParamName]; ok {
				if inputName, err := config.ParseInstallInputReference(unrendered); err == nil {
					topLevelParamName := logicalID + cfnParamName
					_, contractCollision := roleParams[topLevelParamName]
					_, reservedCollision := reservedParamNames[topLevelParamName]
					contractCollision = contractCollision || slices.Contains(AWSCustomStacksOnlyContractParams, topLevelParamName)
					if !contractCollision && !reservedCollision {
						hoistedParams[topLevelParamName] = inputName
						parameters[cfnParamName] = cloudformation.Ref(topLevelParamName)
						delete(defaultParameters, cfnParamName)
						continue
					}
				}
			}
		}

		resolved := templateValue
		if inputName, err := config.ParseInstallInputReference(templateValue); err == nil {
			resolved = ""
			if inp.Install.CurrentInstallInputs != nil {
				if val, ok := inp.Install.CurrentInstallInputs.Values[inputName]; ok && val != nil {
					resolved = *val
				}
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

	if inp.CustomStacksOnly {
		// No VPC/runner nested stack exists in this mode, so a parameter matching
		// one of the frozen contract names is wired to the top-level template
		// parameter of the same name instead of an fcOutputs GetAtt.
		for paramName := range parameters {
			if slices.Contains(AWSCustomStacksOnlyContractParams, paramName) {
				parameters[paramName] = cloudformation.Ref(paramName)
				delete(defaultParameters, paramName)
			}
		}

		// A custom stack's own template parameter can share a name with an
		// install input's CloudFormationStackParamName. getAWSCustomStacksOnlyTemplate
		// already declares that top-level parameter, so it's wired here the same
		// way rather than requiring a Default in the vendor's template.
		installInputNameByParam := make(map[string]string, len(inp.AppCfg.InputConfig.AppInputs))
		for _, appInput := range inp.AppCfg.InputConfig.AppInputs {
			if appInput.Source != app.AppInputSourceCustomer {
				continue
			}
			installInputNameByParam[appInput.CloudFormationStackParamName] = appInput.Name
		}
		for paramName := range parameters {
			if explicitlyConfigured[paramName] {
				continue
			}
			if inputName, ok := installInputNameByParam[paramName]; ok {
				parameters[paramName] = cloudformation.Ref(paramName)
				delete(defaultParameters, paramName)
				installInputParams[paramName] = inputName
			}
		}

		// Anything still unbound and without a template default can never
		// resolve. Fail loudly here rather than deploy a broken template.
		for paramName, cfnParam := range defaultParameters {
			if cfnParam.Default != nil {
				continue
			}
			return nil, nil, nil, nil, nil, fmt.Errorf(
				"parameter %q has no contract match (contract: %s), no explicit config value, and no template default",
				paramName, strings.Join(AWSCustomStacksOnlyContractParams, ", "),
			)
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
	switch {
	case prevLogicalID != "":
		dependsOn = append(dependsOn, prevLogicalID)
	case inp.CustomStacksOnly:
		// No VPC or runner ASG resource exists in this template, so the first
		// custom stack has nothing to depend on.
	default:
		dependsOn = append(dependsOn, "VPC")
		if !tpl.cfg.UseLocalRunners {
			dependsOn = append(dependsOn, "RunnerAutoScalingGroup")
		}
	}
	nestedStack.AWSCloudFormationDependsOn = dependsOn

	return nestedStack, defaultParameters, templateOutputs, hoistedParams, installInputParams, nil
}
