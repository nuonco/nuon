package arm

import (
	"strings"
	"testing"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks"
)

// Pins each armScope method's resource-group output to the literal that was
// inlined at the call site before armScope existed. The scope
// work is only safe if these never drift: a change here silently re-renders every
// existing Azure install's template, and for role assignments a changed name
// expression fails redeploys with RoleAssignmentExists.
func TestArmScope_ResourceGroupExpressionsMatchLegacyLiterals(t *testing.T) {
	s := armScope{}

	for _, tc := range []struct{ name, got, want string }{
		{"rootSchema", s.rootSchema(), "https://schema.management.azure.com/schemas/2019-04-01/deploymentTemplate.json#"},
		{"locationExpr", s.locationExpr(), "[resourceGroup().location]"},
		{"rgNameExpr", s.rgNameExpr(), "[resourceGroup().name]"},
		{"rgIDExpr", s.rgIDExpr(), "[resourceGroup().id]"},
		{
			"rgResourceIDExpr",
			s.rgResourceIDExpr("Microsoft.KeyVault/vaults", keyVaultNameInner),
			"[resourceId('Microsoft.KeyVault/vaults', take(format('{0}', parameters('nuonInstallID')), 24))]",
		},
	} {
		if tc.got != tc.want {
			t.Errorf("%s at resource-group scope:\n got: %s\nwant: %s", tc.name, tc.got, tc.want)
		}
	}
}

func TestArmScope_SubscriptionExpressions(t *testing.T) {
	s := armScope{subscription: true}

	for _, tc := range []struct{ name, got, want string }{
		{"rootSchema", s.rootSchema(), "https://schema.management.azure.com/schemas/2018-05-01/subscriptionDeploymentTemplate.json#"},
		{"locationExpr", s.locationExpr(), "[parameters('location')]"},
		{"rgNameExpr", s.rgNameExpr(), "[parameters('installResourceGroupName')]"},
		{"rgIDExpr", s.rgIDExpr(), "[format('{0}/resourceGroups/{1}', subscription().id, parameters('installResourceGroupName'))]"},
		{
			"rgResourceIDExpr",
			s.rgResourceIDExpr("Microsoft.KeyVault/vaults", keyVaultNameInner),
			"[resourceId(subscription().subscriptionId, parameters('installResourceGroupName'), 'Microsoft.KeyVault/vaults', take(format('{0}', parameters('nuonInstallID')), 24))]",
		},
	} {
		if tc.got != tc.want {
			t.Errorf("%s at subscription scope:\n got: %s\nwant: %s", tc.name, tc.got, tc.want)
		}
		assertNoNestedBrackets(t, []byte(tc.got))
	}
}

func TestScopeFor(t *testing.T) {
	for _, tc := range []struct {
		scope app.StackDeploymentScope
		want  bool
	}{
		{"", false},
		{app.StackDeploymentScopeResourceGroup, false},
		{app.StackDeploymentScopeSubscription, true},
	} {
		inp := &stacks.TemplateInput{DeploymentScope: tc.scope}
		if got := scopeFor(inp).subscription; got != tc.want {
			t.Errorf("scopeFor(%q).subscription = %v, want %v", tc.scope, got, tc.want)
		}
	}
}

// TestGetAzureTemplate_RGScopeUnchanged is the backwards-compatibility contract:
// an app that omits deployment_scope and one that sets it explicitly to
// resource_group must render byte-identical templates, and both must match what
// the renderer produced before the field existed.
func TestGetAzureTemplate_RGScopeUnchanged(t *testing.T) {
	tmpl := &Templates{cfg: &internal.Config{}}

	for _, fixture := range []struct {
		name  string
		build func() *stacks.TemplateInput
	}{
		{"legacy system identity", minimalTemplateInput},
		{"per-operation identities", azureRolesTemplateInput},
	} {
		t.Run(fixture.name, func(t *testing.T) {
			omitted := fixture.build()
			omitted.DeploymentScope = ""

			explicit := fixture.build()
			explicit.DeploymentScope = app.StackDeploymentScopeResourceGroup

			omittedBytes, omittedSum, err := tmpl.Template(omitted)
			if err != nil {
				t.Fatalf("render with omitted scope: %v", err)
			}
			explicitBytes, explicitSum, err := tmpl.Template(explicit)
			if err != nil {
				t.Fatalf("render with explicit resource_group: %v", err)
			}

			if omittedSum != explicitSum {
				t.Errorf("checksum differs between omitted and explicit resource_group:\n omitted:  %s\n explicit: %s", omittedSum, explicitSum)
			}
			if string(omittedBytes) != string(explicitBytes) {
				t.Error("rendered template differs between omitted and explicit resource_group")
			}

			// The literals armScope replaced must still be present verbatim.
			body := string(omittedBytes)
			for _, want := range []string{
				`"$schema": "https://schema.management.azure.com/schemas/2019-04-01/deploymentTemplate.json#"`,
				`[resourceGroup().location]`,
				`[resourceGroup().id]`,
				`[resourceGroup().name]`,
				`[resourceId('Microsoft.KeyVault/vaults', take(format('{0}', parameters('nuonInstallID')), 24))]`,
			} {
				if !strings.Contains(body, want) {
					t.Errorf("rendered template no longer contains %s", want)
				}
			}

			if strings.Contains(body, installRGParamName) {
				t.Errorf("%s must not appear at resource-group scope", installRGParamName)
			}
		})
	}
}

// The guard exists because deployment_scope validates at sync before the
// subscription-scoped root is built: without it a customer gets a template ARM
// rejects with an opaque InvalidTemplate. Delete this test with the guard.
func TestGetAzureTemplate_SubscriptionScopeNotImplementedYet(t *testing.T) {
	tmpl := &Templates{cfg: &internal.Config{}}

	inp := minimalTemplateInput()
	inp.DeploymentScope = app.StackDeploymentScopeSubscription

	_, err := tmpl.getAzureTemplate(inp)
	if err == nil {
		t.Fatal("expected subscription scope to be rejected until the root template is implemented")
	}
	if !strings.Contains(err.Error(), "deployment_scope") {
		t.Errorf("error should name deployment_scope so the operator knows what to unset, got: %v", err)
	}
}
