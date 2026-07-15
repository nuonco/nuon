package gcp

import (
	"reflect"
	"testing"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestExtractGCPStandardRolesRawPolicies(t *testing.T) {
	appCfg := &app.AppConfig{
		PermissionsConfig: app.AppPermissionsConfig{
			Roles: []app.AppAWSIAMRoleConfig{
				{
					CloudPlatform: "gcp",
					Type:          app.AWSIAMRoleTypeRunnerProvision,
					Policies: []app.AppAWSIAMPolicyConfig{
						{Name: "networking", GCPPermissions: []string{"compute.networks.create"}},
						{Name: "dns", GCPPermissions: []string{"dns.changes.create", "dns.managedZones.get"}},
						{Name: "empty"},
					},
				},
			},
		},
	}

	prov, _, _ := ExtractGCPStandardRolesRaw(appCfg)

	want := map[string][]string{
		"networking": {"compute.networks.create"},
		"dns":        {"dns.changes.create", "dns.managedZones.get"},
	}
	if !reflect.DeepEqual(prov.Policies, want) {
		t.Fatalf("provision policies = %#v, want %#v", prov.Policies, want)
	}

	// The flat permissions list stays populated alongside the per-policy map.
	wantPerms := []string{"compute.networks.create", "dns.changes.create", "dns.managedZones.get"}
	if !reflect.DeepEqual(prov.Permissions, wantPerms) {
		t.Fatalf("provision permissions = %#v, want %#v", prov.Permissions, wantPerms)
	}
}

func TestExtractGCPRolesRawUnnamedPolicyFallback(t *testing.T) {
	roles := []app.AppAWSIAMRoleConfig{
		{
			CloudPlatform: "gcp",
			Name:          "break-glass-1",
			Policies: []app.AppAWSIAMPolicyConfig{
				{GCPPermissions: []string{"cloudsql.instances.list"}},
			},
		},
	}

	got := ExtractGCPRolesRaw(roles)
	if len(got) != 1 {
		t.Fatalf("expected 1 role, got %d", len(got))
	}
	want := map[string][]string{"policy-0": {"cloudsql.instances.list"}}
	if !reflect.DeepEqual(got[0].Policies, want) {
		t.Fatalf("role policies = %#v, want %#v", got[0].Policies, want)
	}
}
