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

	// Nuon-managed values. At resource-group scope they are parameters, which is what
	// every existing install renders. At subscription scope they are variables
	// instead: the portal builds its deployment form from a template's parameters and
	// gives no way to hide one, so a parameter here is an editable field in front of
	// the customer. None of these is customer-configurable.
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

	// deployTimestamp has to stay a parameter at both scopes: utcNow() is only legal
	// in a parameter's default value. That is load-bearing rather than cosmetic — it
	// is what re-triggers the phone-home script when the same template is applied
	// again, e.g. retrying a failed deploy.
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

	// At subscription scope there is no ambient resource group, so the group Nuon's
	// own resources live in becomes a named contract instead. The value is the name
	// customers create by hand at resource-group scope, so nothing downstream of the
	// phone-home changes.
	if scope.subscription {
		tmpl.Variables[installRGVarName] = installResourceGroupName(inp.Install.ID)
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

	// The Key Vault and its secrets, at subscription scope only — see
	// getKeyVaultResources for why they cannot stay a customer prerequisite there.
	tmpl.Resources = append(tmpl.Resources, t.getKeyVaultResources(inp, scope)...)
	for name, p := range azureSecretParameters(inp, scope) {
		tmpl.Parameters[name] = p
	}

	// Build VNet linked deployment (or use default inline)
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
	tmpl.Resources = append(tmpl.Resources, t.getPhoneHomeResources(inp, customOutputs, vnetExtraOutputs, scope)...)

	// Add standard outputs (VNet, subnets, key vault)
	t.addStandardOutputs(tmpl, inp, scope)

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
		"nuonInstallID": {typ: "string", value: scope.nuonIDRef("nuonInstallID")},
		"principalId":   {typ: "string", value: "[reference('runnerDeployment').outputs.vmssPrincipalId.value]"},
	}, grants, nil)

	// The grants read the runner's identity and assign a role on the Key Vault, so
	// the wrapper waits on both in addition to the resource group.
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

	// VNet outputs - reference linked deployment outputs
	tmpl.Outputs["vnetId"] = ARMOutput{
		Type:  "string",
		Value: fmt.Sprintf("[reference('%s').outputs.vnetId.value]", vnetDeployment),
	}
	tmpl.Outputs["vnetName"] = ARMOutput{
		Type:  "string",
		Value: fmt.Sprintf("[reference('%s').outputs.vnetName.value]", vnetDeployment),
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
			Value: fmt.Sprintf("[reference('%s').outputs.%s.value]", vnetDeployment, subnet),
		}
	}
	// Key Vault outputs
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
