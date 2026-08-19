package arm

import (
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks"
)

const (
	rgTemplateSchema           = "https://schema.management.azure.com/schemas/2019-04-01/deploymentTemplate.json#"
	subscriptionTemplateSchema = "https://schema.management.azure.com/schemas/2018-05-01/subscriptionDeploymentTemplate.json#"

	// keyVaultNameInner is the install Key Vault's name as an unbracketed ARM
	// expression. Nothing in the stack creates the vault — the template only ever
	// references it by this convention, which is why its resource ID has to be
	// built scope-correctly rather than left to an ambient resource group.
	keyVaultNameInner = "take(format('{0}', parameters('nuonInstallID')), 24)"

	// installRGParamName names the install's resource group in the root template.
	// At subscription scope there is no ambient resource group, so every expression
	// that used to resolve against one has to name it explicitly. At RG scope the
	// parameter is not declared at all.
	installRGParamName = "installResourceGroupName"
)

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

func (s armScope) locationExpr() string {
	if s.subscription {
		return "[parameters('location')]"
	}
	return "[resourceGroup().location]"
}

func (s armScope) rgNameExpr() string {
	if s.subscription {
		return fmt.Sprintf("[parameters('%s')]", installRGParamName)
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
		return fmt.Sprintf("[format('{0}/resourceGroups/{1}', subscription().id, parameters('%s'))]", installRGParamName)
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
	}
}

// installRGDependsOn is the dependency on the declared install resource group.
// Empty at resource-group scope, where there is nothing to depend on.
func (s armScope) installRGDependsOn() []string {
	if !s.subscription {
		return nil
	}
	return []string{fmt.Sprintf("[resourceId('Microsoft.Resources/resourceGroups', parameters('%s'))]", installRGParamName)}
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
	if s.subscription {
		return fmt.Sprintf("[resourceId(subscription().subscriptionId, parameters('%s'), '%s', %s)]",
			installRGParamName, resourceType, nameInner)
	}
	return fmt.Sprintf("[resourceId('%s', %s)]", resourceType, nameInner)
}
