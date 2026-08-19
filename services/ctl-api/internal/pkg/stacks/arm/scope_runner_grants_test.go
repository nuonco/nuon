package arm

import (
	"slices"
	"testing"

	"github.com/nuonco/nuon/services/ctl-api/internal"
)

func runnerGrantsAt(t *testing.T, scope armScope, useOperationIdentities bool) []any {
	t.Helper()
	tmpl := &Templates{cfg: &internal.Config{}}
	out := &ARMTemplate{}
	tmpl.appendRunnerGrants(out, minimalTemplateInput(), scope, useOperationIdentities)
	return out.Resources
}

func assignmentNames(resources []any) []string {
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

func TestRunnerGrants_WrappedIntoInstallRG(t *testing.T) {
	resources := runnerGrantsAt(t, armScope{subscription: true}, false)

	wrapper, ok := resources[0].(map[string]any)
	if !ok || wrapper["name"] != runnerGrantsDeploymentName {
		t.Fatalf("expected %s first, got %v", runnerGrantsDeploymentName, resources[0])
	}
	if got := wrapper["resourceGroup"]; got != "[parameters('installResourceGroupName')]" {
		t.Errorf("wrapper resourceGroup = %v", got)
	}

	// The grants read the runner's identity and live in the install group, so the
	// wrapper has to wait on both.
	deps := wrapper["dependsOn"].([]string)
	if !slices.Contains(deps, "runnerDeployment") {
		t.Errorf("wrapper does not wait for the runner: %v", deps)
	}
	if !slices.Contains(deps, "[resourceId('Microsoft.Resources/resourceGroups', parameters('installResourceGroupName'))]") {
		t.Errorf("wrapper does not wait for the resource group: %v", deps)
	}

	props := wrapper["properties"].(map[string]any)
	innerDecl := props["template"].(map[string]any)["parameters"].(map[string]any)
	for _, name := range []string{"nuonInstallID", "principalId"} {
		if _, ok := innerDecl[name]; !ok {
			t.Errorf("inner template does not declare %q", name)
		}
	}

	// No assignment may be left loose in the root: a subscription-scoped root cannot
	// declare a resource-group-scoped role assignment.
	if got := assignmentNames(resources); len(got) != 0 {
		t.Errorf("role assignments still in the root: %v", got)
	}
}

// Inner expression evaluation hides the root, so reference('runnerDeployment')
// cannot be used inside the wrapper — the principal must arrive as a parameter and
// the per-assignment dependency must be gone, since the wrapper carries it.
func TestRunnerGrants_InnerAssignmentsReadPrincipalFromParameter(t *testing.T) {
	resources := runnerGrantsAt(t, armScope{subscription: true}, false)
	wrapper := resources[0].(map[string]any)

	inner := innerResources(t, wrapper)
	if len(inner) == 0 {
		t.Fatal("wrapper holds no grants")
	}

	for _, r := range inner {
		m := r.(map[string]any)
		name := m["name"].(string)

		if got := m["properties"].(map[string]any)["principalId"]; got != "[parameters('principalId')]" {
			t.Errorf("%s principalId = %v, want the wrapper parameter", name, got)
		}
		if _, ok := m["dependsOn"]; ok {
			t.Errorf("%s still depends on a root resource: %v", name, m["dependsOn"])
		}
	}
}

// Wrapping must not rename an assignment: Azure dedupes by principal+role+scope and
// a changed name fails redeploy with RoleAssignmentExists. Inside the wrapper
// resourceGroup() resolves to the install group again, so the guid() inputs match.
func TestRunnerGrants_AssignmentNamesUnchangedByWrapping(t *testing.T) {
	for _, useOperationIdentities := range []bool{false, true} {
		rgNames := assignmentNames(runnerGrantsAt(t, armScope{}, useOperationIdentities))

		sub := runnerGrantsAt(t, armScope{subscription: true}, useOperationIdentities)
		subNames := assignmentNames(innerResources(t, sub[0].(map[string]any)))

		if len(rgNames) == 0 {
			t.Fatalf("no assignments emitted (useOperationIdentities=%v)", useOperationIdentities)
		}
		if len(rgNames) != len(subNames) {
			t.Fatalf("assignment count changed (useOperationIdentities=%v): %d vs %d", useOperationIdentities, len(rgNames), len(subNames))
		}
		for i := range rgNames {
			if rgNames[i] != subNames[i] {
				t.Errorf("assignment %d changed:\n rg:  %s\n sub: %s", i, rgNames[i], subNames[i])
			}
		}
	}
}

// The custom role deployment targets the subscription, so it cannot go inside the
// wrapper — and it is only emitted on the legacy system-identity path.
func TestRunnerGrants_CustomRoleDeploymentStaysInRoot(t *testing.T) {
	legacy := runnerGrantsAt(t, armScope{subscription: true}, false)
	if got := countResourceType(legacy[1:], "Microsoft.Resources/deployments"); got != 1 {
		t.Errorf("expected the custom role deployment in the root, got %d", got)
	}

	perOperation := runnerGrantsAt(t, armScope{subscription: true}, true)
	if len(perOperation) != 1 {
		t.Errorf("per-operation identities should emit the wrapper only, got %d resources", len(perOperation))
	}
}

func TestRunnerGrants_LocalRunnersEmitNothing(t *testing.T) {
	tmpl := &Templates{cfg: &internal.Config{UseLocalRunners: true}}
	for _, scope := range []armScope{{}, {subscription: true}} {
		out := &ARMTemplate{}
		tmpl.appendRunnerGrants(out, minimalTemplateInput(), scope, false)
		if len(out.Resources) != 0 {
			t.Errorf("local runners should emit no grants, got %d", len(out.Resources))
		}
	}
}
