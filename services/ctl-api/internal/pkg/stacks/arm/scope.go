package arm

import (
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks"
)

const (
	rgTemplateSchema           = "https://schema.management.azure.com/schemas/2019-04-01/deploymentTemplate.json#"
	subscriptionTemplateSchema = "https://schema.management.azure.com/schemas/2018-05-01/subscriptionDeploymentTemplate.json#"

	// runnerGrantsDeploymentName is the nested deployment the runner's system-identity
	// role assignments move into at subscription scope.
	runnerGrantsDeploymentName = "runnerGrantsDeployment"

	// phoneHomeDeploymentName is the nested deployment the phone-home deploymentScripts
	// resource moves into at subscription scope.
	phoneHomeDeploymentName = "phoneHomeDeployment"

	// rgScopeVNetDeploymentName is the VNet nested deployment's name at resource-group
	// scope, where it needs no install namespace — see armScope.vnetDeploymentName.
	rgScopeVNetDeploymentName = "vnetDeployment"

	// maxARMDeploymentNameLen is ARM's cap on a deployment resource's name. Only the
	// names derived from a customer-supplied stack name can reach it; everything else
	// this package generates is a fixed suffix on the 26-character install ID.
	maxARMDeploymentNameLen = 64

	// installRGVarName names the install's resource group in the root template. At
	// subscription scope there is no ambient resource group, so every expression that
	// used to resolve against one has to name it explicitly. At RG scope it is not
	// declared at all.
	//
	// Deliberately a variable rather than a parameter: the portal builds its
	// deployment form from a template's parameters and gives no way to hide one, so a
	// parameter here would put a Nuon-internal value in front of the customer as an
	// editable field. The name is renderer-owned and known at render time, so nothing
	// is lost by baking it.
	installRGVarName = "installResourceGroupName"

	// locationVarName is the install's region in the root at subscription scope. A
	// variable for the same reason installRGVarName is — see rootLocationRef.
	locationVarName = "location"
)

// installResourceGroupName is the resource group Nuon's own install resources live
// in. This is the single source of truth for the name.
//
// It is the same `<install-id>-rg` the customer creates by hand at resource-group
// scope, which is what lets subscription scope drop the `az group create`
// prerequisite without changing anything downstream: the Key Vault, the runner and
// the phone-home all still resolve against a group by this name.
func installResourceGroupName(installID string) string {
	return installID + "-rg"
}

// vnetDeploymentName is the nested deployment that provisions the install's VNet.
// Every reference() and dependsOn that reads its outputs has to agree with this, so
// it is the single source of truth for the name.
//
// A custom VNet template runs at subscription scope, and there a deployment record
// is identified by name alone within the subscription — its location is immutable
// once created. A constant name therefore collides between two installs in the same
// subscription: the second fails with InvalidDeploymentLocation if it is in another
// region, and silently updates the first install's record if it is not. Namespacing
// by install ID gives each install its own.
//
// At resource-group scope the record lives in the install's own group, so the bare
// name is already unique per install and stays as it was.
func (s armScope) vnetDeploymentName(installID string) string {
	if !s.subscription {
		return rgScopeVNetDeploymentName
	}
	return installID + "-vnet-deployment"
}

// customStackDeploymentName is the ARM resource name of a custom nested stack's
// deployment, namespaced for the same reason vnetDeploymentName is — these inherit
// the root scope, so at subscription scope they too become subscription-level
// records. The hazard is worse here: the name comes from app config, so two
// installs of one app collide every time rather than only when they share a
// subscription by accident.
//
// Only the resource name changes. The customer's stack name remains
// customDeploymentOutputs.StackName, which is what keys the phone-home payload, so
// the custom_nested_stacks.<name>.outputs.<key> shape is unaffected.
func (s armScope) customStackDeploymentName(installID, sanitizedStackName string) string {
	if !s.subscription {
		return sanitizedStackName
	}
	return installID + "-" + sanitizedStackName
}

// customStackRoleKey seeds the role definition and assignment a custom stack's
// managed identity needs. Unlike customStackDeploymentName it is namespaced at both
// scopes — see getCustomDeploymentRoleAssignment for why.
func customStackRoleKey(installID, sanitizedStackName string) string {
	return installID + "-" + sanitizedStackName
}

// customStackRoleDeploymentName is the longest name derived from a stack's own, so
// it is what bounds how long that name may be — see validateCustomStackNameLengths.
func customStackRoleDeploymentName(installID, sanitizedStackName string) string {
	return customStackRoleKey(installID, sanitizedStackName) + "-identity-role"
}

// armScope is the ARM scope the root template renders at.
//
// Every method returns exactly the string the renderer emitted before this type
// existed whenever the scope is resource group. That is what keeps apps which do
// not opt in byte-identical, and it is asserted by the golden tests rather than
// left to review.
type armScope struct{ subscription bool }

func scopeFor(inp *stacks.TemplateInput) armScope {
	return armScope{subscription: inp.DeploymentScope == app.StackDeploymentScopeSubscription}
}

func (s armScope) rootSchema() string {
	if s.subscription {
		return subscriptionTemplateSchema
	}
	return rgTemplateSchema
}

// locationExpr is the location for a root resource that had no explicit location
// of its own at resource-group scope, where it inherited the ambient group's.
func (s armScope) locationExpr() string {
	if s.subscription {
		return s.rootLocationRef()
	}
	return "[resourceGroup().location]"
}

// rootLocationRef is how the root refers to the install's region.
//
// At subscription scope this is a variable rather than a parameter, for the same
// reason as installResourceGroupName: the portal renders one form field per
// parameter with no way to hide one, and the region is not customer-configurable —
// it is whatever Nuon recorded on the Azure account, and the sandbox and components
// already assume that value.
//
// Only for references evaluated in the root. A wrapped resource is inside a nested
// template that declares its own `location` parameter, so it keeps using
// parameters('location').
func (s armScope) rootLocationRef() string {
	if s.subscription {
		return fmt.Sprintf("[variables('%s')]", locationVarName)
	}
	return "[parameters('location')]"
}

func (s armScope) rgNameExpr() string {
	if s.subscription {
		return fmt.Sprintf("[variables('%s')]", installRGVarName)
	}
	return "[resourceGroup().name]"
}

// rgIDExpr builds the install resource group's ID. subscription().id is already
// "/subscriptions/{guid}", so it is the right thing to concatenate onto —
// re-prefixing subscription().subscriptionId yields a doubled segment.
//
// subscriptionResourceId('Microsoft.Resources/resourceGroups', …) is not a
// substitute: it emits a providers/-qualified form, not the canonical RG ID.
func (s armScope) rgIDExpr() string {
	if s.subscription {
		return fmt.Sprintf("[format('{0}/resourceGroups/{1}', subscription().id, variables('%s'))]", installRGVarName)
	}
	return "[resourceGroup().id]"
}

// installRGResource declares the install resource group in the root, and returns
// nil at resource-group scope where the customer creates it by hand before
// deploying and nothing in the template may declare it — an RG in an RG-scoped
// root's resources is InvalidTemplate.
func (s armScope) installRGResource() map[string]any {
	if !s.subscription {
		return nil
	}
	return map[string]any{
		"type":       "Microsoft.Resources/resourceGroups",
		"apiVersion": "2021-04-01",
		"name":       s.rgNameExpr(),
		"location":   s.locationExpr(),
		"tags":       "[variables('commonTags')]",
		// Empty but present, matching Microsoft's documented example for declaring a
		// resource group from a subscription-scoped template.
		"properties": map[string]any{},
	}
}

// keyVaultDependsOn orders anything that needs the install Key Vault to exist —
// the runner's role assignment on it, and the phone-home that reports its ID.
// Empty at resource-group scope, where the customer creates the vault beforehand and
// there is no deployment to wait on.
func (s armScope) keyVaultDependsOn() []string {
	if !s.subscription {
		return nil
	}
	return []string{keyVaultDeploymentName}
}

// installRGDependsOn is the dependency on the declared install resource group.
// Empty at resource-group scope, where there is nothing to depend on.
func (s armScope) installRGDependsOn() []string {
	if !s.subscription {
		return nil
	}
	return []string{fmt.Sprintf("[resourceId('Microsoft.Resources/resourceGroups', variables('%s'))]", installRGVarName)}
}

// targetInstallRG points a nested deployment at the install resource group and
// makes it wait for the group to exist. A no-op at resource-group scope, where
// the deployment already runs in that group by virtue of the root's scope.
//
// Mutates deployment in place, and merges into any dependsOn already set rather
// than replacing it — the runner deployment in particular already depends on the
// VNet and on each operation identity.
func (s armScope) targetInstallRG(deployment map[string]any) {
	if !s.subscription {
		return
	}

	deployment["resourceGroup"] = s.rgNameExpr()

	deps := s.installRGDependsOn()
	if existing, ok := deployment["dependsOn"].([]string); ok {
		deps = append(existing, deps...)
	}
	deployment["dependsOn"] = deps
}

// targetSubscription points a nested deployment at the subscription rather than
// at a resource group, which is what lets the template it runs declare its own
// resource groups. A subscription-targeted child requires an explicit location:
// without it the deployment fails.
//
// A no-op at resource-group scope, where the root cannot host one.
func (s armScope) targetSubscription(deployment map[string]any) {
	if !s.subscription {
		return
	}
	deployment["location"] = s.locationExpr()
}

// rgResourceIDExpr addresses a resource in the install resource group. nameInner
// is an unbracketed ARM expression, because ARM does not allow nested [ ].
//
// At subscription scope this uses the unambiguous four-argument overload rather
// than the three-argument one, so resolution never depends on the caller's
// ambient scope — a bare resourceId() at subscription scope produces a malformed
// ID rather than failing loudly.
func (s armScope) rgResourceIDExpr(resourceType, nameInner string) string {
	return "[" + s.rgResourceIDInner(resourceType, nameInner) + "]"
}

// rgResourceIDInner is rgResourceIDExpr without the surrounding brackets, for
// embedding in a larger expression — ARM does not allow nested [ ].
func (s armScope) rgResourceIDInner(resourceType, nameInner string) string {
	if s.subscription {
		return fmt.Sprintf("resourceId(subscription().subscriptionId, variables('%s'), '%s', %s)",
			installRGVarName, resourceType, nameInner)
	}
	return fmt.Sprintf("resourceId('%s', %s)", resourceType, nameInner)
}

// nuonIDNames are the Nuon-managed identifiers the template needs but the customer
// must never be invited to change.
var nuonIDNames = []string{"nuonInstallID", "nuonOrgID", "nuonAppID"}

// nuonIDRef is how the root refers to one of the Nuon-managed IDs, bracketed.
//
// At subscription scope they are variables rather than parameters, so the portal's
// deployment form does not offer them as editable fields. Inner templates are
// unaffected: each declares its own nuonInstallID parameter and receives the value,
// so inside one the reference is genuinely parameters('nuonInstallID').
func (s armScope) nuonIDRef(name string) string {
	return "[" + s.nuonIDInner(name) + "]"
}

// nuonIDInner is nuonIDRef unbracketed, for embedding in a larger expression.
func (s armScope) nuonIDInner(name string) string {
	if s.subscription {
		return fmt.Sprintf("variables('%s')", name)
	}
	return fmt.Sprintf("parameters('%s')", name)
}

// keyVaultNameInner is the install Key Vault's name as an unbracketed expression.
// Nothing in the stack creates the vault — the template only ever references it by
// this convention.
//
// Call with resource-group scope for anything that ends up inside a wrapper: the
// Key Vault role assignment embeds this in a guid(), and changing it would rename a
// live assignment and fail redeploys with RoleAssignmentExists.
func (s armScope) keyVaultNameInner() string {
	return fmt.Sprintf("take(format('{0}', %s), 24)", s.nuonIDInner("nuonInstallID"))
}

// innerCommonTagsExpr is how a resource that wrapInInstallRG may relocate should
// reference the standard tags. At resource-group scope such a resource stays in
// the root, where commonTags is a variable; at subscription scope it moves into a
// nested template, where the root's variables are out of reach and the tags
// arrive as a parameter instead.
func (s armScope) innerCommonTagsExpr() string {
	if s.subscription {
		return "[parameters('commonTags')]"
	}
	return "[variables('commonTags')]"
}

// dependOn prepends dependencies to a resource's existing dependsOn. Used to hand
// a wrapper the dependencies its contents can no longer express themselves.
func dependOn(resource map[string]any, deps []string) {
	if len(deps) == 0 {
		return
	}
	existing, _ := resource["dependsOn"].([]string)
	resource["dependsOn"] = append(append([]string{}, deps...), existing...)
}

// nestedParam is one parameter threaded into a wrapped deployment: the expression
// evaluated at the outer scope, plus the type the inner template declares.
type nestedParam struct {
	typ   string
	value any
}

// wrapInInstallRG relocates resource-group-scoped resources into the install
// resource group via a nested deployment, and returns them untouched at
// resource-group scope where they already live there.
//
// The wrapper uses inner expression evaluation, which is what makes wrapping
// safe: resourceGroup() resolves to the install group again inside it, so
// guid(resourceGroup().id, …) role assignment names come out byte-identical to
// resource-group scope. Rewriting those names instead of wrapping them would
// fail every redeploy with RoleAssignmentExists.
//
// The flip side of inner evaluation is that nothing from the root is visible:
// every value the resources need must be declared in params.
// wrapInInstallRG relocates resources into the install resource group at
// subscription scope, and is a no-op at resource-group scope where they already sit
// in the root.
//
// outputs is how anything left behind in the root reads a value off a wrapped
// resource. It has to be, rather than referencing the resource directly: a
// reference() to a resource the current template does not declare is resolved by ARM
// during preflight, so it races the wrapper that creates it.
func (s armScope) wrapInInstallRG(name string, params map[string]nestedParam, resources []any, outputs map[string]any) []any {
	if !s.subscription {
		return resources
	}

	outerParams := make(map[string]any, len(params))
	innerParams := make(map[string]any, len(params))
	for paramName, p := range params {
		outerParams[paramName] = map[string]any{"value": p.value}
		innerParams[paramName] = map[string]any{"type": p.typ}
	}

	inner := map[string]any{
		"$schema":        rgTemplateSchema,
		"contentVersion": "1.0.0.0",
		"parameters":     innerParams,
		"resources":      resources,
	}
	if len(outputs) > 0 {
		inner["outputs"] = outputs
	}

	return []any{map[string]any{
		"type":          "Microsoft.Resources/deployments",
		"apiVersion":    "2022-09-01",
		"name":          name,
		"resourceGroup": s.rgNameExpr(),
		"dependsOn":     s.installRGDependsOn(),
		"properties": map[string]any{
			"mode": "Incremental",
			"expressionEvaluationOptions": map[string]any{
				"scope": "inner",
			},
			"parameters": outerParams,
			"template":   inner,
		},
	}}
}
