package arm

import (
	"slices"
	"strings"
	"testing"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks"
)

// Pins each armScope method's resource-group output to the literal that was
// inlined at the call site before armScope existed. The scope work is only safe
// if these never drift: a change here silently re-renders every existing Azure
// install's template, and for role assignments a changed name expression fails
// redeploys with RoleAssignmentExists.
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

func TestArmScope_InstallRGResource(t *testing.T) {
	if got := (armScope{}).installRGResource(); got != nil {
		t.Errorf("resource-group scope must not declare a resource group (InvalidTemplate), got %v", got)
	}

	rg := armScope{subscription: true}.installRGResource()
	if rg == nil {
		t.Fatal("subscription scope must declare the install resource group")
	}
	for k, want := range map[string]string{
		"type":       "Microsoft.Resources/resourceGroups",
		"apiVersion": "2021-04-01",
		"name":       "[parameters('installResourceGroupName')]",
		"location":   "[parameters('location')]",
		"tags":       "[variables('commonTags')]",
	} {
		if got := rg[k]; got != want {
			t.Errorf("installRGResource[%q] = %v, want %s", k, got, want)
		}
	}
}

func TestArmScope_TargetInstallRG(t *testing.T) {
	t.Run("no-op at resource group scope", func(t *testing.T) {
		dep := map[string]any{"name": "runnerDeployment", "dependsOn": []string{"vnetDeployment"}}
		armScope{}.targetInstallRG(dep)

		if _, ok := dep["resourceGroup"]; ok {
			t.Error("resourceGroup must not be set at resource-group scope")
		}
		if got := dep["dependsOn"].([]string); len(got) != 1 || got[0] != "vnetDeployment" {
			t.Errorf("dependsOn changed at resource-group scope: %v", got)
		}
	})

	// The runner already depends on the VNet and on each operation identity;
	// replacing dependsOn instead of merging would let the VMSS deploy before its
	// subnet or its identities exist.
	t.Run("merges into existing dependsOn", func(t *testing.T) {
		dep := map[string]any{"name": "runnerDeployment", "dependsOn": []string{"vnetDeployment", "uami"}}
		armScope{subscription: true}.targetInstallRG(dep)

		if got := dep["resourceGroup"]; got != "[parameters('installResourceGroupName')]" {
			t.Errorf("resourceGroup = %v", got)
		}
		want := []string{"vnetDeployment", "uami", "[resourceId('Microsoft.Resources/resourceGroups', parameters('installResourceGroupName'))]"}
		got := dep["dependsOn"].([]string)
		if len(got) != len(want) {
			t.Fatalf("dependsOn = %v, want %v", got, want)
		}
		for i := range want {
			if got[i] != want[i] {
				t.Errorf("dependsOn[%d] = %q, want %q", i, got[i], want[i])
			}
		}
	})

	t.Run("sets dependsOn when absent", func(t *testing.T) {
		dep := map[string]any{"name": "vnetDeployment"}
		armScope{subscription: true}.targetInstallRG(dep)

		if got := dep["dependsOn"].([]string); len(got) != 1 {
			t.Errorf("dependsOn = %v, want the resource group only", got)
		}
	})
}

// A subscription-targeted child deployment fails without an explicit location,
// which is the failure mode a custom VNet template hits when it declares its own
// resource groups.
func TestArmScope_TargetSubscription(t *testing.T) {
	dep := map[string]any{"name": "vnetDeployment"}
	armScope{}.targetSubscription(dep)
	if _, ok := dep["location"]; ok {
		t.Error("location must not be set at resource-group scope")
	}

	armScope{subscription: true}.targetSubscription(dep)
	if got := dep["location"]; got != "[parameters('location')]" {
		t.Errorf("location = %v, want [parameters('location')]", got)
	}
	if _, ok := dep["resourceGroup"]; ok {
		t.Error("a subscription-targeted deployment must not carry resourceGroup")
	}
}

// The built-in VNet and runner hold Nuon's own resources, so they must keep
// landing in the install resource group even when the root moves to subscription
// scope. Only a custom VNet template escapes to subscription scope.
func TestBuiltInDeployments_TargetInstallRGAtSubscriptionScope(t *testing.T) {
	tmpl := &Templates{cfg: &internal.Config{}}
	inp := minimalTemplateInput()

	for name, dep := range map[string]map[string]any{
		"vnet":   tmpl.getDefaultVNetDeployment(inp, armScope{subscription: true}),
		"runner": tmpl.getDefaultRunnerDeployment(inp, nil, armScope{subscription: true}),
	} {
		t.Run(name, func(t *testing.T) {
			if got := dep["resourceGroup"]; got != "[parameters('installResourceGroupName')]" {
				t.Errorf("resourceGroup = %v, want the install resource group", got)
			}
			if _, ok := dep["location"]; ok {
				t.Error("an RG-targeted deployment must not carry location")
			}

			deps, ok := dep["dependsOn"].([]string)
			if !ok {
				t.Fatalf("dependsOn missing or wrong type: %v", dep["dependsOn"])
			}
			if !slices.Contains(deps, "[resourceId('Microsoft.Resources/resourceGroups', parameters('installResourceGroupName'))]") {
				t.Errorf("dependsOn does not wait for the resource group: %v", deps)
			}
		})
	}

	// The runner's pre-existing VNet dependency has to survive the merge.
	runner := tmpl.getDefaultRunnerDeployment(inp, nil, armScope{subscription: true})
	if deps := runner["dependsOn"].([]string); !slices.Contains(deps, "vnetDeployment") {
		t.Errorf("runner lost its vnetDeployment dependency: %v", deps)
	}
}

func TestBuiltInDeployments_UnchangedAtResourceGroupScope(t *testing.T) {
	tmpl := &Templates{cfg: &internal.Config{}}
	inp := minimalTemplateInput()

	for name, dep := range map[string]map[string]any{
		"vnet":   tmpl.getDefaultVNetDeployment(inp, armScope{}),
		"runner": tmpl.getDefaultRunnerDeployment(inp, nil, armScope{}),
	} {
		t.Run(name, func(t *testing.T) {
			for _, key := range []string{"resourceGroup", "location"} {
				if _, ok := dep[key]; ok {
					t.Errorf("%s must not be set at resource-group scope", key)
				}
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
