package helpers

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestIsOrgAdmin(t *testing.T) {
	const orgID = "org_abc"

	adminRole := app.Role{OrgID: generics.NewNullString(orgID), RoleType: app.RoleTypeOrgAdmin}
	otherOrgAdmin := app.Role{OrgID: generics.NewNullString("org_other"), RoleType: app.RoleTypeOrgAdmin}
	sameOrgInstaller := app.Role{OrgID: generics.NewNullString(orgID), RoleType: app.RoleTypeInstaller}

	testCases := []struct {
		name  string
		roles []app.Role
		want  bool
	}{
		{name: "admin for the org", roles: []app.Role{adminRole}, want: true},
		{name: "admin only for a different org", roles: []app.Role{otherOrgAdmin}, want: false},
		{name: "non-admin role for the org", roles: []app.Role{sameOrgInstaller}, want: false},
		{name: "no roles", roles: nil, want: false},
		{name: "admin among several roles", roles: []app.Role{sameOrgInstaller, otherOrgAdmin, adminRole}, want: true},
	}

	h := &Helpers{}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			acct := &app.Account{Roles: tc.roles}
			assert.Equal(t, tc.want, h.IsOrgAdmin(acct, orgID))
		})
	}
}

func TestOrgRoleType(t *testing.T) {
	const orgID = "org_abc"

	testCases := []struct {
		name  string
		roles []app.Role
		want  app.RoleType
	}{
		{
			name:  "role in the org",
			roles: []app.Role{{OrgID: generics.NewNullString(orgID), RoleType: app.RoleTypeOrgAdmin}},
			want:  app.RoleTypeOrgAdmin,
		},
		{
			name:  "role only in another org is not borrowed",
			roles: []app.Role{{OrgID: generics.NewNullString("org_other"), RoleType: app.RoleTypeOrgAdmin}},
			want:  "",
		},
		{
			name: "picks the role scoped to this org",
			roles: []app.Role{
				{OrgID: generics.NewNullString("org_other"), RoleType: app.RoleTypeOrgAdmin},
				{OrgID: generics.NewNullString(orgID), RoleType: app.RoleTypeOrgReadOnly},
			},
			want: app.RoleTypeOrgReadOnly,
		},
		{name: "no roles", roles: nil, want: ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			acct := &app.Account{Roles: tc.roles}
			assert.Equal(t, tc.want, OrgRoleType(acct, orgID))
		})
	}
}
