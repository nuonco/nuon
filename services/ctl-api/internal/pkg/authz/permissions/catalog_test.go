package permissions_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	p "github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
)

func TestCustomerRole(t *testing.T) {
	tests := map[string]struct {
		name p.CustomerRoleName
		kind p.ResourceKind
		want string
	}{
		"install reader": {p.RoleInstallReader, p.KindInstall, "install.read log_stream.read runner.read runner_job.read telemetry.read"},
		"install editor": {p.RoleInstallEditor, p.KindInstall, "install.read install.write log_stream.read runner.read runner_job.read telemetry.read"},
		"install admin":  {p.RoleInstallAdmin, p.KindInstall, "install.access.manage install.read install.write log_stream.read runner.read runner_job.read telemetry.read"},
		"app reader":     {p.RoleAppReader, p.KindApp, "app.read install.read log_stream.read runner.read runner_job.read telemetry.read"},
		"app editor":     {p.RoleAppEditor, p.KindApp, "app.read app.write install.read install.write log_stream.read runner.read runner_job.read telemetry.read"},
		"app admin":      {p.RoleAppAdmin, p.KindApp, "app.access.manage app.read app.write install.access.manage install.read install.write log_stream.read runner.read runner_job.read telemetry.read"},
		"org reader":     {p.RoleOrgReader, p.KindOrg, "app.read install.read log_stream.read org.read runner.read runner_job.read telemetry.read"},
		"org editor":     {p.RoleOrgEditor, p.KindOrg, "app.read app.write install.read install.write log_stream.read org.read runner.read runner_job.read telemetry.read"},
		"org admin":      {p.RoleOrgAdmin, p.KindOrg, "app.access.manage app.read app.write install.access.manage install.read install.write log_stream.read org.read org.write runner.read runner_job.read telemetry.read"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			role, err := p.CurrentCustomerRole(tt.name)
			require.NoError(t, err)
			assert.Equal(t, tt.name, role.Name)
			assert.Equal(t, 1, role.Revision)
			assert.Equal(t, tt.kind, role.ScopeKind)
			var got []string
			for _, permission := range role.Permissions.Values() {
				got = append(got, string(permission))
			}
			assert.Equal(t, strings.Fields(tt.want), got)
			delete(role.Permissions, p.InstallRead)
			role.Permissions[p.RunnerWrite] = struct{}{}
			again, err := p.CurrentCustomerRole(tt.name)
			require.NoError(t, err)
			assert.Contains(t, again.Permissions, p.InstallRead)
			assert.NotContains(t, again.Permissions, p.RunnerWrite)
		})
	}
}

func TestBindCustomerRole(t *testing.T) {
	tests := map[string]struct {
		role    p.CustomerRoleName
		scope   p.Scope
		wantErr bool
	}{
		"org":                 {p.RoleOrgReader, p.Scope{Kind: p.KindOrg, OrgID: "org_acme", ID: "org_acme"}, false},
		"app":                 {p.RoleAppEditor, p.Scope{Kind: p.KindApp, OrgID: "org_acme", ID: "app_one"}, false},
		"install":             {p.RoleInstallReader, p.Scope{Kind: p.KindInstall, OrgID: "org_acme", ID: "install_one"}, false},
		"install role at org": {p.RoleInstallReader, p.Scope{Kind: p.KindOrg, OrgID: "org_acme", ID: "org_acme"}, true},
		"org role at app":     {p.RoleOrgAdmin, p.Scope{Kind: p.KindApp, OrgID: "org_acme", ID: "app_one"}, true},
		"missing org":         {p.RoleInstallReader, p.Scope{Kind: p.KindInstall, ID: "install_one"}, true},
		"wildcard target":     {p.RoleAppReader, p.Scope{Kind: p.KindApp, OrgID: "org_acme", ID: "*"}, true},
		"mismatched org":      {p.RoleOrgAdmin, p.Scope{Kind: p.KindOrg, OrgID: "org_acme", ID: "org_other"}, true},
		"unknown role":        {"unknown", p.Scope{Kind: p.KindOrg, OrgID: "org_acme", ID: "org_acme"}, true},
		"legacy support":      {"org_support", p.Scope{Kind: p.KindOrg, OrgID: "org_acme", ID: "org_acme"}, true},
		"managed runner":      {"runner", p.Scope{Kind: p.KindOrg, OrgID: "org_acme", ID: "org_acme"}, true},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			grant, err := p.BindCurrentCustomerRole(tt.role, tt.scope)
			if tt.wantErr {
				require.Error(t, err)
				assert.Empty(t, grant.Permissions)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.scope, grant.Scope)
			assert.Contains(t, grant.Permissions, p.InstallRead)
		})
	}
}

func TestPermissionSet(t *testing.T) {
	tests := map[string]struct {
		values  []p.PermissionID
		want    []p.PermissionID
		wantErr bool
	}{
		"empty":                     {nil, []p.PermissionID{}, false},
		"sorted finite union":       {[]p.PermissionID{p.InstallWrite, p.AppRead, p.InstallRead, p.InstallWrite}, []p.PermissionID{p.AppRead, p.InstallRead, p.InstallWrite}, false},
		"distinct compound actions": {[]p.PermissionID{p.TelemetryRead, p.TelemetryConfigRead, p.TelemetryIssue}, []p.PermissionID{p.TelemetryConfigRead, p.TelemetryIssue, p.TelemetryRead}, false},
		"unknown after valid":       {[]p.PermissionID{p.InstallRead, "install.future"}, nil, true},
		"legacy all":                {[]p.PermissionID{"all"}, nil, true},
		"legacy read":               {[]p.PermissionID{"read"}, nil, true},
		"wildcard":                  {[]p.PermissionID{"install.*"}, nil, true},
		"empty permission":          {[]p.PermissionID{""}, nil, true},
		"api metadata":              {[]p.PermissionID{"ctl-api"}, nil, true},
		"whitespace":                {[]p.PermissionID{"install.read "}, nil, true},
		"case":                      {[]p.PermissionID{"Install.read"}, nil, true},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			set, err := p.NewPermissionSet(tt.values...)
			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, set)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, set.Values())
		})
	}
}
