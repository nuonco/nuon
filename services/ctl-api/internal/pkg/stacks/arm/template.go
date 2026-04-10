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
	tmpl := &ARMTemplate{
		Schema:         "https://schema.management.azure.com/schemas/2019-04-01/deploymentTemplate.json#",
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

	// Build VNet linked deployment (or use default inline)
	vnetDeployment, vnetParams, err := t.getVNetLinkedDeployment(inp)
	if err != nil {
		return nil, err
	}
	tmpl.Resources = append(tmpl.Resources, vnetDeployment)
	for k, v := range vnetParams {
		tmpl.Parameters[k] = v
	}

	// Key Vault (inline resource, depends on VNet for service endpoints)
	tmpl.Resources = append(tmpl.Resources, t.getKeyVaultResource(inp))

	// Runner linked deployment (or use default inline)
	if !t.cfg.UseLocalRunners {
		runnerDeployment, runnerParams, err := t.getRunnerLinkedDeployment(inp)
		if err != nil {
			return nil, err
		}
		tmpl.Resources = append(tmpl.Resources, runnerDeployment)
		for k, v := range runnerParams {
			tmpl.Parameters[k] = v
		}
	}

	// Phone home deployment script
	tmpl.Resources = append(tmpl.Resources, t.getPhoneHomeResource(inp))

	// Custom role subscription-level deployment (depends on runner VMSS)
	if !t.cfg.UseLocalRunners {
		tmpl.Resources = append(tmpl.Resources, t.getCustomRoleDeployment(inp))
	}

	// Custom linked deployments
	if len(inp.AppCfg.StackConfig.CustomNestedStacks) > 0 {
		customResources, customParams, customIdentities, err := t.getCustomLinkedDeployments(inp)
		if err != nil {
			return nil, err
		}
		tmpl.Resources = append(tmpl.Resources, customResources...)
		for k, v := range customParams {
			tmpl.Parameters[k] = v
		}

		// Create subscription-level role assignments for any managed
		// identities declared in custom nested stacks. This must live in the
		// parent template because ARM does not support subscription-level
		// nested deployments inside linked deployments.
		for _, id := range customIdentities {
			tmpl.Resources = append(tmpl.Resources, t.getCustomDeploymentRoleAssignment(id))
		}
	}

	// Add standard outputs (VNet, subnets, key vault)
	t.addStandardOutputs(tmpl)

	return tmpl, nil
}

func (t *Templates) addStandardOutputs(tmpl *ARMTemplate) {
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
		Value: "[resourceId('Microsoft.KeyVault/vaults', take(format('{0}', parameters('nuonInstallID')), 24))]",
	}
}
