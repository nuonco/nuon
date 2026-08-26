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

// snakeCase converts an ARM output name to the snake_case an install_stack
// output key uses, so ARM's virtualNetworkId reaches a template as
// vnet_virtual_network_id rather than vnet_virtualNetworkId.
func snakeCase(s string) string {
	s = camelBoundaryRegexp.ReplaceAllString(s, "${1}_${2}")
	return strings.ToLower(envNameRegexp.ReplaceAllString(s, "_"))
}

func (t *Templates) getPhoneHomeResources(inp *stacks.TemplateInput, customOutputs []customDeploymentOutputs, vnetExtraOutputs []string, scope armScope) []any {
	phoneHomeURL := inp.CloudFormationStackVersion.PhoneHomeURL

	operationIDs := azureOperationIdentities(inp.AppCfg)

	// The script reports the whole VNet contract, so most of the env vars below are
	// reads off the VNet deployment. vnetOutOptional covers the outputs a custom VNet
	// template is allowed to leave empty.
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

	// Build per-secret env vars and payload fields dynamically.
	var secretEnvVars []map[string]any
	var secretPayloadFields []string
	for _, secret := range inp.AppCfg.SecretsConfig.Secrets {
		envName := fmt.Sprintf("SECRET_%s_ID", secret.Name)
		// Construct the Key Vault secret URI from the vault name and secret name. The
		// secret itself is created by keyVaultDeployment at subscription scope, and by
		// the customer beforehand at resource-group scope; either way it is referenced
		// by this convention rather than read.
		kvSecretName := azureKeyVaultSecretName(secret.Name)
		envValue := fmt.Sprintf("[format('https://{0}.vault.azure.net/secrets/%s', %s)]", kvSecretName, scope.keyVaultNameInner())
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
	// The region the customer picked in the portal, or passed to `az stack sub
	// create`. Only meaningful at subscription scope, where deployment() carries a
	// location and where the record's location is immutable — see
	// AzureStackOutputs.DeploymentLocation. Gated so resource-group installs render
	// unchanged.
	if scope.subscription {
		payloadFields = append(payloadFields, `  "deployment_location": "$DEPLOYMENT_LOCATION"`)
	}
	// Local runners have no runnerDeployment to reference.
	if !t.cfg.UseLocalRunners {
		payloadFields = append(payloadFields, `  "runner_identity_principal_id": "$RUNNER_IDENTITY_PRINCIPAL_ID"`)
	}
	payloadFields = append(payloadFields, secretPayloadFields...)

	customerInputs := azureCustomerInputs(inp)
	if len(customerInputs) > 0 {
		// Unquoted, unlike every other field: the env var already holds a JSON object.
		payloadFields = append(payloadFields, fmt.Sprintf(`  "install_inputs": $%s`, installInputsEnvName))
	}

	// Outputs a custom VNet template declares beyond the fixed contract. Namespaced
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

	// Surface each identity's client ID as a stack output.
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
	// deployment().location exists only for subscription, management-group and
	// tenant deployments; at resource-group scope it is not available at all.
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
	// The runner's system-assigned identity. Secret sync and image sync run as
	// this identity rather than a per-operation one, so a sandbox has to be able
	// to grant it cluster access -- the Azure counterpart of the runner role ARNs
	// the AWS stack outputs.
	if !t.cfg.UseLocalRunners {
		envVars = append(envVars, map[string]any{
			"name":  "RUNNER_IDENTITY_PRINCIPAL_ID",
			"value": "[reference('runnerDeployment').outputs.vmssPrincipalId.value]",
		})
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

	// Depend on the identity role setup so a failed role deployment blocks the
	// outputs rather than reporting half-configured identities.
	dependsOn := []string{vnetDeployment}
	for _, co := range customOutputs {
		dependsOn = append(dependsOn, co.DeploymentName)
	}
	if _, uamiDependsOn := operationIdentityAttachment(operationIDs, scope); len(uamiDependsOn) > 0 {
		dependsOn = append(dependsOn, uamiDependsOn...)
	}
	dependsOn = append(dependsOn, operationIdentitySetupDependencies(operationIDs, scope)...)
	// The payload reports the vault's ID and a URI per secret, so it must not run
	// before they exist.
	dependsOn = append(dependsOn, scope.keyVaultDependsOn()...)

	// Microsoft.Resources/deploymentScripts is resource-group-scoped only, so at
	// subscription scope the script moves into the install resource group.
	//
	// The environment variables stay evaluated at the *outer* scope and cross the
	// boundary as a single array parameter. That is what makes this tractable: every
	// VNet-output reference and identity lookup in there resolves in the root,
	// where those deployments are declared, instead of needing ~25 individual
	// parameters. It also means the scope-sensitive expressions among them —
	// resourceGroup().id and friends — are correctly the subscription-scope forms.
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

	// deploymentScripts supports user-assigned identities only, so there is no
	// system-assigned alternative. The identity travels with the script into whichever
	// scope it lands in, so its resourceId always resolves where the script is declared.
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

	// A UAMI is not subscription-deployable, so it moves into the install resource
	// group alongside the script. Only the identity dependency can be expressed
	// inside; everything else is declared in the root and goes on the wrapper.
	if identityID != "" {
		script["dependsOn"] = []string{identityID}
		// The outer array cannot name the identity: it is created inside this wrapper,
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

	// Everything the script reports on has to exist first, and the script can no
	// longer say so itself — inner evaluation hides the root.
	for _, r := range resources {
		dependOn(r.(map[string]any), dependsOn)
	}
	return resources
}
