package cloudformation

import (
	"fmt"
	"maps"
	"slices"

	"github.com/awslabs/goformation/v7/cloudformation"
	"github.com/iancoleman/strcase"

	pkggenerics "github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks"
)

// getAWSCustomStacksOnlyTemplate renders a CloudFormation template containing
// only the vendor's custom_nested_stacks.
func (t *Templates) getAWSCustomStacksOnlyTemplate(inp *stacks.TemplateInput) (*cloudformation.Template, error) {
	tmpl := cloudformation.NewTemplate()

	tb := tagBuilder{
		installID:  inp.Install.ID,
		orgID:      inp.Install.OrgID,
		appID:      inp.Install.AppID,
		additional: generics.ToStringMap(inp.Settings.AWSTags),
	}

	for _, name := range AWSCustomStacksOnlyContractParams {
		desc := fmt.Sprintf("Nuon-owned contract parameter %s, supplied by the Terraform module from natively-created VPC/runner resources.", name)
		tmpl.Parameters[name] = cloudformation.Parameter{
			Type:        "String",
			Description: &desc,
		}
	}

	customResult, err := t.getCustomNestedStacks(inp, tb, map[string]bool{})
	if err != nil {
		return nil, err
	}
	if len(customResult.resources) == 0 {
		return nil, fmt.Errorf("custom stacks only template requested but app config has no custom_nested_stacks configured")
	}

	for k, v := range customResult.resources {
		tmpl.Resources[k] = v
	}
	maps.Copy(tmpl.Parameters, customResult.params)

	paramLabels := map[string]any{}
	installGroupParameters := t.getInstallInputGroupParameters(inp)
	for _, installGroupParameter := range installGroupParameters {
		maps.Copy(tmpl.Parameters, installGroupParameter)
	}
	installGroupParamLabels := t.getInstallInputGroupParamLable(inp)
	for _, installGroupParamLabel := range installGroupParamLabels {
		maps.Copy(paramLabels, installGroupParamLabel)
	}

	outputs, err := customStacksOnlyOutputs(customResult)
	if err != nil {
		return nil, err
	}

	allStackNames := make([]string, 0, len(inp.AppCfg.StackConfig.CustomNestedStacks))
	for _, stack := range inp.AppCfg.StackConfig.CustomNestedStacks {
		allStackNames = append(allStackNames, stack.Name)
	}
	if err := verifyCustomStacksOnlyOutputsUnambiguous(customResult, allStackNames); err != nil {
		return nil, err
	}
	if err := verifyCustomStacksOnlyInputParametersUnambiguous(customResult, allStackNames); err != nil {
		return nil, err
	}

	tmpl.Outputs = outputs

	pgs := append([]map[string]any{}, customResult.paramGroups...)
	for groupName, installGroupParameter := range installGroupParameters {
		pgs = append(pgs, map[string]any{
			"Label": map[string]any{
				"default": "Install Inputs: " + strcase.ToCamel(groupName),
			},
			"Parameters": pkggenerics.MapToKeys(installGroupParameter),
		})
	}
	if len(pgs) > 0 || len(paramLabels) > 0 {
		tmpl.Metadata["AWS::CloudFormation::Interface"] = map[string]any{
			"ParameterLabels": paramLabels,
			"ParameterGroups": pgs,
		}
	}

	// This is not used by Cloudformation. It's here to support mapping nested
	// stack inputs to app inputs in the parent stack.
	// Callers must strip it before treating tmpl as the deployable artifact.
	if len(customResult.inputParameters) > 0 {
		tmpl.Metadata[customStacksInputParametersMetadataKey] = customResult.inputParameters
	}

	return tmpl, nil
}

// customStacksInputParametersMetadataKey namespaces the nested input key above
// so it can never collide with a real CloudFormation metadata section.
const customStacksInputParametersMetadataKey = "Nuon::CustomStacksInputParameters"

// ExtractAndStripCustomStacksInputParameters pulls the hoisted-parameter
// mapping getAWSCustomStacksOnlyTemplate attaches to Template.Metadata and
// removes it so it never reaches the deployed JSON.
func ExtractAndStripCustomStacksInputParameters(tmpl *cloudformation.Template) map[string]map[string]string {
	v, ok := tmpl.Metadata[customStacksInputParametersMetadataKey]
	if !ok {
		return nil
	}
	delete(tmpl.Metadata, customStacksInputParametersMetadataKey)
	m, _ := v.(map[string]map[string]string)
	return m
}

// customStacksOnlyOutputs emits one top-level CFN output per custom-stack
// output, keyed by the stack's sanitized logical ID directly concatenated
// with the declared output key.
func customStacksOnlyOutputs(result *customNestedStackResult) (map[string]cloudformation.Output, error) {
	outputs := map[string]cloudformation.Output{}
	for logicalID, info := range result.stackOutputs {
		for _, outputKey := range info.OutputKeys {
			name := logicalID + outputKey
			if _, exists := outputs[name]; exists {
				return nil, fmt.Errorf("custom stacks only template: output name %q collides across stacks", name)
			}
			outputs[name] = cloudformation.Output{
				Value: cloudformation.GetAtt(logicalID, "Outputs."+outputKey),
			}
		}
	}
	return outputs, nil
}

// customStacksOnlyOutputPair identifies a single declared custom-stack output by
// its original (unsanitized) stack name and its template-declared output key.
type customStacksOnlyOutputPair struct {
	stackName string
	outputKey string
}

// verifyCustomStacksOnlyOutputsUnambiguous catches stack name collisions
// between both custom stacks, and custom stack outputs.
func verifyCustomStacksOnlyOutputsUnambiguous(result *customNestedStackResult, allStackNames []string) error {
	flatOutputs := make(map[string]string, len(result.stackOutputs))
	want := make(map[string]customStacksOnlyOutputPair, len(result.stackOutputs))
	for logicalID, info := range result.stackOutputs {
		for _, outputKey := range info.OutputKeys {
			flatName := logicalID + outputKey
			flatOutputs[flatName] = flatName
			want[flatName] = customStacksOnlyOutputPair{stackName: info.Name, outputKey: outputKey}
		}
	}

	parsed := SplitCustomStacksOnlyOutputs(flatOutputs, allStackNames)
	got := make(map[string]customStacksOnlyOutputPair, len(flatOutputs))
	for stackName, stackResult := range parsed {
		for outputKey, flatName := range stackResult["outputs"] {
			got[flatName] = customStacksOnlyOutputPair{stackName: stackName, outputKey: outputKey}
		}
	}

	for flatName, wantPair := range want {
		if gotPair, ok := got[flatName]; ok && gotPair == wantPair {
			continue
		}
		gotPair := got[flatName]
		return fmt.Errorf(
			"custom stacks only template: output %q of custom stack %q renders as CloudFormation output %q, "+
				"which is indistinguishable from output %q of custom stack %q — "+
				"rename one of the two custom stacks, or one of the two template outputs, to resolve this",
			wantPair.outputKey, wantPair.stackName, flatName, gotPair.outputKey, gotPair.stackName,
		)
	}
	return nil
}

// splitFlatNamesByLogicalID reconstructs { "<stack_name>": { "<suffix>": value } }
// from a flat name -> value map, matching each flat name against the longest
// known logical-ID prefix first (so "Foo" never wins over "FooBar").
func splitFlatNamesByLogicalID(flatNames map[string]string, stackNames []string) map[string]map[string]string {
	logicalIDs := make([]string, 0, len(stackNames))
	logicalIDToName := make(map[string]string, len(stackNames))
	for _, name := range stackNames {
		logicalID := sanitizeLogicalID(name)
		logicalIDs = append(logicalIDs, logicalID)
		logicalIDToName[logicalID] = name
	}
	slices.SortFunc(logicalIDs, func(a, b string) int {
		return len(b) - len(a)
	})

	result := map[string]map[string]string{}
	for flatName, value := range flatNames {
		for _, logicalID := range logicalIDs {
			if len(flatName) <= len(logicalID) || flatName[:len(logicalID)] != logicalID {
				continue
			}
			key := flatName[len(logicalID):]
			name := logicalIDToName[logicalID]
			if result[name] == nil {
				result[name] = map[string]string{}
			}
			result[name][key] = value
			break
		}
	}
	return result
}

// SplitCustomStacksOnlyOutputs reconstructs
// { "<stack_name>": { "outputs": { "<key>": value } } } from the flat output
// map a deployed custom-stacks-only stack returns.
func SplitCustomStacksOnlyOutputs(flatOutputs map[string]string, stackNames []string) map[string]map[string]map[string]string {
	split := splitFlatNamesByLogicalID(flatOutputs, stackNames)
	result := make(map[string]map[string]map[string]string, len(split))
	for name, outputs := range split {
		result[name] = map[string]map[string]string{"outputs": outputs}
	}
	return result
}

// customStacksOnlyInputParameterPair identifies a single hoisted parameter by
// its original (unsanitized) stack name and its CFN parameter name on that
// stack's nested template.
type customStacksOnlyInputParameterPair struct {
	stackName string
	paramName string
}

// verifyCustomStacksOnlyInputParametersUnambiguous is the input-parameter
// analog of verifyCustomStacksOnlyOutputsUnambiguous, guarding the same
// concatenation scheme against the same prefix-collision hazard.
func verifyCustomStacksOnlyInputParametersUnambiguous(result *customNestedStackResult, allStackNames []string) error {
	flatParams := make(map[string]string)
	want := make(map[string]customStacksOnlyInputParameterPair)
	for stackName, hoisted := range result.inputParameters {
		for topLevelParamName := range hoisted {
			if result.nameMatchedInputParams[topLevelParamName] {
				continue
			}
			flatParams[topLevelParamName] = topLevelParamName
			want[topLevelParamName] = customStacksOnlyInputParameterPair{stackName: stackName, paramName: topLevelParamName}
		}
	}

	split := splitFlatNamesByLogicalID(flatParams, allStackNames)
	got := make(map[string]customStacksOnlyInputParameterPair, len(flatParams))
	for stackName, stackParams := range split {
		for _, flatName := range stackParams {
			got[flatName] = customStacksOnlyInputParameterPair{stackName: stackName, paramName: flatName}
		}
	}

	for flatName, wantPair := range want {
		if gotPair, ok := got[flatName]; ok && gotPair.stackName == wantPair.stackName {
			continue
		}
		gotPair := got[flatName]
		return fmt.Errorf(
			"custom stacks only template: hoisted parameter %q of custom stack %q renders as CloudFormation parameter %q, "+
				"which is indistinguishable from a parameter of custom stack %q — "+
				"rename one of the two custom stacks to resolve this",
			wantPair.paramName, wantPair.stackName, flatName, gotPair.stackName,
		)
	}
	return nil
}
