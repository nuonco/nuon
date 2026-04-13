package arm

import (
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks"
)

func (t *Templates) getKeyVaultResource(inp *stacks.TemplateInput) map[string]any {
	_ = inp // inp reserved for future secret injection
	return map[string]any{
		"type":       "Microsoft.KeyVault/vaults",
		"apiVersion": "2023-02-01",
		"name":       "[take(format('{0}', parameters('nuonInstallID')), 24)]",
		"location":   "[parameters('location')]",
		"tags":       "[variables('commonTags')]",
		"dependsOn":  []string{"vnetDeployment"},
		"properties": map[string]any{
			"enabledForDeployment":         true,
			"enabledForTemplateDeployment": true,
			"enabledForDiskEncryption":     true,
			"tenantId":                     "[subscription().tenantId]",
			"sku": map[string]any{
				"name":   "standard",
				"family": "A",
			},
			"accessPolicies": []any{},
			"networkAcls": map[string]any{
				"defaultAction": "Deny",
				"bypass":        "AzureServices",
				"virtualNetworkRules": []map[string]any{
					{
						"id": "[reference('vnetDeployment').outputs.runnerSubnetId.value]",
					},
					{
						"id": "[reference('vnetDeployment').outputs.privateSubnet1Id.value]",
					},
				},
			},
		},
	}
}
