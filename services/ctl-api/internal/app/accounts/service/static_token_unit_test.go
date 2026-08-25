package service

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestParseTokenDuration(t *testing.T) {
	testCases := []struct {
		name    string
		raw     string
		want    time.Duration
		wantErr bool
	}{
		{name: "empty defaults to one year", raw: "", want: 8760 * time.Hour},
		{name: "valid hours", raw: "720h", want: 720 * time.Hour},
		{name: "valid short", raw: "30m", want: 30 * time.Minute},
		{name: "zero is rejected", raw: "0h", wantErr: true},
		{name: "negative is rejected", raw: "-1h", wantErr: true},
		{name: "unparseable is rejected", raw: "bogus", wantErr: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseTokenDuration(tc.raw)
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

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

	s := &service{}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			acct := &app.Account{Roles: tc.roles}
			assert.Equal(t, tc.want, s.isOrgAdmin(acct, orgID))
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
			assert.Equal(t, tc.want, orgRoleType(acct, orgID))
		})
	}
}

// The guard on deleteTokenServiceAccount. Static tokens created with their own
// dedicated service account take it with them when revoked; tokens issued against
// an account that outlives them — an install stack's, a runner's — must not. Getting
// this wrong deletes a live identity out from under running infrastructure.
func TestDedicatedTokenSubjectPrefix(t *testing.T) {
	const orgID = "org_abc"
	prefix := dedicatedTokenSubjectPrefix(orgID)

	testCases := []struct {
		name          string
		subject       string
		wantDedicated bool
	}{
		{
			name:          "account created for one token",
			subject:       prefix + "acc123",
			wantDedicated: true,
		},
		{
			name:          "install stack account",
			subject:       "istmj4srnucyq26cwdj3nceerb",
			wantDedicated: false,
		},
		{
			name:          "another org's dedicated account",
			subject:       dedicatedTokenSubjectPrefix("org_other") + "acc123",
			wantDedicated: false,
		},
		{
			name:          "empty subject",
			subject:       "",
			wantDedicated: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.wantDedicated, strings.HasPrefix(tc.subject, prefix))
		})
	}
}
