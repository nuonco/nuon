package arm

import (
	"strings"
	"testing"

	"github.com/nuonco/nuon/services/ctl-api/internal"
)

func identityFixture() []azureOperationIdentity {
	return azureOperationIdentities(azureRolesTemplateInput().AppCfg)
}

func innerResources(t *testing.T, wrapper map[string]any) []any {
	t.Helper()
	props, ok := wrapper["properties"].(map[string]any)
	if !ok {
		t.Fatalf("wrapper has no properties: %v", wrapper)
	}
	tmpl, ok := props["template"].(map[string]any)
	if !ok {
		t.Fatalf("wrapper has no inline template: %v", props)
	}
	res, ok := tmpl["resources"].([]any)
	if !ok {
		t.Fatalf("inline template has no resources: %v", tmpl)
	}
	return res
}

func TestOperationIdentities_WrappedIntoInstallRG(t *testing.T) {
	tmpl := &Templates{cfg: &internal.Config{}}
	ids := identityFixture()

	resources := tmpl.getOperationIdentityResources(ids, armScope{subscription: true})

	wrapper, ok := resources[0].(map[string]any)
	if !ok || wrapper["name"] != identitiesDeploymentName {
		t.Fatalf("expected %s first, got %v", identitiesDeploymentName, resources[0])
	}
	if got := wrapper["resourceGroup"]; got != "[variables('installResourceGroupName')]" {
		t.Errorf("wrapper resourceGroup = %v", got)
	}

	props := wrapper["properties"].(map[string]any)
	if got := props["expressionEvaluationOptions"].(map[string]any)["scope"]; got != "inner" {
		t.Errorf("wrapper must use inner evaluation, got %v", got)
	}

	innerDecl := props["template"].(map[string]any)["parameters"].(map[string]any)
	for _, name := range []string{"nuonInstallID", "location", "commonTags"} {
		if _, ok := innerDecl[name]; !ok {
			t.Errorf("inner template does not declare %q", name)
		}
	}

	inner := innerResources(t, wrapper)
	if got := countResourceType(inner, uamiResourceType); got != len(ids) {
		t.Errorf("wrapper holds %d identities, want %d", got, len(ids))
	}
	if got := countResourceType(inner, "Microsoft.Authorization/roleAssignments"); got == 0 {
		t.Error("wrapper holds no built-in role assignments")
	}

	root := resources[1:]
	if got := countResourceType(root, "Microsoft.Resources/deployments"); got != len(ids) {
		t.Errorf("root holds %d custom role deployments, want %d", got, len(ids))
	}
	if got := countResourceType(root, uamiResourceType); got != 0 {
		t.Errorf("root still holds %d identities directly", got)
	}
}

func TestOperationIdentities_RoleAssignmentNamesUnchangedByWrapping(t *testing.T) {
	tmpl := &Templates{cfg: &internal.Config{}}
	ids := identityFixture()

	names := func(resources []any) []string {
		var out []string
		for _, r := range resources {
			m, ok := r.(map[string]any)
			if !ok || m["type"] != "Microsoft.Authorization/roleAssignments" {
				continue
			}
			out = append(out, m["name"].(string))
		}
		return out
	}

	rgNames := names(tmpl.getOperationIdentityResources(ids, armScope{}))

	subResources := tmpl.getOperationIdentityResources(ids, armScope{subscription: true})
	subNames := names(innerResources(t, subResources[0].(map[string]any)))

	if len(rgNames) == 0 {
		t.Fatal("fixture produced no built-in role assignments")
	}
	if len(rgNames) != len(subNames) {
		t.Fatalf("assignment count changed: %d at resource-group scope, %d at subscription scope", len(rgNames), len(subNames))
	}
	for i := range rgNames {
		if rgNames[i] != subNames[i] {
			t.Errorf("assignment name %d changed:\n rg:  %s\n sub: %s", i, rgNames[i], subNames[i])
		}
	}
}

func TestOperationIdentities_RootReadsIdentitiesAcrossTheWrapper(t *testing.T) {
	tmpl := &Templates{cfg: &internal.Config{}}
	id := identityFixture()[0]

	role := tmpl.getOperationIdentityCustomRole(id, armScope{subscription: true})

	deps := role["dependsOn"].([]string)
	if len(deps) != 1 || deps[0] != identitiesDeploymentName {
		t.Errorf("custom role must depend on %s, got %v", identitiesDeploymentName, deps)
	}

	principal := role["properties"].(map[string]any)["parameters"].(map[string]any)["principalID"].(map[string]any)["value"].(string)
	want := "[reference('identitiesDeployment').outputs.provisionPrincipalId.value]"
	if principal != want {
		t.Errorf("principalID:\n got: %s\nwant: %s", principal, want)
	}
	assertNoNestedBrackets(t, []byte(principal))

	wrapper := tmpl.getOperationIdentityResources(identityFixture(), armScope{subscription: true})[0].(map[string]any)
	outputs, ok := wrapper["properties"].(map[string]any)["template"].(map[string]any)["outputs"].(map[string]any)
	if !ok {
		t.Fatal("identitiesDeployment declares no outputs")
	}
	if _, ok := outputs["provisionPrincipalId"]; !ok {
		t.Errorf("identitiesDeployment does not export provisionPrincipalId, got %v", sortedKeys(outputs))
	}
}

func TestOperationIdentities_EveryRootReadHasAMatchingOutput(t *testing.T) {
	tmpl := &Templates{cfg: &internal.Config{}}
	ids := identityFixture()
	sub := armScope{subscription: true}

	wrapper := tmpl.getOperationIdentityResources(ids, sub)[0].(map[string]any)
	outputs, _ := wrapper["properties"].(map[string]any)["template"].(map[string]any)["outputs"].(map[string]any)

	for _, id := range ids {
		for _, expr := range []string{identityPrincipalIDExpr(id, sub), identityClientIDExpr(id, sub)} {
			key := strings.TrimSuffix(strings.TrimPrefix(expr, "[reference('identitiesDeployment').outputs."), ".value]")
			if _, ok := outputs[key]; !ok {
				t.Errorf("root reads %s but the wrapper exports only %v", expr, sortedKeys(outputs))
			}
		}
	}

	for _, id := range ids {
		if got := identityPrincipalIDExpr(id, armScope{}); strings.Contains(got, "outputs.") {
			t.Errorf("resource-group scope should read the identity directly, got %s", got)
		}
	}
}

func TestOperationIdentityAttachment_ScopeOfTheReader(t *testing.T) {
	ids := identityFixture()

	innerMap, _ := operationIdentityAttachment(ids, armScope{})
	if _, ok := innerMap[uamiResourceIDExpr(ids[0].suffix, armScope{})]; !ok {
		t.Errorf("inner attachment map is not keyed at resource-group scope: %v", innerMap)
	}

	rootMap, _ := operationIdentityAttachment(ids, armScope{subscription: true})
	if _, ok := rootMap[uamiResourceIDExpr(ids[0].suffix, armScope{subscription: true})]; !ok {
		t.Errorf("root attachment map is not fully qualified: %v", rootMap)
	}
}

func TestOperationIdentityAttachment_RootDependsOnWrapperOnce(t *testing.T) {
	ids := identityFixture()
	if len(ids) < 2 {
		t.Fatalf("fixture needs multiple identities, got %d", len(ids))
	}

	_, deps := operationIdentityAttachment(ids, armScope{subscription: true})
	if len(deps) != 1 || deps[0] != identitiesDeploymentName {
		t.Errorf("root should depend on %s exactly once, got %v", identitiesDeploymentName, deps)
	}

	_, rgDeps := operationIdentityAttachment(ids, armScope{})
	if len(rgDeps) != len(ids) {
		t.Errorf("resource-group scope should depend on each identity: got %d, want %d", len(rgDeps), len(ids))
	}
}
