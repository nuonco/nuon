package arm

import (
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks"
)

// ARMTemplate represents an Azure Resource Manager deployment template.
type ARMTemplate struct {
	Schema         string                  `json:"$schema"`
	ContentVersion string                  `json:"contentVersion"`
	Parameters     map[string]ARMParameter `json:"parameters,omitempty"`
	Variables      map[string]any          `json:"variables,omitempty"`
	Resources      []any                   `json:"resources"`
	Outputs        map[string]ARMOutput    `json:"outputs,omitempty"`
}

type ARMParameter struct {
	Type         string                `json:"type"`
	DefaultValue any                   `json:"defaultValue,omitempty"`
	Metadata     *ARMParameterMetadata `json:"metadata,omitempty"`
}

type ARMParameterMetadata struct {
	Description string `json:"description,omitempty"`
}

type ARMOutput struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

// ReservedParamNames are always provided by Nuon, never exposed to the customer.
var ReservedParamNames = []string{"nuonInstallID", "nuonOrgID", "nuonAppID", "location", "deployTimestamp"}

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

	// Add Nuon-managed parameters (always present, never customer-facing)
	tmpl.Parameters["nuonInstallID"] = ARMParameter{
		Type:         "string",
		DefaultValue: inp.Install.ID,
		Metadata:     &ARMParameterMetadata{Description: "The Nuon Install ID; prefixed to resource names."},
	}
	tmpl.Parameters["nuonOrgID"] = ARMParameter{
		Type:         "string",
		DefaultValue: inp.Runner.OrgID,
		Metadata:     &ARMParameterMetadata{Description: "The Nuon Org ID. Used in tags."},
	}
	tmpl.Parameters["nuonAppID"] = ARMParameter{
		Type:         "string",
		DefaultValue: inp.Install.AppID,
		Metadata:     &ARMParameterMetadata{Description: "The Nuon App ID. Used in tags."},
	}
	tmpl.Parameters["location"] = ARMParameter{
		Type:         "string",
		DefaultValue: inp.Install.AzureAccount.Location,
		Metadata:     &ARMParameterMetadata{Description: "The location for all resources."},
	}
	tmpl.Parameters["deployTimestamp"] = ARMParameter{
		Type:         "string",
		DefaultValue: "[utcNow()]",
		Metadata:     &ARMParameterMetadata{Description: "Force re-run of deployment scripts on each deploy."},
	}

	// At subscription scope there is no ambient resource group, so the group Nuon's
	// own resources live in becomes a named contract instead. The default is the
	// name customers create by hand at resource-group scope, so nothing downstream
	// of the phone-home changes.
	if scope.subscription {
		tmpl.Parameters[installRGParamName] = ARMParameter{
			Type:         "string",
			DefaultValue: "[format('{0}-rg', parameters('nuonInstallID'))]",
			Metadata:     &ARMParameterMetadata{Description: "Resource group the install's Nuon-managed resources are created in."},
		}
	}

	// Add common variables
	tmpl.Variables["commonTags"] = map[string]string{
		"install_nuon_co_id": "[parameters('nuonInstallID')]",
		"org_nuon_co_id":     "[parameters('nuonOrgID')]",
		"app_nuon_co_id":     "[parameters('nuonAppID')]",
	}

	// When the app declares Azure roles, deploy work runs as per-operation
	// identities and the system identity is stripped of deploy grants.
	operationIDs := azureOperationIdentities(inp.AppCfg)
	useOperationIdentities := len(operationIDs) > 0

	// The install resource group has to exist before anything targets it, so it is
	// declared ahead of every other resource.
	if rg := scope.installRGResource(); rg != nil {
		tmpl.Resources = append(tmpl.Resources, rg)
	}

	// Build VNet linked deployment (or use default inline)
	vnetDeployment, vnetParams, err := t.getVNetLinkedDeployment(inp, scope)
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

	// Runner linked deployment (or use default inline)
	if !t.cfg.UseLocalRunners {
		runnerDeployment, runnerParams, err := t.getRunnerLinkedDeployment(inp, operationIDs, scope)
		if err != nil {
			return nil, err
		}
		tmpl.Resources = append(tmpl.Resources, runnerDeployment)
		for k, v := range runnerParams {
			tmpl.Parameters[k] = v
		}
	}

	t.appendRunnerGrants(tmpl, inp, scope, useOperationIdentities)

	// Custom linked deployments (before phone home, which reports their outputs)
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

		// Create subscription-level role assignments for any managed
		// identities declared in custom nested stacks. This must live in the
		// parent template because ARM does not support subscription-level
		// nested deployments inside linked deployments.
		for _, id := range customIdentities {
			tmpl.Resources = append(tmpl.Resources, t.getCustomDeploymentRoleAssignment(id, scope))
		}
	}

	// Phone home deployment script
	tmpl.Resources = append(tmpl.Resources, t.getPhoneHomeResources(inp, customOutputs, scope)...)

	// Add standard outputs (VNet, subnets, key vault)
	t.addStandardOutputs(tmpl, scope)

	return tmpl, nil
}

// appendRunnerGrants emits the role assignments held by the runner's system
// identity. They are all resource-group-scoped, so at subscription scope they move
// into the install resource group together as runnerGrantsDeployment.
//
// The custom role deployment stays in the root either way: it targets the
// subscription, which ARM will not allow inside another nested deployment.
func (t *Templates) appendRunnerGrants(tmpl *ARMTemplate, inp *stacks.TemplateInput, scope armScope, useOperationIdentities bool) {
	if t.cfg.UseLocalRunners {
		return
	}

	// Legacy broad grants on the system identity, only when per-operation
	// identities are not in use.
	legacyGrants := !useOperationIdentities

	if !scope.subscription {
		// Emission order is load-bearing for the byte-identical guarantee.
		if legacyGrants {
			tmpl.Resources = append(tmpl.Resources, t.getVMSSRoleAssignments(runnerGrantContextFor(scope))...)
			tmpl.Resources = append(tmpl.Resources, t.getCustomRoleDeployment(inp, scope))
		}
		// Key Vault Secrets User and ACR pull/push stay on the system identity:
		// secret-sync and image-sync run as the ambient identity.
		tmpl.Resources = append(tmpl.Resources, t.getKeyVaultRoleAssignment(runnerGrantContextFor(scope)))
		tmpl.Resources = append(tmpl.Resources, t.getACRRoleAssignments(runnerGrantContextFor(scope))...)
		return
	}

	// Inside the wrapper the grants are back at resource-group scope, so their
	// guid() names are unchanged; only the principal and the dependency move.
	inner := runnerGrantContextFor(armScope{subscription: true})

	var grants []any
	if legacyGrants {
		grants = append(grants, t.getVMSSRoleAssignments(inner)...)
	}
	grants = append(grants, t.getKeyVaultRoleAssignment(inner))
	grants = append(grants, t.getACRRoleAssignments(inner)...)

	wrapper := scope.wrapInInstallRG(runnerGrantsDeploymentName, map[string]nestedParam{
		"nuonInstallID": {typ: "string", value: "[parameters('nuonInstallID')]"},
		"principalId":   {typ: "string", value: "[reference('runnerDeployment').outputs.vmssPrincipalId.value]"},
	}, grants)

	// The grants read the runner's identity, so the wrapper waits on the runner in
	// addition to the resource group.
	for _, r := range wrapper {
		dependOn(r.(map[string]any), []string{"runnerDeployment"})
	}
	tmpl.Resources = append(tmpl.Resources, wrapper...)

	if legacyGrants {
		tmpl.Resources = append(tmpl.Resources, t.getCustomRoleDeployment(inp, scope))
	}
}

func (t *Templates) addStandardOutputs(tmpl *ARMTemplate, scope armScope) {
	// VNet outputs - reference linked deployment outputs
	tmpl.Outputs["vnetId"] = ARMOutput{
		Type:  "string",
		Value: "[reference('vnetDeployment').outputs.vnetId.value]",
	}
	tmpl.Outputs["vnetName"] = ARMOutput{
		Type:  "string",
		Value: "[reference('vnetDeployment').outputs.vnetName.value]",
	}
	// Subnet outputs
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
			Value: fmt.Sprintf("[reference('vnetDeployment').outputs.%s.value]", subnet),
		}
	}
	// Key Vault outputs
	tmpl.Outputs["keyVaultName"] = ARMOutput{
		Type:  "string",
		Value: "[take(format('{0}', parameters('nuonInstallID')), 24)]",
	}
	tmpl.Outputs["keyVaultId"] = ARMOutput{
		Type:  "string",
		Value: scope.rgResourceIDExpr("Microsoft.KeyVault/vaults", keyVaultNameInner),
	}
	tmpl.Outputs["keyVaultUri"] = ARMOutput{
		Type:  "string",
		Value: "[format('https://{0}.vault.azure.net/', take(format('{0}', parameters('nuonInstallID')), 24))]",
	}
}
