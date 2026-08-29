package azure

import (
	"reflect"
	"testing"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

const contributorGUID = "b24988ac-6180-42a0-ab88-20f7382dd24c"

func TestExtractAzureStandardRolesRawSplitsGrantAxes(t *testing.T) {
	appCfg := &app.AppConfig{
		PermissionsConfig: app.AppPermissionsConfig{
			Roles: []app.AppAWSIAMRoleConfig{
				{
					CloudPlatform: "azure",
					Type:          app.AWSIAMRoleTypeRunnerProvision,
					Policies: []app.AppAWSIAMPolicyConfig{
						{Name: "networking", AzureActions: []string{"Microsoft.Network/virtualNetworks/write"}},
						{Name: "builtin", AzureBuiltInRoles: []string{"Contributor"}},
					},
				},
				{
					CloudPlatform: "azure",
					Type:          app.AWSIAMRoleTypeRunnerDeprovision,
					Policies: []app.AppAWSIAMPolicyConfig{
						{Name: "teardown", AzureActions: []string{"Microsoft.Resources/subscriptions/resourceGroups/delete"}},
					},
				},
			},
		},
	}

	prov, maint, deprov := ExtractAzureStandardRolesRaw(appCfg)

	if want := []string{"Microsoft.Network/virtualNetworks/write"}; !reflect.DeepEqual(prov.Actions, want) {
		t.Fatalf("provision actions = %#v, want %#v", prov.Actions, want)
	}
	// Built-in roles are resolved to GUIDs here, not forwarded by name: the
	// Terraform module builds a role definition ID from the value verbatim.
	if want := []string{contributorGUID}; !reflect.DeepEqual(prov.BuiltInRoles, want) {
		t.Fatalf("provision built-in roles = %#v, want %#v", prov.BuiltInRoles, want)
	}
	if want := []string{"Microsoft.Resources/subscriptions/resourceGroups/delete"}; !reflect.DeepEqual(deprov.Actions, want) {
		t.Fatalf("deprovision actions = %#v, want %#v", deprov.Actions, want)
	}
	// No maintenance role declared, so it stays zero — which is what makes the
	// module skip the identity and fall back to the runner's ambient one.
	if len(maint.Actions) != 0 || len(maint.BuiltInRoles) != 0 {
		t.Fatalf("maintenance = %#v, want empty", maint)
	}
}

// An unmapped built-in role passes through unchanged, so a literal GUID in the
// app config still works and a typo reaches Azure rather than being silently
// dropped here.
func TestExtractAzureStandardRolesRawPassesThroughUnmappedRole(t *testing.T) {
	appCfg := &app.AppConfig{
		PermissionsConfig: app.AppPermissionsConfig{
			Roles: []app.AppAWSIAMRoleConfig{
				{
					CloudPlatform: "azure",
					Type:          app.AWSIAMRoleTypeRunnerProvision,
					Policies: []app.AppAWSIAMPolicyConfig{
						{AzureBuiltInRoles: []string{"00000000-1111-2222-3333-444444444444"}},
					},
				},
			},
		},
	}

	prov, _, _ := ExtractAzureStandardRolesRaw(appCfg)

	if want := []string{"00000000-1111-2222-3333-444444444444"}; !reflect.DeepEqual(prov.BuiltInRoles, want) {
		t.Fatalf("provision built-in roles = %#v, want %#v", prov.BuiltInRoles, want)
	}
}

func TestExtractAzureStandardRolesRawSkipsOtherClouds(t *testing.T) {
	appCfg := &app.AppConfig{
		PermissionsConfig: app.AppPermissionsConfig{
			Roles: []app.AppAWSIAMRoleConfig{
				{
					CloudPlatform: "gcp",
					Type:          app.AWSIAMRoleTypeRunnerProvision,
					Policies: []app.AppAWSIAMPolicyConfig{
						{GCPPermissions: []string{"compute.networks.create"}},
					},
				},
				{
					CloudPlatform: "aws",
					Type:          app.AWSIAMRoleTypeRunnerProvision,
					Policies: []app.AppAWSIAMPolicyConfig{
						{Name: "ec2"},
					},
				},
			},
		},
	}

	prov, maint, deprov := ExtractAzureStandardRolesRaw(appCfg)

	if len(prov.Actions) != 0 || len(prov.BuiltInRoles) != 0 {
		t.Fatalf("provision = %#v, want empty for a non-azure app config", prov)
	}
	if len(maint.Actions) != 0 || len(deprov.Actions) != 0 {
		t.Fatal("maintenance/deprovision should be empty for a non-azure app config")
	}
}

func TestExtractAzureStandardRolesRawNilConfig(t *testing.T) {
	prov, maint, deprov := ExtractAzureStandardRolesRaw(nil)
	if len(prov.Actions) != 0 || len(maint.Actions) != 0 || len(deprov.Actions) != 0 {
		t.Fatal("nil app config should yield empty roles")
	}
}

func TestExtractAzureRolesRawKeepsNamesAndSkipsEmpty(t *testing.T) {
	roles := []app.AppAWSIAMRoleConfig{
		{
			CloudPlatform: "azure",
			Name:          "break-glass-1",
			Policies: []app.AppAWSIAMPolicyConfig{
				{AzureActions: []string{"Microsoft.Sql/servers/read"}},
				{AzureBuiltInRoles: []string{"Contributor"}},
			},
		},
		{
			// No azure grants at all: skipped rather than emitted as an identity
			// with no access.
			CloudPlatform: "azure",
			Name:          "empty-role",
			Policies:      []app.AppAWSIAMPolicyConfig{{Name: "nothing"}},
		},
		{
			CloudPlatform: "gcp",
			Name:          "gcp-role",
			Policies: []app.AppAWSIAMPolicyConfig{
				{GCPPermissions: []string{"cloudsql.instances.list"}},
			},
		},
	}

	out := ExtractAzureRolesRaw(roles)

	if len(out) != 1 {
		t.Fatalf("got %d roles, want 1: %#v", len(out), out)
	}
	if out[0].Name != "break-glass-1" {
		t.Fatalf("name = %q, want break-glass-1", out[0].Name)
	}
	if want := []string{"Microsoft.Sql/servers/read"}; !reflect.DeepEqual(out[0].Actions, want) {
		t.Fatalf("actions = %#v, want %#v", out[0].Actions, want)
	}
	if want := []string{contributorGUID}; !reflect.DeepEqual(out[0].BuiltInRoles, want) {
		t.Fatalf("built-in roles = %#v, want %#v", out[0].BuiltInRoles, want)
	}
}
