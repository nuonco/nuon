package arm

import (
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks"
)

func (t *Templates) getPhoneHomeResource(inp *stacks.TemplateInput) map[string]any {
	phoneHomeURL := inp.CloudFormationStackVersion.PhoneHomeURL

	scriptContent := `#!/bin/bash

PAYLOAD=$(cat << EOF
{
  "request_type": "Create",
  "phone_home_type": "azure",
  "resource_group_id": "$RESOURCE_GROUP_ID",
  "resource_group_name": "$RESOURCE_GROUP_NAME",
  "resource_group_location": "$RESOURCE_GROUP_LOCATION",
  "network_id": "$VNET_ID",
  "network_name": "$VNET_NAME",
  "key_vault_id": "$KEY_VAULT_ID",
  "key_vault_name": "$KEY_VAULT_NAME",
  "public_subnet_ids": "$PUBLIC_SUBNET_IDS_CSV",
  "public_subnet_names": "$PUBLIC_SUBNET_NAMES_CSV",
  "private_subnet_ids": "$PRIVATE_SUBNET_IDS_CSV",
  "private_subnet_names": "$PRIVATE_SUBNET_NAMES_CSV",
  "subscription_id": "$SUBSCRIPTION_ID",
  "subscription_tenant_id": "$SUBSCRIPTION_TENANT_ID"
}
EOF
)

curl -X POST \
  "` + phoneHomeURL + `" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json" \
  -d "$PAYLOAD" \
  --fail \
  --silent \
  --show-error

if [ $? -eq 0 ]; then
  echo "Phone home request sent successfully"
else
  echo "Failed to send phone home request"
  exit 1
fi
`

	return map[string]any{
		"type":       "Microsoft.Resources/deploymentScripts",
		"apiVersion": "2023-08-01",
		"name":       "[format('{0}-phone-home-script', parameters('nuonInstallID'))]",
		"location":   "[parameters('location')]",
		"tags":       "[variables('commonTags')]",
		"kind":       "AzureCLI",
		"dependsOn":  []string{"vnetDeployment", "[resourceId('Microsoft.KeyVault/vaults', take(format('{0}', parameters('nuonInstallID')), 24))]"},
		"properties": map[string]any{
			"forceUpdateTag":    "[parameters('deployTimestamp')]",
			"azCliVersion":      "2.40.0",
			"timeout":           "PT30M",
			"retentionInterval": "PT1H",
			"environmentVariables": []map[string]any{
				{"name": "SUBSCRIPTION_ID", "value": "[subscription().subscriptionId]"},
				{"name": "SUBSCRIPTION_TENANT_ID", "value": "[subscription().tenantId]"},
				{"name": "RESOURCE_GROUP_ID", "value": "[resourceGroup().id]"},
				{"name": "RESOURCE_GROUP_NAME", "value": "[resourceGroup().name]"},
				{"name": "RESOURCE_GROUP_LOCATION", "value": "[resourceGroup().location]"},
				{"name": "VNET_ID", "value": "[reference('vnetDeployment').outputs.vnetId.value]"},
				{"name": "VNET_NAME", "value": "[reference('vnetDeployment').outputs.vnetName.value]"},
				{"name": "KEY_VAULT_ID", "value": "[resourceId('Microsoft.KeyVault/vaults', take(format('{0}', parameters('nuonInstallID')), 24))]"},
				{"name": "KEY_VAULT_NAME", "value": "[take(format('{0}', parameters('nuonInstallID')), 24)]"},
				{"name": "PUBLIC_SUBNET_IDS_CSV", "value": "[reference('vnetDeployment').outputs.publicSubnetIds.value]"},
				{"name": "PUBLIC_SUBNET_NAMES_CSV", "value": "[reference('vnetDeployment').outputs.publicSubnetNames.value]"},
				{"name": "PRIVATE_SUBNET_IDS_CSV", "value": "[reference('vnetDeployment').outputs.privateSubnetIds.value]"},
				{"name": "PRIVATE_SUBNET_NAMES_CSV", "value": "[reference('vnetDeployment').outputs.privateSubnetNames.value]"},
			},
			"scriptContent": scriptContent,
		},
	}
}
