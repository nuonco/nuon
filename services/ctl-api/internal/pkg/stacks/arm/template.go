package arm

import (
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks"
)

type ARMTemplate struct {
	Schema         string                  `json:"$schema"`
	ContentVersion string                  `json:"contentVersion"`
	Parameters     map[string]ARMParameter `json:"parameters,omitempty"`
	Variables      map[string]any          `json:"variables,omitempty"`
	Resources      []any                   `json:"resources"`
	Outputs        map[string]ARMOutput    `json:"outputs,omitempty"`
}

type ARMParameter struct {
	Type          string                `json:"type"`
	DefaultValue  any                   `json:"defaultValue,omitempty"`
	AllowedValues []any                 `json:"allowedValues,omitempty"`
	Metadata      *ARMParameterMetadata `json:"metadata,omitempty"`
}

type ARMParameterMetadata struct {
	Description string `json:"description,omitempty"`
}

type ARMOutput struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

var ReservedParamNames = []string{"nuonInstallID", "nuonOrgID", "nuonAppID", "location", "deployTimestamp", runnerVmSizeParamName}

func (t *Templates) getAzureTemplate(inp *stacks.TemplateInput) (*ARMTemplate, error) {
	scope := scopeFor(inp)

	tmpl := &ARMTemplate{
		Schema:         scope.rootSchema(),
		ContentVersion: "1.0.0.0",
		Parameters:     make(map[string]ARMParameter),
		Variables:      make(map[string]any),
		Resources:      []any{},
		Outputs:        make(map[string]ARMOutput),
	}

	nuonValues := map[string]struct {
		value       string
		description string
	}{
		"nuonInstallID": {inp.Install.ID, "The Nuon Install ID; prefixed to resource names."},
		"nuonOrgID":     {inp.Runner.OrgID, "The Nuon Org ID. Used in tags."},
		"nuonAppID":     {inp.Install.AppID, "The Nuon App ID. Used in tags."},
		locationVarName: {inp.Install.AzureAccount.Location, "The location for all resources."},
	}
	for name, v := range nuonValues {
		if scope.subscription {
			tmpl.Variables[name] = v.value
			continue
		}
		tmpl.Parameters[name] = ARMParameter{
			Type:         "string",
			DefaultValue: v.value,
			Metadata:     &ARMParameterMetadata{Description: v.description},
		}
	}

	tmpl.Parameters["deployTimestamp"] = ARMParameter{
		Type:         "string",
		DefaultValue: "[utcNow()]",
		Metadata:     &ARMParameterMetadata{Description: "Force re-run of deployment scripts on each deploy."},
	}

	tmpl.Variables["commonTags"] = map[string]string{
		"install_nuon_co_id": scope.nuonIDRef("nuonInstallID"),
		"org_nuon_co_id":     scope.nuonIDRef("nuonOrgID"),
		"app_nuon_co_id":     scope.nuonIDRef("nuonAppID"),
	}

	if scope.subscription {
		tmpl.Variables[installRGVarName] = installResourceGroupName(inp.Install.ID)
	}

	operationIDs := azureOperationIdentities(inp.AppCfg)
	useOperationIdentities := len(operationIDs) > 0

	if rg := scope.installRGResource(); rg != nil {
		tmpl.Resources = append(tmpl.Resources, rg)
	}

	tmpl.Resources = append(tmpl.Resources, t.getKeyVaultResources(inp, scope)...)
	for name, p := range azureSecretParameters(inp, scope) {
		tmpl.Parameters[name] = p
	}

	vnetDeployment, vnetParams, vnetExtraOutputs, err := t.getVNetLinkedDeployment(inp, scope)
	if err != nil {
		return nil, err
	}
	tmpl.Resources = append(tmpl.Resources, vnetDeployment)
	for k, v := range vnetParams {
		tmpl.Parameters[k] = v
	}

	if useOperationIdentities {
		tmpl.Resources = append(tmpl.Resources, t.getOperationIdentityResources(operationIDs, scope)...)
	}

	telemetryEndpoint := ""
	if !t.cfg.UseLocalRunners {
		runnerDeployment, runnerParams, err := t.getRunnerLinkedDeployment(inp, operationIDs, scope)
		if err != nil {
			return nil, err
		}
		tmpl.Resources = append(tmpl.Resources, runnerDeployment)
		for k, v := range runnerParams {
			tmpl.Parameters[k] = v
		}
		if _, supported := runnerParams["enableTelemetryIngress"]; supported {
			telemetryEndpoint = "[reference('runnerDeployment').outputs.telemetryEndpoint.value]"
			tmpl.Outputs["telemetryEndpoint"] = ARMOutput{Type: "string", Value: telemetryEndpoint}
		}
	}

	t.appendRunnerGrants(tmpl, inp, scope, useOperationIdentities)

	var customOutputs []customDeploymentOutputs
	if len(inp.AppCfg.StackConfig.CustomNestedStacks) > 0 {
		customResources, customParams, customIdentities, customOutputsMeta, err := t.getCustomLinkedDeployments(inp)
		if err != nil {
			return nil, err
		}
		customOutputs = customOutputsMeta
		tmpl.Resources = append(tmpl.Resources, customResources...)
		for k, v := range customParams {
			tmpl.Parameters[k] = v
		}

		// why: Create subscription-level role assignments for any managed
		// identities declared in custom nested stacks. This must live in the
		// parent template because ARM does not support subscription-level
		// nested deployments inside linked deployments.
		for _, id := range customIdentities {
			tmpl.Resources = append(tmpl.Resources, t.getCustomDeploymentRoleAssignment(id, inp.Install.ID, scope))
		}
	}

	if err := addCustomerInputParameters(tmpl, inp); err != nil {
		return nil, err
	}

	tmpl.Resources = append(tmpl.Resources, t.getPhoneHomeResources(inp, customOutputs, vnetExtraOutputs, scope, telemetryEndpoint)...)

	t.addStandardOutputs(tmpl, inp, scope)

	return tmpl, nil
}

func (t *Templates) appendRunnerGrants(tmpl *ARMTemplate, inp *stacks.TemplateInput, scope armScope, useOperationIdentities bool) {
	if t.cfg.UseLocalRunners {
		return
	}

	legacyGrants := !useOperationIdentities

	if !scope.subscription {
		if legacyGrants {
			tmpl.Resources = append(tmpl.Resources, t.getVMSSRoleAssignments(runnerGrantContextFor(scope))...)
			tmpl.Resources = append(tmpl.Resources, t.getCustomRoleDeployment(inp, scope))
		}
		tmpl.Resources = append(tmpl.Resources, t.getKeyVaultRoleAssignment(runnerGrantContextFor(scope)))
		tmpl.Resources = append(tmpl.Resources, t.getACRRoleAssignments(runnerGrantContextFor(scope))...)
		return
	}

	inner := runnerGrantContextFor(armScope{subscription: true})

	var grants []any
	if legacyGrants {
		grants = append(grants, t.getVMSSRoleAssignments(inner)...)
	}
	grants = append(grants, t.getKeyVaultRoleAssignment(inner))
	grants = append(grants, t.getACRRoleAssignments(inner)...)

	wrapper := scope.wrapInInstallRG(runnerGrantsDeploymentName, map[string]nestedParam{
		"nuonInstallID": {typ: "string", value: scope.nuonIDRef("nuonInstallID")},
		"principalId":   {typ: "string", value: "[reference('runnerDeployment').outputs.vmssPrincipalId.value]"},
	}, grants, nil)

	for _, r := range wrapper {
		dependOn(r.(map[string]any), append([]string{"runnerDeployment"}, scope.keyVaultDependsOn()...))
	}
	tmpl.Resources = append(tmpl.Resources, wrapper...)

	if legacyGrants {
		tmpl.Resources = append(tmpl.Resources, t.getCustomRoleDeployment(inp, scope))
	}
}

func (t *Templates) addStandardOutputs(tmpl *ARMTemplate, inp *stacks.TemplateInput, scope armScope) {
	vnetDeployment := scope.vnetDeploymentName(inp.Install.ID)

	tmpl.Outputs["vnetId"] = ARMOutput{
		Type:  "string",
		Value: fmt.Sprintf("[reference('%s').outputs.vnetId.value]", vnetDeployment),
	}
	tmpl.Outputs["vnetName"] = ARMOutput{
		Type:  "string",
		Value: fmt.Sprintf("[reference('%s').outputs.vnetName.value]", vnetDeployment),
	}
	for _, subnet := range []string{
		"publicSubnet1Id", "publicSubnet1Name",
		"publicSubnet2Id", "publicSubnet2Name",
		"publicSubnet3Id", "publicSubnet3Name",
		"privateSubnet1Id", "privateSubnet1Name",
		"privateSubnet2Id", "privateSubnet2Name",
		"privateSubnet3Id", "privateSubnet3Name",
		"runnerSubnetId", "runnerSubnetName",
	} {
		tmpl.Outputs[subnet] = ARMOutput{
			Type:  "string",
			Value: fmt.Sprintf("[reference('%s').outputs.%s.value]", vnetDeployment, subnet),
		}
	}
	tmpl.Outputs["keyVaultName"] = ARMOutput{
		Type:  "string",
		Value: "[" + scope.keyVaultNameInner() + "]",
	}
	tmpl.Outputs["keyVaultId"] = ARMOutput{
		Type:  "string",
		Value: scope.rgResourceIDExpr("Microsoft.KeyVault/vaults", scope.keyVaultNameInner()),
	}
	tmpl.Outputs["keyVaultUri"] = ARMOutput{
		Type:  "string",
		Value: "[format('https://{0}.vault.azure.net/', " + scope.keyVaultNameInner() + ")]",
	}
}
