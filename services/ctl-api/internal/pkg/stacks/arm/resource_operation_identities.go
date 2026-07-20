package arm

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// azureIdentitySuffixRegexp keeps user-supplied role names to characters that are
// valid in an Azure user-assigned managed identity resource name.
var azureIdentitySuffixRegexp = regexp.MustCompile(`[^A-Za-z0-9-]`)

// azureBuiltInRoleGUIDs maps common Azure built-in role names to their role
// definition GUIDs. Role assignments reference the definition by GUID, and there
// is no ARM function to look one up by name. Unknown values are assumed to
// already be a role definition GUID.
var azureBuiltInRoleGUIDs = map[string]string{
	"Owner":                     "8e3af657-a8ff-443c-a75c-2fe8c4bcb635",
	"Contributor":               "b24988ac-6180-42a0-ab88-20f7382dd24c",
	"Reader":                    "acdd72a7-3385-48ef-bd42-f606fba81ae7",
	"User Access Administrator": "18d7d88d-d35e-4fb5-a5c3-7773c20a72d9",
	"Role Based Access Control Administrator":     "f58310d9-a9f6-439a-9e8d-f62e7b41a168",
	"Azure Kubernetes Service RBAC Cluster Admin": "b1ff04bb-8a4e-4dc4-8eb5-8693973ce19b",
	"Azure Kubernetes Service Cluster Admin Role": "0ab0b1a8-8aac-4efd-b8c2-3ee1fb270be8",
	"Azure Kubernetes Service Cluster User Role":  "4abbcc35-e782-43d8-92c5-2d3f1bd2253f",
	"Key Vault Administrator":                     "00482a5a-887f-4fb3-b363-3b7fe8e74483",
	"Key Vault Secrets User":                      "4633458b-17de-408a-b874-0445c86b69e6",
	"Key Vault Secrets Officer":                   "b86a8fe4-44ce-4948-aee5-eccb2c155cd7",
	"Storage Blob Data Contributor":               "ba92f5b4-2d11-453d-a403-e96b0029c9fe",
	"Storage Blob Data Owner":                     "b7e6dc6d-f1e8-4753-8033-0f276bb0955b",
	"Network Contributor":                         "4d97b98b-1d4f-4787-a291-c67834d212e7",
}

// azureOperationIdentity is a per-operation user-assigned managed identity derived
// from an app config Azure role. The runner assumes it (by client ID) to run an
// operation, so its own system identity holds no deploy permissions.
type azureOperationIdentity struct {
	// roleName is the (rendered) app-config role name, used as the output map key
	// for custom/break-glass identities and to build the identity resource name.
	roleName string
	// suffix is appended to the install ID to form the identity resource name.
	suffix string
	// kind is one of: provision, maintenance, deprovision, custom, breakglass.
	kind string
	// actions are custom RBAC actions, rolled into a subscription-level custom role.
	actions []string
	// builtInRoles are Azure built-in role names or GUIDs, assigned at RG scope.
	builtInRoles []string
}

func sanitizeAzureIdentitySuffix(name string) string {
	s := azureIdentitySuffixRegexp.ReplaceAllString(name, "-")
	s = strings.Trim(s, "-")
	if len(s) > 40 {
		s = s[:40]
	}
	return strings.ToLower(s)
}

func azureBuiltInRoleGUID(nameOrGUID string) string {
	if guid, ok := azureBuiltInRoleGUIDs[nameOrGUID]; ok {
		return guid
	}
	return nameOrGUID
}

// azureOperationIdentities extracts the per-operation identities to create from an
// app config's Azure-marked roles (standard, custom, and break-glass).
func azureOperationIdentities(appCfg *app.AppConfig) []azureOperationIdentity {
	var ids []azureOperationIdentity

	for _, role := range appCfg.PermissionsConfig.Roles {
		if role.CloudPlatform != string(app.CloudPlatformAzure) {
			continue
		}
		var suffix, kind string
		switch role.Type {
		case app.AWSIAMRoleTypeRunnerProvision:
			suffix, kind = "provision", "provision"
		case app.AWSIAMRoleTypeRunnerMaintenance:
			suffix, kind = "maintenance", "maintenance"
		case app.AWSIAMRoleTypeRunnerDeprovision:
			suffix, kind = "deprovision", "deprovision"
		case app.AWSIAMRoleTypeCustom:
			suffix, kind = "custom-"+sanitizeAzureIdentitySuffix(role.Name), "custom"
		default:
			continue
		}
		actions, builtIn := flattenAzurePolicies(role.Policies)
		ids = append(ids, azureOperationIdentity{
			roleName:     role.Name,
			suffix:       suffix,
			kind:         kind,
			actions:      actions,
			builtInRoles: builtIn,
		})
	}

	for _, role := range appCfg.BreakGlassConfig.Roles {
		if role.CloudPlatform != string(app.CloudPlatformAzure) {
			continue
		}
		actions, builtIn := flattenAzurePolicies(role.Policies)
		ids = append(ids, azureOperationIdentity{
			roleName:     role.Name,
			suffix:       "bg-" + sanitizeAzureIdentitySuffix(role.Name),
			kind:         "breakglass",
			actions:      actions,
			builtInRoles: builtIn,
		})
	}

	return ids
}

func flattenAzurePolicies(policies []app.AppAWSIAMPolicyConfig) (actions []string, builtInRoles []string) {
	for _, p := range policies {
		actions = append(actions, p.AzureActions...)
		builtInRoles = append(builtInRoles, p.AzureBuiltInRoles...)
	}
	return actions, builtInRoles
}

// uamiResourceIDExpr builds the ARM expression for an operation identity's
// resource ID. It is stable across the parent template and the runner nested
// deployment (both resolve to the same resource group).
func uamiResourceIDExpr(suffix string) string {
	return fmt.Sprintf("[resourceId('Microsoft.ManagedIdentity/userAssignedIdentities', format('{0}-%s', parameters('nuonInstallID')))]", suffix)
}

func uamiPrincipalIDExpr(suffix string) string {
	return fmt.Sprintf("[reference(resourceId('Microsoft.ManagedIdentity/userAssignedIdentities', format('{0}-%s', parameters('nuonInstallID'))), '2023-01-31').principalId]", suffix)
}

func uamiClientIDExpr(suffix string) string {
	return fmt.Sprintf("[reference(resourceId('Microsoft.ManagedIdentity/userAssignedIdentities', format('{0}-%s', parameters('nuonInstallID'))), '2023-01-31').clientId]", suffix)
}

func uamiNameExpr(suffix string) string {
	return "[" + uamiNameInner(suffix) + "]"
}

// uamiNameInner is the unbracketed form for embedding inside another ARM
// expression (e.g. guid(...)), avoiding invalid nested brackets.
func uamiNameInner(suffix string) string {
	return fmt.Sprintf("format('{0}-%s', parameters('nuonInstallID'))", suffix)
}

// getOperationIdentityResources builds every ARM resource needed for the
// per-operation managed identities: the identities themselves, a subscription
// level custom role per identity (register-action + declared actions), and
// resource-group-scoped assignments for any declared built-in roles.
func (t *Templates) getOperationIdentityResources(ids []azureOperationIdentity) []any {
	var resources []any

	for _, id := range ids {
		resources = append(resources, map[string]any{
			"type":       "Microsoft.ManagedIdentity/userAssignedIdentities",
			"apiVersion": "2023-01-31",
			"name":       uamiNameExpr(id.suffix),
			"location":   "[parameters('location')]",
			"tags":       "[variables('commonTags')]",
		})
	}

	for _, id := range ids {
		resources = append(resources, t.getOperationIdentityCustomRole(id))
		resources = append(resources, t.getOperationIdentityBuiltInRoleAssignments(id)...)
	}

	return resources
}

// azureRoleDeploymentToken returns a short, stable token for the subscription-level
// role deployment name. Custom/break-glass role names are user-defined and often
// long (and may repeat the install ID), which overflows ARM's 64-char
// deployment-name limit, so those hash to a short token.
func azureRoleDeploymentToken(id azureOperationIdentity) string {
	switch id.kind {
	case "provision", "maintenance", "deprovision":
		return id.suffix
	default:
		sum := sha256.Sum256([]byte(id.suffix))
		prefix := "custom"
		if id.kind == "breakglass" {
			prefix = "bg"
		}
		return prefix + "-" + hex.EncodeToString(sum[:])[:8]
	}
}

// getOperationIdentityCustomRole creates a subscription-level custom role for the
// identity and assigns it. The role always includes */register/action so the
// azurerm provider can register resource providers during apply — that action is
// subscription-scoped and was previously granted to the runner's system identity.
func (t *Templates) getOperationIdentityCustomRole(id azureOperationIdentity) map[string]any {
	roleActions := append([]string{"*/register/action"}, id.actions...)

	return map[string]any{
		"type":           "Microsoft.Resources/deployments",
		"apiVersion":     "2022-09-01",
		"name":           fmt.Sprintf("[format('{0}-%s-role', parameters('nuonInstallID'))]", azureRoleDeploymentToken(id)),
		"subscriptionId": "[subscription().subscriptionId]",
		"location":       "[resourceGroup().location]",
		"dependsOn":      []string{uamiResourceIDExpr(id.suffix)},
		"properties": map[string]any{
			"expressionEvaluationOptions": map[string]any{
				"scope": "inner",
			},
			"mode": "Incremental",
			"parameters": map[string]any{
				"roleName":    map[string]any{"value": fmt.Sprintf("[format('{0}-%s-role', parameters('nuonInstallID'))]", id.suffix)},
				"principalID": map[string]any{"value": uamiPrincipalIDExpr(id.suffix)},
			},
			"template": map[string]any{
				"$schema":        "https://schema.management.azure.com/schemas/2018-05-01/subscriptionDeploymentTemplate.json#",
				"contentVersion": "1.0.0.0",
				"parameters": map[string]any{
					"roleName":    map[string]any{"type": "string"},
					"principalID": map[string]any{"type": "string"},
				},
				"resources": []map[string]any{
					{
						"type":       "Microsoft.Authorization/roleDefinitions",
						"apiVersion": "2022-04-01",
						"name":       "[guid(subscription().id, parameters('roleName'))]",
						"properties": map[string]any{
							"roleName":    "[parameters('roleName')]",
							"description": "Nuon per-operation runner identity role",
							"assignableScopes": []string{
								"[subscription().id]",
							},
							"permissions": []map[string]any{
								{
									"actions":        roleActions,
									"notActions":     []string{},
									"dataActions":    []string{},
									"notDataActions": []string{},
								},
							},
						},
					},
					{
						"type":       "Microsoft.Authorization/roleAssignments",
						"apiVersion": "2022-04-01",
						// Name is derived from the (unique) role name rather than the
						// identity's principalId so ARM what-if can compute the
						// resource ID without a runtime reference() lookup.
						"name": "[guid(subscription().id, parameters('roleName'), 'roleassignment')]",
						"dependsOn": []string{
							"[subscriptionResourceId('Microsoft.Authorization/roleDefinitions', guid(subscription().id, parameters('roleName')))]",
						},
						"properties": map[string]any{
							"roleDefinitionId": "[subscriptionResourceId('Microsoft.Authorization/roleDefinitions', guid(subscription().id, parameters('roleName')))]",
							"principalId":      "[parameters('principalID')]",
							"principalType":    "ServicePrincipal",
						},
					},
				},
			},
		},
	}
}

// getOperationIdentityBuiltInRoleAssignments assigns each declared built-in role
// to the identity at resource-group scope.
func (t *Templates) getOperationIdentityBuiltInRoleAssignments(id azureOperationIdentity) []any {
	var assignments []any
	for _, role := range id.builtInRoles {
		guid := azureBuiltInRoleGUID(role)
		assignments = append(assignments, map[string]any{
			"type":       "Microsoft.Authorization/roleAssignments",
			"apiVersion": "2022-04-01",
			"name":       fmt.Sprintf("[guid(resourceGroup().id, %s, '%s')]", uamiNameInner(id.suffix), guid),
			"dependsOn":  []string{uamiResourceIDExpr(id.suffix)},
			"properties": map[string]any{
				"roleDefinitionId": fmt.Sprintf("[subscriptionResourceId('Microsoft.Authorization/roleDefinitions', '%s')]", guid),
				"principalId":      uamiPrincipalIDExpr(id.suffix),
				"principalType":    "ServicePrincipal",
			},
		})
	}
	return assignments
}

// azureIdentityEnvName derives a shell-safe env var name for an identity's client
// ID inside the phone-home script.
func azureIdentityEnvName(suffix string) string {
	s := strings.ToUpper(strings.ReplaceAll(suffix, "-", "_"))
	s = regexp.MustCompile(`[^A-Z0-9_]`).ReplaceAllString(s, "")
	return s + "_IDENTITY_CLIENT_ID"
}

// operationIdentityPhoneHomeFields returns the phone-home env vars and payload
// lines that surface each identity's client ID to the control plane as stack
// outputs. Custom and break-glass identities are emitted as native JSON objects
// keyed by role name so they decode into map[string]string.
func operationIdentityPhoneHomeFields(ids []azureOperationIdentity) (envVars []map[string]any, payloadFields []string) {
	if len(ids) == 0 {
		return nil, nil
	}

	custom := map[string]string{}
	breakGlass := map[string]string{}

	for _, id := range ids {
		envName := azureIdentityEnvName(id.suffix)
		envVars = append(envVars, map[string]any{"name": envName, "value": uamiClientIDExpr(id.suffix)})

		switch id.kind {
		case "provision":
			payloadFields = append(payloadFields, `  "provision_identity_client_id": "$`+envName+`"`)
		case "maintenance":
			payloadFields = append(payloadFields, `  "maintenance_identity_client_id": "$`+envName+`"`)
		case "deprovision":
			payloadFields = append(payloadFields, `  "deprovision_identity_client_id": "$`+envName+`"`)
		case "custom":
			custom[id.roleName] = envName
		case "breakglass":
			breakGlass[id.roleName] = envName
		}
	}

	if len(custom) > 0 {
		payloadFields = append(payloadFields, `  "custom_identity_client_ids": `+jsonEnvObject(custom))
	}
	if len(breakGlass) > 0 {
		payloadFields = append(payloadFields, `  "break_glass_identity_client_ids": `+jsonEnvObject(breakGlass))
	}

	return envVars, payloadFields
}

// jsonEnvObject builds a native JSON object literal mapping role names to shell
// variable references, e.g. {"my-role":"$MY_ROLE_IDENTITY_CLIENT_ID"}.
func jsonEnvObject(nameToEnv map[string]string) string {
	names := make([]string, 0, len(nameToEnv))
	for name := range nameToEnv {
		names = append(names, name)
	}
	sort.Strings(names)

	parts := make([]string, 0, len(names))
	for _, name := range names {
		parts = append(parts, fmt.Sprintf("%q:\"$%s\"", name, nameToEnv[name]))
	}
	return "{" + strings.Join(parts, ",") + "}"
}

// operationIdentitySetupDependencies returns the resource identifiers (per-operation
// role deployments and built-in role assignments) that must complete before the
// phone-home reports identity client IDs. Gating on these means a failed role
// setup blocks the outputs instead of reporting half-configured identities.
func operationIdentitySetupDependencies(ids []azureOperationIdentity) []string {
	var deps []string
	for _, id := range ids {
		deps = append(deps, fmt.Sprintf("[format('{0}-%s-role', parameters('nuonInstallID'))]", azureRoleDeploymentToken(id)))
		for _, role := range id.builtInRoles {
			guid := azureBuiltInRoleGUID(role)
			deps = append(deps, fmt.Sprintf("[resourceId('Microsoft.Authorization/roleAssignments', guid(resourceGroup().id, %s, '%s'))]", uamiNameInner(id.suffix), guid))
		}
	}
	return deps
}

// operationIdentityAttachment returns the VMSS userAssignedIdentities object and
// the identity resource IDs the runner deployment must depend on.
func operationIdentityAttachment(ids []azureOperationIdentity) (userAssigned map[string]any, dependsOn []string) {
	if len(ids) == 0 {
		return nil, nil
	}
	userAssigned = map[string]any{}
	for _, id := range ids {
		userAssigned[uamiResourceIDExpr(id.suffix)] = map[string]any{}
		dependsOn = append(dependsOn, uamiResourceIDExpr(id.suffix))
	}
	return userAssigned, dependsOn
}
