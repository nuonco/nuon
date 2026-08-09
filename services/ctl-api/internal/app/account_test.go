package app

import (
	"testing"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
)

func TestAccountAfterQueryTypeGrantsAreOrgScoped(t *testing.T) {
	acct := Account{
		Grants: []ResourceGrant{
			{
				OrgID:        "org_a",
				ResourceType: GrantResourceTypeInstall,
				ResourceID:   GrantResourceWildcard,
				Permission:   string(permissions.PermissionAll),
			},
			{
				OrgID:        "org_b",
				ResourceType: GrantResourceTypeWebhook,
				ResourceID:   GrantResourceWildcard,
				Permission:   string(permissions.PermissionRead),
			},
		},
	}

	if err := acct.AfterQuery(nil); err != nil {
		t.Fatal(err)
	}

	if got := acct.OrgTypeGrants("org_a")["install"]; got != permissions.PermissionAll {
		t.Fatalf("org_a install wildcard = %v, want all", got)
	}
	if got, ok := acct.OrgTypeGrants("org_b")["install"]; ok {
		t.Fatalf("org_a install wildcard leaked into org_b: %v", got)
	}
	if got := acct.OrgTypeGrants("org_b")["webhook"]; got != permissions.PermissionRead {
		t.Fatalf("org_b webhook wildcard = %v, want read", got)
	}
	if acct.OrgTypeGrants("org_c") != nil {
		t.Fatal("expected nil wildcard map for org with no grants")
	}

	if !acct.HasOrg("org_a") || !acct.HasOrg("org_b") {
		t.Fatal("wildcard grants should confer org membership")
	}
}

func TestAccountAfterQueryTypeGrantsPreferAll(t *testing.T) {
	acct := Account{
		Grants: []ResourceGrant{
			{
				OrgID:        "org_a",
				ResourceType: GrantResourceTypeInstall,
				ResourceID:   GrantResourceWildcard,
				Permission:   string(permissions.PermissionRead),
			},
			{
				OrgID:        "org_a",
				ResourceType: GrantResourceTypeInstall,
				ResourceID:   GrantResourceWildcard,
				Permission:   string(permissions.PermissionAll),
			},
		},
	}

	if err := acct.AfterQuery(nil); err != nil {
		t.Fatal(err)
	}

	if got := acct.OrgTypeGrants("org_a")["install"]; got != permissions.PermissionAll {
		t.Fatalf("install wildcard = %v, want all (read must not shadow all)", got)
	}
}
