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
	return scope.rgResourceIDExpr(uamiResourceType, uamiNameInner(suffix, scope))
}

func uamiPrincipalIDExpr(suffix string, scope armScope) string {
	return fmt.Sprintf("[reference(%s, '2023-01-31').principalId]", scope.rgResourceIDInner(uamiResourceType, uamiNameInner(suffix, scope)))
}

func uamiClientIDExpr(suffix string, scope armScope) string {
	return fmt.Sprintf("[reference(%s, '2023-01-31').clientId]", scope.rgResourceIDInner(uamiResourceType, uamiNameInner(suffix, scope)))
}

// azureIdentityOutputKey names identitiesDeployment's output carrying one field of
// one identity. Read back with dot notation, so the token's dashes are stripped.
func azureIdentityOutputKey(id azureOperationIdentity, field string) string {
	return strings.ReplaceAll(azureRoleDeploymentToken(id), "-", "") + field
}

// identityPrincipalIDExpr and identityClientIDExpr are how a root resource reads an
// identity, and the reason identitiesDeployment has outputs at all.
//
// At subscription scope the identities are declared inside that wrapper, so from the
// root they look like pre-existing resources. ARM resolves a reference() to a
// resource the current template does not declare during preflight — before the
// wrapper has created anything — and dependsOn does NOT defer it. That read fails
// the whole deployment with ResourceGroupNotFound while the resource group itself
// reports Created, because the read raced its creation. Coming back out as a nested
// deployment output is the only ordering ARM guarantees here, and is what the
// runner's vmssPrincipalId already relies on.
func identityPrincipalIDExpr(id azureOperationIdentity, scope armScope) string {
	if scope.subscription {
		return identityOutputRef(id, "PrincipalId")
	}
	return uamiPrincipalIDExpr(id.suffix, scope)
}

func identityClientIDExpr(id azureOperationIdentity, scope armScope) string {
	if scope.subscription {
		return identityOutputRef(id, "ClientId")
	}
	return uamiClientIDExpr(id.suffix, scope)
}

func identityOutputRef(id azureOperationIdentity, field string) string {
	return fmt.Sprintf("[reference('%s').outputs.%s.value]", identitiesDeploymentName, azureIdentityOutputKey(id, field))
}

// identityWrapperOutputs exposes every identity's principal and client ID from
// inside identitiesDeployment, where the identities are declared and reference()
// resolves normally.
func identityWrapperOutputs(ids []azureOperationIdentity) map[string]any {
	inner := armScope{}
	outputs := make(map[string]any, len(ids)*2)
	for _, id := range ids {
		outputs[azureIdentityOutputKey(id, "PrincipalId")] = map[string]any{
			"type":  "string",
			"value": uamiPrincipalIDExpr(id.suffix, inner),
		}
		outputs[azureIdentityOutputKey(id, "ClientId")] = map[string]any{
			"type":  "string",
			"value": uamiClientIDExpr(id.suffix, inner),
		}
	}
	return outputs
}

func uamiNameExpr(suffix string, scope armScope) string {
	return "[" + uamiNameInner(suffix, scope) + "]"
}

// uamiNameInner is unbracketed for embedding in another ARM expression; ARM does
// not allow nested [ ]. scope is that of the expression's reader — see the uami*Expr
// helpers below.
func uamiNameInner(suffix string, scope armScope) string {
	return fmt.Sprintf("format('{0}-%s', %s)", suffix, scope.nuonIDInner("nuonInstallID"))
}

// uamiResource declares an identity. scope is the root's, which only the tags
// depend on; the name and location are always read from wherever the declaration
// itself lands — the root at resource-group scope, the install-group wrapper at
// subscription scope — and both of those declare nuonInstallID and location as
// parameters. Reading them off the root's variables here fails the deploy with
// "The template variable 'nuonInstallID' is not found".
func (t *Templates) uamiResource(id azureOperationIdentity, scope armScope) map[string]any {
	return map[string]any{
		"type":       uamiResourceType,
		"apiVersion": "2023-01-31",
		"name":       uamiNameExpr(id.suffix, armScope{}),
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
		"nuonInstallID": {typ: "string", value: scope.nuonIDRef("nuonInstallID")},
		"location":      {typ: "string", value: scope.rootLocationRef()},
		"commonTags":    {typ: "object", value: "[variables('commonTags')]"},
	}, grouped, identityWrapperOutputs(ids))

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

// roleDeploymentNameExpr names a subscription-targeted role deployment, which always
// sits in the root — so it reads the install ID at the root's scope.
func roleDeploymentNameExpr(id azureOperationIdentity, scope armScope) string {
	return fmt.Sprintf("[format('{0}-%s-role', %s)]", azureRoleDeploymentToken(id), scope.nuonIDInner("nuonInstallID"))
}

// getOperationIdentityCustomRole always includes */register/action so the azurerm
// provider can register resource providers on apply (a subscription-scoped action
// previously held by the runner's system identity).
func (t *Templates) getOperationIdentityCustomRole(id azureOperationIdentity, scope armScope) map[string]any {
	roleActions := append([]string{"*/register/action"}, id.actions...)

	return map[string]any{
		"type":           "Microsoft.Resources/deployments",
		"apiVersion":     "2022-09-01",
		"name":           roleDeploymentNameExpr(id, scope),
		"subscriptionId": "[subscription().subscriptionId]",
		"location":       scope.locationExpr(),
		"dependsOn":      []string{identityDependency(id, scope)},
		"properties": map[string]any{
			"expressionEvaluationOptions": map[string]any{
				"scope": "inner",
			},
			"mode": "Incremental",
			"parameters": map[string]any{
				"roleName":    map[string]any{"value": fmt.Sprintf("[format('{0}-%s-role', %s)]", id.suffix, scope.nuonIDInner("nuonInstallID"))},
				"principalID": map[string]any{"value": identityPrincipalIDExpr(id, scope)},
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
			"name":       fmt.Sprintf("[guid(resourceGroup().id, %s, '%s')]", uamiNameInner(id.suffix, scope), guid),
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
		envVars = append(envVars, map[string]any{"name": envName, "value": identityClientIDExpr(id, scope)})

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
		deps = append(deps, roleDeploymentNameExpr(id, scope))
		if scope.subscription {
			// The built-in assignments are inside identitiesDeployment, which the
			// caller already depends on via operationIdentityAttachment.
			continue
		}
		for _, role := range id.builtInRoles {
			guid := azureBuiltInRoleGUID(role)
			deps = append(deps, fmt.Sprintf("[resourceId('Microsoft.Authorization/roleAssignments', guid(resourceGroup().id, %s, '%s'))]", uamiNameInner(id.suffix, armScope{}), guid))
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
