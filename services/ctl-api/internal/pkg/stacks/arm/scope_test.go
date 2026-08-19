package arm

import (
	"encoding/json"
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
		{"rootLocationRef", s.rootLocationRef(), "[parameters('location')]"},
		{
			"rgResourceIDExpr",
			s.rgResourceIDExpr("Microsoft.KeyVault/vaults", s.keyVaultNameInner()),
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
		{"locationExpr", s.locationExpr(), "[variables('location')]"},
		{"rootLocationRef", s.rootLocationRef(), "[variables('location')]"},
		{"rgNameExpr", s.rgNameExpr(), "[variables('installResourceGroupName')]"},
		{"rgIDExpr", s.rgIDExpr(), "[format('{0}/resourceGroups/{1}', subscription().id, variables('installResourceGroupName'))]"},
		{
			"rgResourceIDExpr",
			s.rgResourceIDExpr("Microsoft.KeyVault/vaults", s.keyVaultNameInner()),
			"[resourceId(subscription().subscriptionId, variables('installResourceGroupName'), 'Microsoft.KeyVault/vaults', take(format('{0}', variables('nuonInstallID')), 24))]",
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

			if strings.Contains(body, installRGVarName) {
				t.Errorf("%s must not appear at resource-group scope", installRGVarName)
			}

			// The install resource group is created by the template only at
			// subscription scope. At resource-group scope the customer creates it
			// before deploying, and declaring one here is InvalidTemplate.
			armTmpl, err := tmpl.getAzureTemplate(fixture.build())
			if err != nil {
				t.Fatalf("render: %v", err)
			}
			if got := countResourceType(armTmpl.Resources, "Microsoft.Resources/resourceGroups"); got != 0 {
				t.Errorf("resource-group scope declared %d resource groups; it must declare none", got)
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
		"name":       "[variables('installResourceGroupName')]",
		"location":   "[variables('location')]",
		"tags":       "[variables('commonTags')]",
	} {
		if got := rg[k]; got != want {
			t.Errorf("installRGResource[%q] = %v, want %s", k, got, want)
		}
	}
	if _, ok := rg["properties"].(map[string]any); !ok {
		t.Errorf("installRGResource has no properties object: %v", rg["properties"])
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

		if got := dep["resourceGroup"]; got != "[variables('installResourceGroupName')]" {
			t.Errorf("resourceGroup = %v", got)
		}
		want := []string{"vnetDeployment", "uami", "[resourceId('Microsoft.Resources/resourceGroups', variables('installResourceGroupName'))]"}
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
	if got := dep["location"]; got != "[variables('location')]" {
		t.Errorf("location = %v, want [variables('location')]", got)
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
			if got := dep["resourceGroup"]; got != "[variables('installResourceGroupName')]" {
				t.Errorf("resourceGroup = %v, want the install resource group", got)
			}
			if _, ok := dep["location"]; ok {
				t.Error("an RG-targeted deployment must not carry location")
			}

			deps, ok := dep["dependsOn"].([]string)
			if !ok {
				t.Fatalf("dependsOn missing or wrong type: %v", dep["dependsOn"])
			}
			if !slices.Contains(deps, "[resourceId('Microsoft.Resources/resourceGroups', variables('installResourceGroupName'))]") {
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

func subscriptionTemplateInput() *stacks.TemplateInput {
	inp := azureRolesTemplateInput()
	inp.DeploymentScope = app.StackDeploymentScopeSubscription
	return inp
}

func TestGetAzureTemplate_SubscriptionScopeRoot(t *testing.T) {
	tmpl := &Templates{cfg: &internal.Config{}}

	armTmpl, err := tmpl.getAzureTemplate(subscriptionTemplateInput())
	if err != nil {
		t.Fatalf("render at subscription scope: %v", err)
	}

	if armTmpl.Schema != subscriptionTemplateSchema {
		t.Errorf("root $schema = %s, want the subscription schema", armTmpl.Schema)
	}

	// The portal renders a form from the parameters, so a parameter without a
	// default is an empty required field the customer has to guess.
	// A variable, so the portal never renders a form field for it, holding a literal
	// that matches the group name the resource-group-scope flow tells customers to
	// create by hand — nothing downstream moves.
	want := minimalTemplateInput().Install.ID + "-rg"
	if got := armTmpl.Variables[installRGVarName]; got != want {
		t.Errorf("variable %s = %v, want the literal %q", installRGVarName, got, want)
	}
	if _, ok := armTmpl.Parameters[installRGVarName]; ok {
		t.Errorf("%s must not be a parameter: the portal would show it as an editable field", installRGVarName)
	}
	for name, param := range armTmpl.Parameters {
		if param.DefaultValue == nil {
			t.Errorf("parameter %q has no default", name)
		}
	}

	if got := countResourceType(armTmpl.Resources, "Microsoft.Resources/resourceGroups"); got != 1 {
		t.Errorf("expected exactly one resource group declaration, got %d", got)
	}
}

// Both values are Nuon-internal and not customer-configurable, and the plain
// deployment blade renders a field for every parameter with no way to hide one. As
// variables they stay out of the form entirely.
func TestGetAzureTemplate_SubscriptionScopeHidesNuonInternalsFromTheForm(t *testing.T) {
	tmpl := &Templates{cfg: &internal.Config{}}
	inp := subscriptionTemplateInput()

	armTmpl, err := tmpl.getAzureTemplate(inp)
	if err != nil {
		t.Fatalf("render at subscription scope: %v", err)
	}

	for name, want := range map[string]string{
		installRGVarName: installResourceGroupName(inp.Install.ID),
		locationVarName:  inp.Install.AzureAccount.Location,
	} {
		if got := armTmpl.Variables[name]; got != want {
			t.Errorf("variable %s = %v, want %q", name, got, want)
		}
		if _, ok := armTmpl.Parameters[name]; ok {
			t.Errorf("%s must not be a parameter: the portal would render it as an editable field", name)
		}
	}

	// Nothing may still be reaching for the parameter that no longer exists.
	blob, err := json.Marshal(armTmpl)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if strings.Contains(string(blob), "parameters('installResourceGroupName')") {
		t.Error("template references parameters('installResourceGroupName'), which is a variable at subscription scope")
	}
}

// The portal prompts for a Region that a quick link cannot pre-set, and a
// subscription-scoped deployment record's location is immutable — so Nuon has to
// learn where the customer actually deployed rather than assume.
func TestGetAzureTemplate_DeploymentLocationReportedOnlyAtSubscriptionScope(t *testing.T) {
	tmpl := &Templates{cfg: &internal.Config{}}

	sub := tmpl.getPhoneHomeResources(subscriptionTemplateInput(), nil, nil, armScope{subscription: true})
	blob, err := json.Marshal(sub)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	for _, want := range []string{"DEPLOYMENT_LOCATION", "[deployment().location]", "deployment_location"} {
		if !strings.Contains(string(blob), want) {
			t.Errorf("subscription scope phone-home missing %s", want)
		}
	}

	// deployment().location does not exist at resource-group scope, and emitting the
	// field there would drift the golden template for every existing install.
	rg := tmpl.getPhoneHomeResources(minimalTemplateInput(), nil, nil, armScope{})
	rgBlob, err := json.Marshal(rg)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	for _, unwanted := range []string{"DEPLOYMENT_LOCATION", "deployment().location", "deployment_location"} {
		if strings.Contains(string(rgBlob), unwanted) {
			t.Errorf("resource-group scope phone-home must not contain %s", unwanted)
		}
	}
}

// The assertion that proves the phase: at subscription scope there is no ambient
// resource group, so resourceGroup() may only appear inside the inline template of
// an RG-targeted nested deployment. Anywhere else it is a render-time bug that
// surfaces as an opaque ARM error at the customer.
func TestGetAzureTemplate_SubscriptionScopeResourceGroupFuncOnlyInsideWrappers(t *testing.T) {
	tmpl := &Templates{cfg: &internal.Config{}}

	armTmpl, err := tmpl.getAzureTemplate(subscriptionTemplateInput())
	if err != nil {
		t.Fatalf("render at subscription scope: %v", err)
	}

	for _, r := range armTmpl.Resources {
		res, ok := r.(map[string]any)
		if !ok {
			continue
		}

		// An RG-targeted nested deployment re-establishes an ambient resource group
		// for everything in its inline template, so skip its contents.
		if res["resourceGroup"] != nil {
			continue
		}

		blob, err := json.Marshal(res)
		if err != nil {
			t.Fatalf("marshal resource: %v", err)
		}
		if strings.Contains(string(blob), "resourceGroup()") {
			t.Errorf("resourceGroup() outside an RG-targeted wrapper, in %v:\n%s", res["name"], blob)
		}
	}
}

func TestGetAzureTemplate_SubscriptionScopeWrapsRGScopedResources(t *testing.T) {
	tmpl := &Templates{cfg: &internal.Config{}}

	armTmpl, err := tmpl.getAzureTemplate(subscriptionTemplateInput())
	if err != nil {
		t.Fatalf("render at subscription scope: %v", err)
	}

	// None of these types are subscription-deployable, so any left directly in the
	// root would be rejected as InvalidTemplate.
	for _, resourceType := range []string{
		uamiResourceType,
		"Microsoft.Authorization/roleAssignments",
		"Microsoft.Resources/deploymentScripts",
	} {
		if got := countResourceType(armTmpl.Resources, resourceType); got != 0 {
			t.Errorf("%d %s left directly in a subscription-scoped root", got, resourceType)
		}
	}

	wrappers := map[string]bool{}
	for _, r := range armTmpl.Resources {
		res, ok := r.(map[string]any)
		if !ok || res["resourceGroup"] == nil {
			continue
		}
		if name, ok := res["name"].(string); ok {
			wrappers[name] = true
		}
	}
	for _, name := range []string{identitiesDeploymentName, runnerGrantsDeploymentName, phoneHomeDeploymentName} {
		if !wrappers[name] {
			t.Errorf("%s missing or not targeted at a resource group", name)
		}
	}
}
