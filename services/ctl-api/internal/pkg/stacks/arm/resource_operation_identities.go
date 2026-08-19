package arm

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/azureroles"
)

var azureIdentitySuffixRegexp = regexp.MustCompile(`[^A-Za-z0-9-]`)

type azureOperationIdentity struct {
	// roleName is the rendered app-config role name; also the output map key for
	// custom/break-glass identities.
	roleName     string
	suffix       string
	kind         string
	actions      []string
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
	return azureroles.GUID(nameOrGUID)
}

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

// identitiesDeploymentName is the nested deployment the UAMIs and their built-in
// role assignments move into at subscription scope. Anything in the root that
// needs one of them depends on this rather than on the identity directly: a root
// resource cannot depend on a resource declared inside a nested deployment.
const identitiesDeploymentName = "identitiesDeployment"

const uamiResourceType = "Microsoft.ManagedIdentity/userAssignedIdentities"

// The uami*Expr helpers take the scope of the expression's *reader*, not of the
// identity. Read from the root at subscription scope they need the fully
// qualified four-argument form; read from inside identitiesDeployment they are
// back at resource-group scope and take the short form, which is also exactly
// what resource-group-scope installs have always emitted.
func uamiResourceIDExpr(suffix string, scope armScope) string {
	return scope.rgResourceIDExpr(uamiResourceType, uamiNameInner(suffix))
}

func uamiPrincipalIDExpr(suffix string, scope armScope) string {
	return fmt.Sprintf("[reference(%s, '2023-01-31').principalId]", scope.rgResourceIDInner(uamiResourceType, uamiNameInner(suffix)))
}

func uamiClientIDExpr(suffix string, scope armScope) string {
	return fmt.Sprintf("[reference(%s, '2023-01-31').clientId]", scope.rgResourceIDInner(uamiResourceType, uamiNameInner(suffix)))
}

func uamiNameExpr(suffix string) string {
	return "[" + uamiNameInner(suffix) + "]"
}

// uamiNameInner is unbracketed for embedding in another ARM expression; ARM does
// not allow nested [ ].
func uamiNameInner(suffix string) string {
	return fmt.Sprintf("format('{0}-%s', parameters('nuonInstallID'))", suffix)
}

func (t *Templates) uamiResource(id azureOperationIdentity, scope armScope) map[string]any {
	return map[string]any{
		"type":       uamiResourceType,
		"apiVersion": "2023-01-31",
		"name":       uamiNameExpr(id.suffix),
		"location":   "[parameters('location')]",
		"tags":       scope.innerCommonTagsExpr(),
	}
}

func (t *Templates) getOperationIdentityResources(ids []azureOperationIdentity, scope armScope) []any {
	// Built-in role assignments are always emitted at resource-group scope: at
	// resource-group scope that is the root, and at subscription scope it is inside
	// the wrapper, where resourceGroup() resolves to the install group again. Either
	// way their guid(resourceGroup().id, …) names are identical, which is what keeps
	// redeploys off RoleAssignmentExists.
	inner := armScope{}

	if !scope.subscription {
		// Emission order is load-bearing for the byte-identical guarantee, so keep
		// the historical sequence exactly: every identity, then each identity's
		// subscription role deployment followed by its built-in grants.
		var resources []any
		for _, id := range ids {
			resources = append(resources, t.uamiResource(id, scope))
		}
		for _, id := range ids {
			resources = append(resources, t.getOperationIdentityCustomRole(id, scope))
			resources = append(resources, t.getOperationIdentityBuiltInRoleAssignments(id, inner)...)
		}
		return resources
	}

	// UAMIs and RG-scoped role assignments are not subscription-deployable, so they
	// move into the install resource group together.
	var grouped []any
	for _, id := range ids {
		grouped = append(grouped, t.uamiResource(id, scope))
	}
	for _, id := range ids {
		grouped = append(grouped, t.getOperationIdentityBuiltInRoleAssignments(id, inner)...)
	}

	resources := scope.wrapInInstallRG(identitiesDeploymentName, map[string]nestedParam{
		"nuonInstallID": {typ: "string", value: "[parameters('nuonInstallID')]"},
		"location":      {typ: "string", value: "[parameters('location')]"},
		"commonTags":    {typ: "object", value: "[variables('commonTags')]"},
	}, grouped)

	// The custom role deployments target the subscription, and ARM does not allow a
	// subscription-scoped nested deployment inside another nested deployment, so
	// they stay in the root and read the identities across the wrapper boundary.
	for _, id := range ids {
		resources = append(resources, t.getOperationIdentityCustomRole(id, scope))
	}

	return resources
}

// azureRoleDeploymentToken keeps the subscription-level role deployment name under
// ARM's 64-char limit; custom/break-glass role names are user-defined and can
// overflow, so they hash to a short token.
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

func roleDeploymentNameExpr(id azureOperationIdentity) string {
	return fmt.Sprintf("[format('{0}-%s-role', parameters('nuonInstallID'))]", azureRoleDeploymentToken(id))
}

// getOperationIdentityCustomRole always includes */register/action so the azurerm
// provider can register resource providers on apply (a subscription-scoped action
// previously held by the runner's system identity).
func (t *Templates) getOperationIdentityCustomRole(id azureOperationIdentity, scope armScope) map[string]any {
	roleActions := append([]string{"*/register/action"}, id.actions...)

	return map[string]any{
		"type":           "Microsoft.Resources/deployments",
		"apiVersion":     "2022-09-01",
		"name":           roleDeploymentNameExpr(id),
		"subscriptionId": "[subscription().subscriptionId]",
		"location":       scope.locationExpr(),
		"dependsOn":      []string{identityDependency(id, scope)},
		"properties": map[string]any{
			"expressionEvaluationOptions": map[string]any{
				"scope": "inner",
			},
			"mode": "Incremental",
			"parameters": map[string]any{
				"roleName":    map[string]any{"value": fmt.Sprintf("[format('{0}-%s-role', parameters('nuonInstallID'))]", id.suffix)},
				"principalID": map[string]any{"value": uamiPrincipalIDExpr(id.suffix, scope)},
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
						// Do NOT change: renaming an existing assignment fails redeploys
						// with RoleAssignmentExists (Azure dedupes by principal+role+scope).
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

// getOperationIdentityBuiltInRoleAssignments is always called with resource-group
// scope — see getOperationIdentityResources. The scope parameter is explicit
// rather than hardcoded so the assumption is visible at the call site.
func (t *Templates) getOperationIdentityBuiltInRoleAssignments(id azureOperationIdentity, scope armScope) []any {
	var assignments []any
	for _, role := range id.builtInRoles {
		guid := azureBuiltInRoleGUID(role)
		assignments = append(assignments, map[string]any{
			"type":       "Microsoft.Authorization/roleAssignments",
			"apiVersion": "2022-04-01",
			"name":       fmt.Sprintf("[guid(resourceGroup().id, %s, '%s')]", uamiNameInner(id.suffix), guid),
			"dependsOn":  []string{uamiResourceIDExpr(id.suffix, scope)},
			"properties": map[string]any{
				"roleDefinitionId": fmt.Sprintf("[subscriptionResourceId('Microsoft.Authorization/roleDefinitions', '%s')]", guid),
				"principalId":      uamiPrincipalIDExpr(id.suffix, scope),
				"principalType":    "ServicePrincipal",
			},
		})
	}
	return assignments
}

// identityDependency is what a root resource waits on to know an identity exists.
// At subscription scope the identities live inside identitiesDeployment, and a
// root resource cannot depend on a resource declared inside a nested deployment —
// it depends on the deployment itself.
func identityDependency(id azureOperationIdentity, scope armScope) string {
	if scope.subscription {
		return identitiesDeploymentName
	}
	return uamiResourceIDExpr(id.suffix, scope)
}

func azureIdentityEnvName(suffix string) string {
	s := strings.ToUpper(strings.ReplaceAll(suffix, "-", "_"))
	s = regexp.MustCompile(`[^A-Z0-9_]`).ReplaceAllString(s, "")
	return s + "_IDENTITY_CLIENT_ID"
}

// operationIdentityPhoneHomeFields returns phone-home env vars and payload lines
// carrying each identity's client ID. Custom/break-glass are native JSON objects
// keyed by role name so they decode into map[string]string.
func operationIdentityPhoneHomeFields(ids []azureOperationIdentity, scope armScope) (envVars []map[string]any, payloadFields []string) {
	if len(ids) == 0 {
		return nil, nil
	}

	custom := map[string]string{}
	breakGlass := map[string]string{}

	for _, id := range ids {
		envName := azureIdentityEnvName(id.suffix)
		envVars = append(envVars, map[string]any{"name": envName, "value": uamiClientIDExpr(id.suffix, scope)})

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

// operationIdentitySetupDependencies lists the role deployments and built-in
// assignments the phone-home must wait on, so a failed role setup blocks the
// outputs instead of reporting half-configured identities.
func operationIdentitySetupDependencies(ids []azureOperationIdentity, scope armScope) []string {
	var deps []string
	for _, id := range ids {
		deps = append(deps, roleDeploymentNameExpr(id))
		if scope.subscription {
			// The built-in assignments are inside identitiesDeployment, which the
			// caller already depends on via operationIdentityAttachment.
			continue
		}
		for _, role := range id.builtInRoles {
			guid := azureBuiltInRoleGUID(role)
			deps = append(deps, fmt.Sprintf("[resourceId('Microsoft.Authorization/roleAssignments', guid(resourceGroup().id, %s, '%s'))]", uamiNameInner(id.suffix), guid))
		}
	}
	return deps
}

// operationIdentityAttachment returns the VMSS userAssignedIdentities map and the
// dependencies a consumer needs. scope is the consumer's scope: the runner's inner
// template reads at resource-group scope, while the root reads at its own.
func operationIdentityAttachment(ids []azureOperationIdentity, scope armScope) (userAssigned map[string]any, dependsOn []string) {
	if len(ids) == 0 {
		return nil, nil
	}
	userAssigned = map[string]any{}
	seen := map[string]bool{}
	for _, id := range ids {
		userAssigned[uamiResourceIDExpr(id.suffix, scope)] = map[string]any{}

		dep := identityDependency(id, scope)
		if seen[dep] {
			continue
		}
		seen[dep] = true
		dependsOn = append(dependsOn, dep)
	}
	return userAssigned, dependsOn
}
