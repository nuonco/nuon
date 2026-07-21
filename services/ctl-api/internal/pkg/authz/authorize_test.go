package authz

import (
	"testing"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
)

func TestAuthorize(t *testing.T) {
	const (
		installID = "inl_1"
		appID     = "app_1"
		orgID     = "org_1"
	)
	chain := []string{installID, appID, orgID}

	tests := []struct {
		name    string
		perms   permissions.Set
		perm    permissions.Permission
		allowed bool
	}{
		{
			name:    "grant on the resource itself",
			perms:   permissions.Set{installID: permissions.PermissionRead},
			perm:    permissions.PermissionRead,
			allowed: true,
		},
		{
			name:    "ancestor app grant cascades to install",
			perms:   permissions.Set{appID: permissions.PermissionRead},
			perm:    permissions.PermissionRead,
			allowed: true,
		},
		{
			name:    "org-wide all is the last link",
			perms:   permissions.Set{orgID: permissions.PermissionAll},
			perm:    permissions.PermissionUpdate,
			allowed: true,
		},
		{
			name:    "read grant does not satisfy a write",
			perms:   permissions.Set{installID: permissions.PermissionRead},
			perm:    permissions.PermissionUpdate,
			allowed: false,
		},
		{
			name:    "no grant anywhere in the chain",
			perms:   permissions.Set{"other": permissions.PermissionAll},
			perm:    permissions.PermissionRead,
			allowed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Authorize(tt.perms, chain, tt.perm)
			if got := err == nil; got != tt.allowed {
				t.Fatalf("Authorize allowed=%v, want %v (err=%v)", got, tt.allowed, err)
			}
		})
	}
}

func TestSetGrantNeverDowngrades(t *testing.T) {
	s := permissions.NewSet()

	permissions.Set(s).Grant("org_1", permissions.PermissionAll)
	permissions.Set(s).Grant("org_1", permissions.PermissionRead)
	if s["org_1"] != permissions.PermissionAll {
		t.Fatalf("read grant downgraded an all grant: got %v", s["org_1"])
	}

	permissions.Set(s).Grant("app_1", permissions.PermissionRead)
	permissions.Set(s).Grant("app_1", permissions.PermissionAll)
	if s["app_1"] != permissions.PermissionAll {
		t.Fatalf("all grant did not upgrade a read grant: got %v", s["app_1"])
	}
}
