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

var camelBoundaryRegexp = regexp.MustCompile(`([a-z0-9])([A-Z])`)

func snakeCase(s string) string {
	s = camelBoundaryRegexp.ReplaceAllString(s, "${1}_${2}")
	return strings.ToLower(envNameRegexp.ReplaceAllString(s, "_"))
}

func (t *Templates) getPhoneHomeResources(inp *stacks.TemplateInput, customOutputs []customDeploymentOutputs, vnetExtraOutputs []string, scope armScope, telemetryEndpoint string) []any {
	phoneHomeURL := inp.CloudFormationStackVersion.PhoneHomeURL

	operationIDs := azureOperationIdentities(inp.AppCfg)

	vnetDeployment := scope.vnetDeploymentName(inp.Install.ID)
	vnetOut := func(output string) string {
		return fmt.Sprintf("[reference('%s').outputs.%s.value]", vnetDeployment, output)
	}
	vnetOutOptional := func(output string) string {
		return fmt.Sprintf(
			"[if(not(empty(reference('%[1]s').outputs.%[2]s.value)), reference('%[1]s').outputs.%[2]s.value, '')]",
			vnetDeployment, output,
		)
	}

	var secretEnvVars []map[string]any
	var secretPayloadFields []string
	for _, secret := range inp.AppCfg.SecretsConfig.Secrets {
		envName := fmt.Sprintf("SECRET_%s_ID", secret.Name)
		kvSecretName := azureKeyVaultSecretName(secret.Name)
		envValue := fmt.Sprintf("[format('https://{0}.vault.azure.net/secrets/%s', %s)]", kvSecretName, scope.keyVaultNameInner())
		secretEnvVars = append(secretEnvVars, map[string]any{"name": envName, "value": envValue})
		secretPayloadFields = append(secretPayloadFields, fmt.Sprintf(`  "%s_secret_id": "$%s"`, secret.Name, envName))
	}

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
	if scope.subscription {
		payloadFields = append(payloadFields, `  "deployment_location": "$DEPLOYMENT_LOCATION"`)
	}
	if !t.cfg.UseLocalRunners {
		payloadFields = append(payloadFields, `  "runner_identity_principal_id": "$RUNNER_IDENTITY_PRINCIPAL_ID"`)
	}
	if telemetryEndpoint != "" {
		payloadFields = append(payloadFields, `  "telemetry_endpoint": "$TELEMETRY_ENDPOINT"`)
	}
	payloadFields = append(payloadFields, secretPayloadFields...)

	customerInputs := azureCustomerInputs(inp)
	if len(customerInputs) > 0 {
		payloadFields = append(payloadFields, fmt.Sprintf(`  "install_inputs": $%s`, installInputsEnvName))
	}

	// why: Outputs a custom VNet template declares beyond the fixed contract. Namespaced
	// under vnet_ because the raw names collide: a VNet stack that makes its own
	// resource group emits resourceGroupName, which is already Nuon's install group.
	var vnetExtraEnvVars []map[string]any
	for _, key := range vnetExtraOutputs {
		envName := "VNET_OUT_" + envToken(key)
		vnetExtraEnvVars = append(vnetExtraEnvVars, map[string]any{
			"name":  envName,
			"value": fmt.Sprintf("[string(reference('%s').outputs.%s.value)]", vnetDeployment, key),
		})
		payloadFields = append(payloadFields, fmt.Sprintf(`  "vnet_%s": "$%s"`, snakeCase(key), envName))
	}

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

	identityEnvVars, identityPayloadFields := operationIdentityPhoneHomeFields(operationIDs, scope)
	payloadFields = append(payloadFields, identityPayloadFields...)

	payloadJSON := "{\n" + strings.Join(payloadFields, ",\n") + "\n}"

	authPreamble := ""
	authFlag := ""
	if inp.PhoneHomeIdentityName != "" {
		authPreamble = phoneHomeAuthScript
		authFlag = "  -K \"$CURL_CONFIG\" \\\n"
	}

	scriptContent := `#!/bin/bash
` + authPreamble + `
PAYLOAD=$(cat << EOF
` + payloadJSON + `
EOF
)

curl -X POST \
  "` + phoneHomeURL + `" \
` + authFlag + `  -H "Content-Type: application/json" \
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
	}
	if scope.subscription {
		envVars = append(envVars, map[string]any{"name": "DEPLOYMENT_LOCATION", "value": "[deployment().location]"})
	}
	envVars = append(envVars, []map[string]any{
		{"name": "RESOURCE_GROUP_ID", "value": scope.rgIDExpr()},
		{"name": "RESOURCE_GROUP_NAME", "value": scope.rgNameExpr()},
		{"name": "RESOURCE_GROUP_LOCATION", "value": scope.locationExpr()},
		{"name": "VNET_ID", "value": vnetOut("vnetId")},
		{"name": "VNET_NAME", "value": vnetOut("vnetName")},
		{"name": "KEY_VAULT_ID", "value": scope.rgResourceIDExpr("Microsoft.KeyVault/vaults", scope.keyVaultNameInner())},
		{"name": "KEY_VAULT_NAME", "value": "[" + scope.keyVaultNameInner() + "]"},
		{"name": "PUBLIC_SUBNET_1_ID", "value": vnetOut("publicSubnet1Id")},
		{"name": "PUBLIC_SUBNET_1_NAME", "value": vnetOut("publicSubnet1Name")},
		{"name": "PUBLIC_SUBNET_2_ID", "value": vnetOutOptional("publicSubnet2Id")},
		{"name": "PUBLIC_SUBNET_2_NAME", "value": vnetOutOptional("publicSubnet2Name")},
		{"name": "PUBLIC_SUBNET_3_ID", "value": vnetOutOptional("publicSubnet3Id")},
		{"name": "PUBLIC_SUBNET_3_NAME", "value": vnetOutOptional("publicSubnet3Name")},
		{"name": "PRIVATE_SUBNET_1_ID", "value": vnetOut("privateSubnet1Id")},
		{"name": "PRIVATE_SUBNET_1_NAME", "value": vnetOut("privateSubnet1Name")},
		{"name": "PRIVATE_SUBNET_2_ID", "value": vnetOutOptional("privateSubnet2Id")},
		{"name": "PRIVATE_SUBNET_2_NAME", "value": vnetOutOptional("privateSubnet2Name")},
		{"name": "PRIVATE_SUBNET_3_ID", "value": vnetOutOptional("privateSubnet3Id")},
		{"name": "PRIVATE_SUBNET_3_NAME", "value": vnetOutOptional("privateSubnet3Name")},
		{"name": "PUBLIC_SUBNET_IDS_CSV", "value": vnetOut("publicSubnetIds")},
		{"name": "PUBLIC_SUBNET_NAMES_CSV", "value": vnetOut("publicSubnetNames")},
		{"name": "PRIVATE_SUBNET_IDS_CSV", "value": vnetOut("privateSubnetIds")},
		{"name": "PRIVATE_SUBNET_NAMES_CSV", "value": vnetOut("privateSubnetNames")},
	}...)
	if !t.cfg.UseLocalRunners {
		envVars = append(envVars, map[string]any{
			"name":  "RUNNER_IDENTITY_PRINCIPAL_ID",
			"value": "[reference('runnerDeployment').outputs.vmssPrincipalId.value]",
		})
	}
	if telemetryEndpoint != "" {
		envVars = append(envVars, map[string]any{"name": "TELEMETRY_ENDPOINT", "value": telemetryEndpoint})
	}
	envVars = append(envVars, secretEnvVars...)
	if len(customerInputs) > 0 {
		envVars = append(envVars, map[string]any{
			"name":  installInputsEnvName,
			"value": installInputsObjectExpr(customerInputs),
		})
	}
	envVars = append(envVars, vnetExtraEnvVars...)
	envVars = append(envVars, customEnvVars...)
	envVars = append(envVars, identityEnvVars...)

	dependsOn := []string{vnetDeployment}
	if telemetryEndpoint != "" {
		dependsOn = append(dependsOn, "runnerDeployment")
	}
	for _, co := range customOutputs {
		dependsOn = append(dependsOn, co.DeploymentName)
	}
	if _, uamiDependsOn := operationIdentityAttachment(operationIDs, scope); len(uamiDependsOn) > 0 {
		dependsOn = append(dependsOn, uamiDependsOn...)
	}
	dependsOn = append(dependsOn, operationIdentitySetupDependencies(operationIDs, scope)...)
	dependsOn = append(dependsOn, scope.keyVaultDependsOn()...)

	environmentVariables := any(envVars)
	if scope.subscription {
		environmentVariables = "[parameters('environmentVariables')]"
	}

	script := map[string]any{
		"type":       "Microsoft.Resources/deploymentScripts",
		"apiVersion": "2023-08-01",
		"name":       "[format('{0}-phone-home-script', parameters('nuonInstallID'))]",
		"location":   "[parameters('location')]",
		"tags":       scope.innerCommonTagsExpr(),
		"kind":       "AzureCLI",
		"properties": map[string]any{
			"forceUpdateTag":       "[parameters('deployTimestamp')]",
			"azCliVersion":         "2.40.0",
			"timeout":              "PT30M",
			"retentionInterval":    "PT1H",
			"environmentVariables": environmentVariables,
			"scriptContent":        scriptContent,
		},
	}

	var identity []any
	identityID := ""
	if inp.PhoneHomeIdentityName != "" {
		identityID = phoneHomeIdentityResourceID(inp.PhoneHomeIdentityName)
		script["identity"] = map[string]any{
			"type": "UserAssigned",
			"userAssignedIdentities": map[string]any{
				identityID: map[string]any{},
			},
		}
		identity = append(identity, getPhoneHomeIdentityResource(inp.PhoneHomeIdentityName, scope))
	}

	if !scope.subscription {
		if identityID != "" {
			dependsOn = append(dependsOn, identityID)
			script["properties"].(map[string]any)["environmentVariables"] =
				append(envVars, phoneHomeIdentityClientIDEnvVar(inp.PhoneHomeIdentityName))
		}
		script["dependsOn"] = dependsOn

		return append(identity, script)
	}

	if identityID != "" {
		script["dependsOn"] = []string{identityID}
		// why: The outer array cannot name the identity: it is created inside this wrapper,
		// so resourceId() in the root resolves against no resource group at all.
		script["properties"].(map[string]any)["environmentVariables"] =
			phoneHomeInnerEnvVarsExpr(inp.PhoneHomeIdentityName)
	}

	resources := scope.wrapInInstallRG(phoneHomeDeploymentName, map[string]nestedParam{
		"nuonInstallID":        {typ: "string", value: scope.nuonIDRef("nuonInstallID")},
		"location":             {typ: "string", value: scope.rootLocationRef()},
		"commonTags":           {typ: "object", value: "[variables('commonTags')]"},
		"deployTimestamp":      {typ: "string", value: "[parameters('deployTimestamp')]"},
		"environmentVariables": {typ: "array", value: envVars},
	}, append(identity, script), nil)

	for _, r := range resources {
		dependOn(r.(map[string]any), dependsOn)
	}
	return resources
}
