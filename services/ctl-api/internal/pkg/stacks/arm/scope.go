package arm

import (
	"fmt"
	"strings"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks"
)

const (
	rgTemplateSchema           = "https://schema.management.azure.com/schemas/2019-04-01/deploymentTemplate.json#"
	subscriptionTemplateSchema = "https://schema.management.azure.com/schemas/2018-05-01/subscriptionDeploymentTemplate.json#"

	runnerGrantsDeploymentName = "runnerGrantsDeployment"

	phoneHomeDeploymentName = "phoneHomeDeployment"

	rgScopeVNetDeploymentName = "vnetDeployment"

	maxARMDeploymentNameLen = 64

	installRGVarName = "installResourceGroupName"

	locationVarName = "location"
)

func installResourceGroupName(installID string) string {
	return installID + "-rg"
}

func (s armScope) vnetDeploymentName(installID string) string {
	if !s.subscription {
		return rgScopeVNetDeploymentName
	}
	return installID + "-vnet-deployment"
}

func (s armScope) customStackDeploymentName(installID, sanitizedStackName string) string {
	if !s.subscription {
		return sanitizedStackName
	}
	return installID + "-" + sanitizedStackName
}

func customStackRoleKey(installID, sanitizedStackName string) string {
	return installID + "-" + sanitizedStackName
}

func customStackRoleDeploymentName(installID, sanitizedStackName string) string {
	return customStackRoleKey(installID, sanitizedStackName) + "-identity-role"
}

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
		return s.rootLocationRef()
	}
	return "[resourceGroup().location]"
}

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

func (s armScope) rgIDExpr() string {
	if s.subscription {
		return fmt.Sprintf("[format('{0}/resourceGroups/{1}', subscription().id, variables('%s'))]", installRGVarName)
	}
	return "[resourceGroup().id]"
}

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
		"properties": map[string]any{},
	}
}

func (s armScope) keyVaultDependsOn() []string {
	if !s.subscription {
		return nil
	}
	return []string{keyVaultDeploymentName}
}

func (s armScope) installRGDependsOn() []string {
	if !s.subscription {
		return nil
	}
	return []string{fmt.Sprintf("[resourceId('Microsoft.Resources/resourceGroups', variables('%s'))]", installRGVarName)}
}

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

func (s armScope) targetSubscription(deployment map[string]any) {
	if !s.subscription {
		return
	}
	deployment["location"] = s.locationExpr()
}

// why: rgResourceIDExpr addresses a resource in the install resource group. nameInner
// is an unbracketed ARM expression, because ARM does not allow nested [ ].
//
// At subscription scope this uses the unambiguous four-argument overload rather
// than the three-argument one, so resolution never depends on the caller's
// ambient scope — a bare resourceId() at subscription scope produces a malformed
// ID rather than failing loudly.
func (s armScope) rgResourceIDExpr(resourceType, nameInner string) string {
	return "[" + s.rgResourceIDInner(resourceType, nameInner) + "]"
}

func (s armScope) rgResourceIDInner(resourceType, nameInner string) string {
	if s.subscription {
		return fmt.Sprintf("resourceId(subscription().subscriptionId, variables('%s'), '%s', %s)",
			installRGVarName, resourceType, nameInner)
	}
	return fmt.Sprintf("resourceId('%s', %s)", resourceType, nameInner)
}

// why: nuonIDNames are the Nuon-managed identifiers the template needs but the customer
// must never be invited to change.
var nuonIDNames = []string{"nuonInstallID", "nuonOrgID", "nuonAppID"}

func (s armScope) nuonIDRef(name string) string {
	return "[" + s.nuonIDInner(name) + "]"
}

func (s armScope) nuonIDInner(name string) string {
	if s.subscription {
		return fmt.Sprintf("variables('%s')", name)
	}
	return fmt.Sprintf("parameters('%s')", name)
}

func (s armScope) keyVaultNameInner() string {
	return fmt.Sprintf("take(format('{0}', %s), 24)", s.nuonIDInner("nuonInstallID"))
}

func (s armScope) innerCommonTagsExpr() string {
	if s.subscription {
		return "[parameters('commonTags')]"
	}
	return "[variables('commonTags')]"
}

func dependOn(resource map[string]any, deps []string) {
	if len(deps) == 0 {
		return
	}
	existing, _ := resource["dependsOn"].([]string)
	resource["dependsOn"] = append(append([]string{}, deps...), existing...)
}

type nestedParam struct {
	typ   string
	value any
}

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

func isSubscriptionScopedTemplate(tmpl *armTemplateShape) bool {
	if tmpl == nil {
		return false
	}
	return strings.TrimSuffix(tmpl.Schema, "#") == strings.TrimSuffix(subscriptionTemplateSchema, "#")
}
