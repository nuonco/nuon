package arm

import (
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
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
	if scope.subscription {
		// The scope now threads through the renderer, but the root still declares
		// resource-group-scoped resources directly (UAMIs, role assignments, the
		// phone-home deploymentScripts) and never declares installResourceGroupName,
		// so rendering here would emit a template ARM rejects. Fail at generation
		// with a message that says so rather than at the customer's az stack sub
		// create with an opaque InvalidTemplate. Remove once the root template is
		// built out.
		return nil, fmt.Errorf(
			"deployment_scope %q is accepted by app config but the subscription-scoped root template is not implemented yet; unset it or use %q",
			app.StackDeploymentScopeSubscription, app.StackDeploymentScopeResourceGroup,
		)
	}

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

	// Build VNet linked deployment (or use default inline)
	vnetDeployment, vnetParams, err := t.getVNetLinkedDeployment(inp)
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
		runnerDeployment, runnerParams, err := t.getRunnerLinkedDeployment(inp, operationIDs)
		if err != nil {
			return nil, err
		}
		tmpl.Resources = append(tmpl.Resources, runnerDeployment)
		for k, v := range runnerParams {
			tmpl.Parameters[k] = v
		}
	}

	// Legacy broad grants on the system identity, only when per-operation
	// identities are not in use.
	if !t.cfg.UseLocalRunners && !useOperationIdentities {
		tmpl.Resources = append(tmpl.Resources, t.getVMSSRoleAssignments()...)
		tmpl.Resources = append(tmpl.Resources, t.getCustomRoleDeployment(inp, scope))
	}

	// Key Vault Secrets User and ACR pull/push stay on the system identity:
	// secret-sync and image-sync run as the ambient identity.
	if !t.cfg.UseLocalRunners {
		tmpl.Resources = append(tmpl.Resources, t.getKeyVaultRoleAssignment())
		tmpl.Resources = append(tmpl.Resources, t.getACRRoleAssignments()...)
	}

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
	tmpl.Resources = append(tmpl.Resources, t.getPhoneHomeResource(inp, customOutputs, scope))

	// Add standard outputs (VNet, subnets, key vault)
	t.addStandardOutputs(tmpl, scope)

	return tmpl, nil
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
