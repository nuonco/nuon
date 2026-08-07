package service

import (
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

func TestParseTokenRole(t *testing.T) {
	testCases := []struct {
		name    string
		raw     string
		want    app.RoleType
		wantErr bool
	}{
		{name: "empty defaults to read-only", raw: "", want: app.RoleTypeOrgReadOnly},
		{name: "admin", raw: "org_admin", want: app.RoleTypeOrgAdmin},
		{name: "support", raw: "org_support", want: app.RoleTypeOrgSupport},
		{name: "read-only", raw: "org_read_only", want: app.RoleTypeOrgReadOnly},
		{name: "deprecated builder is rejected", raw: "org_builder", wantErr: true},
		{name: "internal role is rejected", raw: "installer", wantErr: true},
		{name: "unknown role is rejected", raw: "superuser", wantErr: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseTokenRole(tc.raw)
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
