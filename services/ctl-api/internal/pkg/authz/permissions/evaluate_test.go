package permissions_test

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	p "github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
)

func TestEvaluate(t *testing.T) {
	org := p.Scope{Kind: p.KindOrg, OrgID: "org_acme", ID: "org_acme"}
	app := p.Scope{Kind: p.KindApp, OrgID: "org_acme", ID: "app_one"}
	install := p.Scope{Kind: p.KindInstall, OrgID: "org_acme", ID: "install_one"}
	otherInstall := p.Scope{Kind: p.KindInstall, OrgID: "org_acme", ID: "install_two"}
	installTarget := p.ResolvedTarget{Scope: install, AppID: app.ID, InstallID: install.ID}
	otherTarget := p.ResolvedTarget{Scope: otherInstall, AppID: app.ID, InstallID: otherInstall.ID}
	appTarget := p.ResolvedTarget{Scope: app, AppID: app.ID}
	stream := p.Scope{Kind: p.KindLogStream, OrgID: org.ID, ID: "stream_one"}
	logTarget := p.ResolvedTarget{Scope: stream, AppID: app.ID, InstallID: install.ID}
	buildLog := p.ResolvedTarget{Scope: p.Scope{Kind: p.KindLogStream, OrgID: org.ID, ID: "build_log"}, AppID: app.ID}
	telemetry := p.ResolvedTarget{Scope: p.Scope{Kind: p.KindTelemetry, OrgID: org.ID, ID: "telemetry_one"}, AppID: app.ID, InstallID: install.ID}
	job := p.ResolvedTarget{Scope: p.Scope{Kind: p.KindRunnerJob, OrgID: org.ID, ID: "job_one"}, AppID: app.ID, InstallID: install.ID}
	runner := p.ResolvedTarget{Scope: p.Scope{Kind: p.KindRunner, OrgID: org.ID, ID: "runner_one"}, AppID: app.ID, InstallID: install.ID}

	grant := func(scope p.Scope, values ...p.PermissionID) p.ScopedPermissions {
		set, err := p.NewPermissionSet(values...)
		require.NoError(t, err)
		return p.ScopedPermissions{Scope: scope, Permissions: set}
	}
	role := func(name p.CustomerRoleName, scope p.Scope) p.ScopedPermissions {
		g, err := p.BindCurrentCustomerRole(name, scope)
		require.NoError(t, err)
		return g
	}
	orgAdmin := role(p.RoleOrgAdmin, org)
	appReader := role(p.RoleAppReader, app)
	installReader := role(p.RoleInstallReader, install)
	installEditor := role(p.RoleInstallEditor, install)

	tests := map[string]struct {
		permission p.PermissionID
		target     p.ResolvedTarget
		grants     []p.ScopedPermissions
		ceiling    []p.ScopedPermissions
		want       p.PermissionMatch
	}{
		"org grant with install ceiling":              {p.InstallRead, installTarget, []p.ScopedPermissions{orgAdmin}, []p.ScopedPermissions{installReader}, p.PermissionMatch{Matched: true, Reason: p.ReasonMatched}},
		"ceiling rejects sibling":                     {p.InstallRead, otherTarget, []p.ScopedPermissions{orgAdmin}, []p.ScopedPermissions{installReader}, p.PermissionMatch{Reason: p.ReasonCeilingExceeded}},
		"ceiling cannot add principal permission":     {p.InstallWrite, installTarget, []p.ScopedPermissions{installReader}, []p.ScopedPermissions{orgAdmin}, p.PermissionMatch{Reason: p.ReasonNoGrant}},
		"read ceiling rejects admin write":            {p.InstallWrite, installTarget, []p.ScopedPermissions{orgAdmin}, []p.ScopedPermissions{installReader}, p.PermissionMatch{Reason: p.ReasonCeilingExceeded}},
		"missing ceiling":                             {p.InstallRead, installTarget, []p.ScopedPermissions{orgAdmin}, nil, p.PermissionMatch{Reason: p.ReasonCeilingExceeded}},
		"missing grants":                              {p.InstallRead, installTarget, nil, []p.ScopedPermissions{orgAdmin}, p.PermissionMatch{Reason: p.ReasonNoGrant}},
		"empty set":                                   {p.InstallRead, installTarget, []p.ScopedPermissions{{Scope: org}}, []p.ScopedPermissions{orgAdmin}, p.PermissionMatch{Reason: p.ReasonNoGrant}},
		"distinct grants union":                       {p.InstallWrite, installTarget, []p.ScopedPermissions{grant(install, p.InstallWrite), grant(install, p.InstallRead)}, []p.ScopedPermissions{orgAdmin}, p.PermissionMatch{Matched: true, Reason: p.ReasonMatched}},
		"write does not imply read":                   {p.InstallRead, installTarget, []p.ScopedPermissions{grant(install, p.InstallWrite)}, []p.ScopedPermissions{orgAdmin}, p.PermissionMatch{Reason: p.ReasonNoGrant}},
		"mixed roles permit exact write":              {p.InstallWrite, installTarget, []p.ScopedPermissions{appReader, installEditor}, []p.ScopedPermissions{orgAdmin}, p.PermissionMatch{Matched: true, Reason: p.ReasonMatched}},
		"mixed roles retain sibling read":             {p.InstallRead, otherTarget, []p.ScopedPermissions{appReader, installEditor}, []p.ScopedPermissions{orgAdmin}, p.PermissionMatch{Matched: true, Reason: p.ReasonMatched}},
		"mixed roles cannot spread write":             {p.InstallWrite, otherTarget, []p.ScopedPermissions{appReader, installEditor}, []p.ScopedPermissions{orgAdmin}, p.PermissionMatch{Reason: p.ReasonNoGrant}},
		"mixed ceilings cannot spread write":          {p.InstallWrite, otherTarget, []p.ScopedPermissions{orgAdmin}, []p.ScopedPermissions{appReader, installEditor}, p.PermissionMatch{Reason: p.ReasonCeilingExceeded}},
		"app reader sees build logs":                  {p.LogStreamRead, buildLog, []p.ScopedPermissions{appReader}, []p.ScopedPermissions{orgAdmin}, p.PermissionMatch{Matched: true, Reason: p.ReasonMatched}},
		"install reader cannot see parent build logs": {p.LogStreamRead, buildLog, []p.ScopedPermissions{installReader}, []p.ScopedPermissions{orgAdmin}, p.PermissionMatch{Reason: p.ReasonNoGrant}},
		"install reader sees own logs":                {p.LogStreamRead, logTarget, []p.ScopedPermissions{installReader}, []p.ScopedPermissions{orgAdmin}, p.PermissionMatch{Matched: true, Reason: p.ReasonMatched}},
		"bare install read cannot query logs":         {p.LogStreamRead, logTarget, []p.ScopedPermissions{grant(install, p.InstallRead)}, []p.ScopedPermissions{orgAdmin}, p.PermissionMatch{Reason: p.ReasonNoGrant}},
		"bare install read cannot query telemetry":    {p.TelemetryRead, telemetry, []p.ScopedPermissions{grant(install, p.InstallRead)}, []p.ScopedPermissions{orgAdmin}, p.PermissionMatch{Reason: p.ReasonNoGrant}},
		"reader sees telemetry":                       {p.TelemetryRead, telemetry, []p.ScopedPermissions{installReader}, []p.ScopedPermissions{orgAdmin}, p.PermissionMatch{Matched: true, Reason: p.ReasonMatched}},
		"reader sees job history":                     {p.RunnerJobRead, job, []p.ScopedPermissions{installReader}, []p.ScopedPermissions{orgAdmin}, p.PermissionMatch{Matched: true, Reason: p.ReasonMatched}},
		"reader sees runner status":                   {p.RunnerRead, runner, []p.ScopedPermissions{installReader}, []p.ScopedPermissions{orgAdmin}, p.PermissionMatch{Matched: true, Reason: p.ReasonMatched}},
		"reader cannot fetch collector config":        {p.TelemetryConfigRead, telemetry, []p.ScopedPermissions{installReader}, []p.ScopedPermissions{grant(install, p.TelemetryConfigRead)}, p.PermissionMatch{Reason: p.ReasonNoGrant}},
		"reader cannot fetch execution config":        {p.RunnerJobConfigRead, job, []p.ScopedPermissions{installReader}, []p.ScopedPermissions{grant(install, p.RunnerJobConfigRead)}, p.PermissionMatch{Reason: p.ReasonNoGrant}},
		"editor cannot report jobs":                   {p.RunnerJobWrite, job, []p.ScopedPermissions{installEditor}, []p.ScopedPermissions{grant(install, p.RunnerJobWrite)}, p.PermissionMatch{Reason: p.ReasonNoGrant}},
		"admin cannot claim jobs":                     {p.RunnerJobClaim, job, []p.ScopedPermissions{orgAdmin}, []p.ScopedPermissions{grant(install, p.RunnerJobClaim)}, p.PermissionMatch{Reason: p.ReasonNoGrant}},
		"exact stream writer":                         {p.LogStreamWrite, logTarget, []p.ScopedPermissions{grant(stream, p.LogStreamWrite)}, []p.ScopedPermissions{grant(install, p.LogStreamWrite)}, p.PermissionMatch{Matched: true, Reason: p.ReasonMatched}},
		"exact stream writer cannot use other stream": {p.LogStreamWrite, buildLog, []p.ScopedPermissions{grant(stream, p.LogStreamWrite)}, []p.ScopedPermissions{grant(org, p.LogStreamWrite)}, p.PermissionMatch{Reason: p.ReasonNoGrant}},
		"matching ID with wrong resource kind":        {p.LogStreamRead, logTarget, []p.ScopedPermissions{grant(p.Scope{Kind: p.KindRunnerJob, OrgID: org.ID, ID: stream.ID}, p.LogStreamRead)}, []p.ScopedPermissions{orgAdmin}, p.PermissionMatch{Reason: p.ReasonNoGrant}},
		"other app in same org":                       {p.LogStreamRead, p.ResolvedTarget{Scope: stream, AppID: "app_two"}, []p.ScopedPermissions{appReader}, []p.ScopedPermissions{orgAdmin}, p.PermissionMatch{Reason: p.ReasonNoGrant}},
		"same resource ID in different tenant":        {p.LogStreamRead, p.ResolvedTarget{Scope: p.Scope{Kind: p.KindLogStream, OrgID: "org_other", ID: stream.ID}, AppID: app.ID, InstallID: install.ID}, []p.ScopedPermissions{installReader}, []p.ScopedPermissions{orgAdmin}, p.PermissionMatch{Reason: p.ReasonNoGrant}},
		"install creation checks parent app":          {p.InstallWrite, appTarget, []p.ScopedPermissions{role(p.RoleAppEditor, app)}, []p.ScopedPermissions{orgAdmin}, p.PermissionMatch{Matched: true, Reason: p.ReasonMatched}},
		"install editor cannot create sibling":        {p.InstallWrite, appTarget, []p.ScopedPermissions{installEditor}, []p.ScopedPermissions{orgAdmin}, p.PermissionMatch{Reason: p.ReasonNoGrant}},
		"app creation checks parent org":              {p.AppWrite, p.ResolvedTarget{Scope: org}, []p.ScopedPermissions{role(p.RoleOrgEditor, org)}, []p.ScopedPermissions{orgAdmin}, p.PermissionMatch{Matched: true, Reason: p.ReasonMatched}},
		"app editor cannot create sibling":            {p.AppWrite, p.ResolvedTarget{Scope: org}, []p.ScopedPermissions{role(p.RoleAppEditor, app)}, []p.ScopedPermissions{orgAdmin}, p.PermissionMatch{Reason: p.ReasonNoGrant}},
		"install admin access management":             {p.InstallAccessManage, installTarget, []p.ScopedPermissions{role(p.RoleInstallAdmin, install)}, []p.ScopedPermissions{orgAdmin}, p.PermissionMatch{Matched: true, Reason: p.ReasonMatched}},
		"editor has no access management":             {p.InstallAccessManage, installTarget, []p.ScopedPermissions{installEditor}, []p.ScopedPermissions{orgAdmin}, p.PermissionMatch{Reason: p.ReasonNoGrant}},
		"read ceiling restricts admin management":     {p.InstallAccessManage, installTarget, []p.ScopedPermissions{orgAdmin}, []p.ScopedPermissions{installReader}, p.PermissionMatch{Reason: p.ReasonCeilingExceeded}},
		"invalid permission":                          {"install.*", installTarget, []p.ScopedPermissions{orgAdmin}, []p.ScopedPermissions{orgAdmin}, p.PermissionMatch{Reason: p.ReasonInvalidPermission}},
		"permission target mismatch":                  {p.OrgRead, installTarget, []p.ScopedPermissions{orgAdmin}, []p.ScopedPermissions{orgAdmin}, p.PermissionMatch{Reason: p.ReasonInvalidTarget}},
		"missing parent":                              {p.InstallRead, p.ResolvedTarget{Scope: install, InstallID: install.ID}, []p.ScopedPermissions{orgAdmin}, []p.ScopedPermissions{orgAdmin}, p.PermissionMatch{Reason: p.ReasonInvalidTarget}},
		"contradictory install identity":              {p.InstallRead, p.ResolvedTarget{Scope: install, AppID: app.ID, InstallID: otherInstall.ID}, []p.ScopedPermissions{orgAdmin}, []p.ScopedPermissions{orgAdmin}, p.PermissionMatch{Reason: p.ReasonInvalidTarget}},
		"org cannot forge child ancestry":             {p.AppWrite, p.ResolvedTarget{Scope: org, AppID: app.ID}, []p.ScopedPermissions{role(p.RoleAppEditor, app)}, []p.ScopedPermissions{orgAdmin}, p.PermissionMatch{Reason: p.ReasonInvalidTarget}},
		"app cannot forge install ancestry":           {p.InstallWrite, p.ResolvedTarget{Scope: app, AppID: app.ID, InstallID: install.ID}, []p.ScopedPermissions{installEditor}, []p.ScopedPermissions{orgAdmin}, p.PermissionMatch{Reason: p.ReasonInvalidTarget}},
		"missing target":                              {p.InstallRead, p.ResolvedTarget{}, []p.ScopedPermissions{orgAdmin}, []p.ScopedPermissions{orgAdmin}, p.PermissionMatch{Reason: p.ReasonInvalidTarget}},
		"unknown target kind":                         {p.InstallRead, p.ResolvedTarget{Scope: p.Scope{Kind: "unknown", OrgID: org.ID, ID: install.ID}}, []p.ScopedPermissions{orgAdmin}, []p.ScopedPermissions{orgAdmin}, p.PermissionMatch{Reason: p.ReasonInvalidTarget}},
		"invalid grant does not hide behind allow":    {p.InstallRead, installTarget, []p.ScopedPermissions{orgAdmin, {Scope: org, Permissions: p.PermissionSet{"unknown": {}}}}, []p.ScopedPermissions{orgAdmin}, p.PermissionMatch{Reason: p.ReasonInvalidGrant}},
		"wildcard grant":                              {p.InstallRead, installTarget, []p.ScopedPermissions{orgAdmin, {Scope: p.Scope{Kind: p.KindOrg, OrgID: "*", ID: "*"}}}, []p.ScopedPermissions{orgAdmin}, p.PermissionMatch{Reason: p.ReasonInvalidGrant}},
		"invalid ceiling does not hide behind allow":  {p.InstallRead, installTarget, []p.ScopedPermissions{orgAdmin}, []p.ScopedPermissions{orgAdmin, {Scope: org, Permissions: p.PermissionSet{"read": {}}}}, p.PermissionMatch{Reason: p.ReasonInvalidCeiling}},
		"missing ceiling scope":                       {p.InstallRead, installTarget, []p.ScopedPermissions{orgAdmin}, []p.ScopedPermissions{{Permissions: orgAdmin.Permissions}}, p.PermissionMatch{Reason: p.ReasonInvalidCeiling}},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			input := p.PermissionCheck{Permission: tt.permission, Target: tt.target, Grants: tt.grants, Ceiling: tt.ceiling}
			assert.Equal(t, tt.want, p.MatchPermissions(input))
			input.Grants = slices.Clone(tt.grants)
			input.Ceiling = slices.Clone(tt.ceiling)
			slices.Reverse(input.Grants)
			slices.Reverse(input.Ceiling)
			assert.Equal(t, tt.want, p.MatchPermissions(input))
		})
	}
}

func TestEvaluateWorkloadPermissions(t *testing.T) {
	tests := map[string]struct {
		kind        p.ResourceKind
		permissions []p.PermissionID
	}{
		"runner":       {p.KindRunner, []p.PermissionID{p.RunnerConfigRead, p.RunnerWrite}},
		"job":          {p.KindRunnerJob, []p.PermissionID{p.RunnerJobConfigRead, p.RunnerJobClaim, p.RunnerJobWrite}},
		"telemetry":    {p.KindTelemetry, []p.PermissionID{p.TelemetryConfigRead, p.TelemetryIssue, p.TelemetryWrite}},
		"stack":        {p.KindStack, []p.PermissionID{p.StackRead, p.StackReport}},
		"phone home":   {p.KindPhoneHome, []p.PermissionID{p.StackReport}},
		"state":        {p.KindExecutionState, []p.PermissionID{p.ExecutionStateRead, p.ExecutionStateWrite}},
		"ingress":      {p.KindEvent, []p.PermissionID{p.EventSubmit}},
		"notification": {p.KindNotification, []p.PermissionID{p.NotificationRead, p.NotificationWrite, p.NotificationDeliver}},
		"operation":    {p.KindOperation, []p.PermissionID{p.OperationExecute}},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			for _, permission := range tt.permissions {
				parsed, err := p.ParsePermissionID(string(permission))
				require.NoError(t, err)
				assert.Equal(t, permission, parsed)
				_, err = p.NewPermission(string(permission))
				require.Error(t, err)
				scope := p.Scope{Kind: tt.kind, OrgID: "org_acme", ID: "resource_one"}
				set, err := p.NewPermissionSet(permission)
				require.NoError(t, err)
				input := p.PermissionCheck{
					Permission: permission,
					Target:     p.ResolvedTarget{Scope: scope, AppID: "app_one", InstallID: "install_one"},
					Grants:     []p.ScopedPermissions{{Scope: scope, Permissions: set}},
					Ceiling:    []p.ScopedPermissions{{Scope: scope, Permissions: set}},
				}
				assert.Equal(t, p.PermissionMatch{Matched: true, Reason: p.ReasonMatched}, p.MatchPermissions(input))
				input.Target.ID = "resource_two"
				assert.Equal(t, p.PermissionMatch{Reason: p.ReasonNoGrant}, p.MatchPermissions(input))
				input.Target.ID = scope.ID
				input.Ceiling[0].Scope.ID = "resource_two"
				assert.Equal(t, p.PermissionMatch{Reason: p.ReasonCeilingExceeded}, p.MatchPermissions(input))
			}
		})
	}
}
