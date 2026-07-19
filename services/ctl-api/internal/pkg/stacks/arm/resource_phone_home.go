package arm

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks"
)

var envNameRegexp = regexp.MustCompile(`[^A-Za-z0-9]`)

func envToken(s string) string {
	return strings.ToUpper(envNameRegexp.ReplaceAllString(s, "_"))
}

func (t *Templates) getPhoneHomeResource(inp *stacks.TemplateInput, customOutputs []customDeploymentOutputs) map[string]any {
	phoneHomeURL := inp.CloudFormationStackVersion.PhoneHomeURL

	operationIDs := azureOperationIdentities(inp.AppCfg)

	// Build per-secret env vars and payload fields dynamically.
	var secretEnvVars []map[string]any
	var secretPayloadFields []string
	for _, secret := range inp.AppCfg.SecretsConfig.Secrets {
		envName := fmt.Sprintf("SECRET_%s_ID", secret.Name)
		// Construct the Key Vault secret URI from the vault name and secret name.
		// Secrets are customer pre-created; we reference them by convention.
		// Azure Key Vault secret names only allow alphanumeric characters and hyphens.
		kvSecretName := strings.ReplaceAll(secret.Name, "_", "-")
		envValue := fmt.Sprintf("[format('https://{0}.vault.azure.net/secrets/%s', take(format('{0}', parameters('nuonInstallID')), 24))]", kvSecretName)
		secretEnvVars = append(secretEnvVars, map[string]any{"name": envName, "value": envValue})
		secretPayloadFields = append(secretPayloadFields, fmt.Sprintf(`  "%s_secret_id": "$%s"`, secret.Name, envName))
	}

	// Build the payload JSON with optional secret fields.
	payloadFields := []string{
		`  "request_type": "Create"`,
		`  "phone_home_type": "azure"`,
		`  "resource_group_id": "$RESOURCE_GROUP_ID"`,
		`  "resource_group_name": "$RESOURCE_GROUP_NAME"`,
		`  "resource_group_location": "$RESOURCE_GROUP_LOCATION"`,
		`  "network_id": "$VNET_ID"`,
		`  "network_name": "$VNET_NAME"`,
		`  "key_vault_id": "$KEY_VAULT_ID"`,
		`  "key_vault_name": "$KEY_VAULT_NAME"`,
		`  "public_subnet_ids": "$PUBLIC_SUBNET_IDS_CSV"`,
		`  "public_subnet_names": "$PUBLIC_SUBNET_NAMES_CSV"`,
		`  "private_subnet_ids": "$PRIVATE_SUBNET_IDS_CSV"`,
		`  "private_subnet_names": "$PRIVATE_SUBNET_NAMES_CSV"`,
		`  "subscription_id": "$SUBSCRIPTION_ID"`,
		`  "subscription_tenant_id": "$SUBSCRIPTION_TENANT_ID"`,
	}
	payloadFields = append(payloadFields, secretPayloadFields...)

	// Custom nested stack outputs, mirroring the AWS phone-home shape:
	// custom_nested_stacks.<name>.outputs.<key>. Non-string ARM outputs are
	// serialized with string().
	var customEnvVars []map[string]any
	if len(customOutputs) > 0 {
		var stackFields []string
		for _, co := range customOutputs {
			var outFields []string
			for _, key := range co.OutputKeys {
				envName := fmt.Sprintf("CUSTOM_%s_%s", envToken(co.StackName), envToken(key))
				customEnvVars = append(customEnvVars, map[string]any{
					"name":  envName,
					"value": fmt.Sprintf("[string(reference('%s').outputs.%s.value)]", co.DeploymentName, key),
				})
				outFields = append(outFields, fmt.Sprintf(`      "%s": "$%s"`, key, envName))
			}
			stackFields = append(stackFields, fmt.Sprintf("    \"%s\": {\n      \"outputs\": {\n  %s\n      }\n    }", co.StackName, strings.Join(outFields, ",\n  ")))
		}
		payloadFields = append(payloadFields, "  \"custom_nested_stacks\": {\n"+strings.Join(stackFields, ",\n")+"\n  }")
	}

	// Per-operation managed identity client IDs, surfaced as stack outputs so the
	// control plane can resolve which identity the runner assumes for each op.
	identityEnvVars, identityPayloadFields := operationIdentityPhoneHomeFields(operationIDs)
	payloadFields = append(payloadFields, identityPayloadFields...)

	payloadJSON := "{\n" + strings.Join(payloadFields, ",\n") + "\n}"

	scriptContent := `#!/bin/bash

PAYLOAD=$(cat << EOF
` + payloadJSON + `
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

	envVars := []map[string]any{
		{"name": "SUBSCRIPTION_ID", "value": "[subscription().subscriptionId]"},
		{"name": "SUBSCRIPTION_TENANT_ID", "value": "[subscription().tenantId]"},
		{"name": "RESOURCE_GROUP_ID", "value": "[resourceGroup().id]"},
		{"name": "RESOURCE_GROUP_NAME", "value": "[resourceGroup().name]"},
		{"name": "RESOURCE_GROUP_LOCATION", "value": "[resourceGroup().location]"},
		{"name": "VNET_ID", "value": "[reference('vnetDeployment').outputs.vnetId.value]"},
		{"name": "VNET_NAME", "value": "[reference('vnetDeployment').outputs.vnetName.value]"},
		{"name": "KEY_VAULT_ID", "value": "[resourceId('Microsoft.KeyVault/vaults', take(format('{0}', parameters('nuonInstallID')), 24))]"},
		{"name": "KEY_VAULT_NAME", "value": "[take(format('{0}', parameters('nuonInstallID')), 24)]"},
		{"name": "PUBLIC_SUBNET_1_ID", "value": "[reference('vnetDeployment').outputs.publicSubnet1Id.value]"},
		{"name": "PUBLIC_SUBNET_1_NAME", "value": "[reference('vnetDeployment').outputs.publicSubnet1Name.value]"},
		{"name": "PUBLIC_SUBNET_2_ID", "value": "[if(not(empty(reference('vnetDeployment').outputs.publicSubnet2Id.value)), reference('vnetDeployment').outputs.publicSubnet2Id.value, '')]"},
		{"name": "PUBLIC_SUBNET_2_NAME", "value": "[if(not(empty(reference('vnetDeployment').outputs.publicSubnet2Name.value)), reference('vnetDeployment').outputs.publicSubnet2Name.value, '')]"},
		{"name": "PUBLIC_SUBNET_3_ID", "value": "[if(not(empty(reference('vnetDeployment').outputs.publicSubnet3Id.value)), reference('vnetDeployment').outputs.publicSubnet3Id.value, '')]"},
		{"name": "PUBLIC_SUBNET_3_NAME", "value": "[if(not(empty(reference('vnetDeployment').outputs.publicSubnet3Name.value)), reference('vnetDeployment').outputs.publicSubnet3Name.value, '')]"},
		{"name": "PRIVATE_SUBNET_1_ID", "value": "[reference('vnetDeployment').outputs.privateSubnet1Id.value]"},
		{"name": "PRIVATE_SUBNET_1_NAME", "value": "[reference('vnetDeployment').outputs.privateSubnet1Name.value]"},
		{"name": "PRIVATE_SUBNET_2_ID", "value": "[if(not(empty(reference('vnetDeployment').outputs.privateSubnet2Id.value)), reference('vnetDeployment').outputs.privateSubnet2Id.value, '')]"},
		{"name": "PRIVATE_SUBNET_2_NAME", "value": "[if(not(empty(reference('vnetDeployment').outputs.privateSubnet2Name.value)), reference('vnetDeployment').outputs.privateSubnet2Name.value, '')]"},
		{"name": "PRIVATE_SUBNET_3_ID", "value": "[if(not(empty(reference('vnetDeployment').outputs.privateSubnet3Id.value)), reference('vnetDeployment').outputs.privateSubnet3Id.value, '')]"},
		{"name": "PRIVATE_SUBNET_3_NAME", "value": "[if(not(empty(reference('vnetDeployment').outputs.privateSubnet3Name.value)), reference('vnetDeployment').outputs.privateSubnet3Name.value, '')]"},
		{"name": "PUBLIC_SUBNET_IDS_CSV", "value": "[reference('vnetDeployment').outputs.publicSubnetIds.value]"},
		{"name": "PUBLIC_SUBNET_NAMES_CSV", "value": "[reference('vnetDeployment').outputs.publicSubnetNames.value]"},
		{"name": "PRIVATE_SUBNET_IDS_CSV", "value": "[reference('vnetDeployment').outputs.privateSubnetIds.value]"},
		{"name": "PRIVATE_SUBNET_NAMES_CSV", "value": "[reference('vnetDeployment').outputs.privateSubnetNames.value]"},
	}
	envVars = append(envVars, secretEnvVars...)
	envVars = append(envVars, customEnvVars...)
	envVars = append(envVars, identityEnvVars...)

	// The phone-home script reads custom-stack outputs and each identity's
	// clientId, so those resources must exist first.
	dependsOn := []string{"vnetDeployment"}
	for _, co := range customOutputs {
		dependsOn = append(dependsOn, co.DeploymentName)
	}
	if _, uamiDependsOn := operationIdentityAttachment(operationIDs); len(uamiDependsOn) > 0 {
		dependsOn = append(dependsOn, uamiDependsOn...)
	}

	return map[string]any{
		"type":       "Microsoft.Resources/deploymentScripts",
		"apiVersion": "2023-08-01",
		"name":       "[format('{0}-phone-home-script', parameters('nuonInstallID'))]",
		"location":   "[parameters('location')]",
		"tags":       "[variables('commonTags')]",
		"kind":       "AzureCLI",
		"dependsOn":  dependsOn,
		"properties": map[string]any{
			"forceUpdateTag":       "[parameters('deployTimestamp')]",
			"azCliVersion":         "2.40.0",
			"timeout":              "PT30M",
			"retentionInterval":    "PT1H",
			"environmentVariables": envVars,
			"scriptContent":        scriptContent,
		},
	}
}
