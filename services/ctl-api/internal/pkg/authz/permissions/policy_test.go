package permissions_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	p "github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
)

func TestPolicyCompilation(t *testing.T) {
	org := p.Scope{Kind: p.KindOrg, OrgID: "org_acme", ID: "org_acme"}
	app := p.Scope{Kind: p.KindApp, OrgID: "org_acme", ID: "app_one"}
	const bundle = `{"kind":"bundle","api":"ctl-api","binding_kind":"app","permissions":["log_stream.read","app.read","app.read"]}`
	const workload = `{"kind":"resource_rules","api":"runner","rules":[
		{"scope":{"kind":"runner-job","org_id":"org_acme","id":"job_one"},"permissions":["runner_job.write"]},
		{"scope":{"kind":"log_stream","org_id":"org_acme","id":"stream_one"},"permissions":["log_stream.write"]}
	]}`
	tests := map[string]struct {
		doc     string
		owner   string
		binding p.Scope
		want    []p.ScopedPermissions
		wantErr bool
	}{
		"bundle uses binding without expanding current role": {
			bundle, org.ID, app,
			[]p.ScopedPermissions{{Scope: app, Permissions: p.PermissionSet{p.AppRead: {}, p.LogStreamRead: {}}}}, false,
		},
		"telemetry bot bundle retains read and ingestion at its install": {
			`{"kind":"bundle","api":"ctl-api","binding_kind":"install","permissions":["telemetry.read","telemetry.write"]}`,
			org.ID, p.Scope{Kind: p.KindInstall, OrgID: org.ID, ID: "install_one"},
			[]p.ScopedPermissions{{Scope: p.Scope{Kind: p.KindInstall, OrgID: org.ID, ID: "install_one"}, Permissions: p.PermissionSet{p.TelemetryRead: {}, p.TelemetryWrite: {}}}}, false,
		},
		"protocol permission is not excluded from bundles": {
			`{"kind":"bundle","api":"ctl-api","binding_kind":"org","permissions":["runner_job.claim"]}`, org.ID, org,
			[]p.ScopedPermissions{{Scope: org, Permissions: p.PermissionSet{p.RunnerJobClaim: {}}}}, false,
		},
		"workload keeps permission resource pairs": {
			workload, org.ID, org,
			[]p.ScopedPermissions{
				{Scope: p.Scope{Kind: p.KindRunnerJob, OrgID: org.ID, ID: "job_one"}, Permissions: p.PermissionSet{p.RunnerJobWrite: {}}},
				{Scope: p.Scope{Kind: p.KindLogStream, OrgID: org.ID, ID: "stream_one"}, Permissions: p.PermissionSet{p.LogStreamWrite: {}}},
			}, false,
		},
		"empty bundle grants nothing": {
			`{"kind":"bundle","api":"ctl-api","binding_kind":"app","permissions":[]}`, org.ID, app,
			[]p.ScopedPermissions{{Scope: app, Permissions: p.PermissionSet{}}}, false,
		},
		"empty private policy grants nothing":             {`{"kind":"resource_rules","api":"runner","rules":[]}`, org.ID, org, []p.ScopedPermissions{}, false},
		"wrong bundle binding kind":                       {bundle, org.ID, org, nil, true},
		"missing owner":                                   {bundle, "", app, nil, true},
		"owner differs from binding":                      {bundle, "org_other", app, nil, true},
		"malformed binding":                               {bundle, org.ID, p.Scope{Kind: p.KindApp, OrgID: org.ID, ID: "*"}, nil, true},
		"private rules cannot widen association":          {workload, "org_other", p.Scope{Kind: p.KindOrg, OrgID: "org_other", ID: "org_other"}, nil, true},
		"private policy does not infer narrower ancestry": {workload, org.ID, app, nil, true},
		"later cross tenant rule invalidates entire policy": {
			`{"kind":"resource_rules","api":"runner","rules":[
				{"scope":{"kind":"log_stream","org_id":"org_acme","id":"stream_one"},"permissions":["log_stream.write"]},
				{"scope":{"kind":"log_stream","org_id":"org_other","id":"stream_two"},"permissions":["log_stream.write"]}
			]}`, org.ID, org, nil, true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			policy, err := p.DecodePolicy(3, []byte(tt.doc))
			require.NoError(t, err)
			got, err := policy.Compile(tt.owner, tt.binding)
			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, got)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
			encoded, err := json.Marshal(policy)
			require.NoError(t, err)
			roundTrip, err := p.DecodePolicy(3, encoded)
			require.NoError(t, err)
			again, err := roundTrip.Compile(tt.owner, tt.binding)
			require.NoError(t, err)
			assert.Equal(t, tt.want, again)
			for i := range got {
				clear(got[i].Permissions)
				got[i].Permissions[p.OrgWrite] = struct{}{}
			}
			assert.Equal(t, tt.want, again)
			fresh, err := policy.Compile(tt.owner, tt.binding)
			require.NoError(t, err)
			assert.Equal(t, tt.want, fresh)
		})
	}
}

func TestTelemetryBundleRespectsScopeAndCeiling(t *testing.T) {
	policy, err := p.DecodePolicy(3, []byte(`{"kind":"bundle","api":"ctl-api","binding_kind":"install","permissions":["telemetry.read","telemetry.write"]}`))
	require.NoError(t, err)
	grants, err := policy.Compile("org_acme", p.Scope{Kind: p.KindInstall, OrgID: "org_acme", ID: "install_one"})
	require.NoError(t, err)
	for name, tc := range map[string]struct {
		permission p.PermissionID
		orgID      string
		installID  string
		ceiling    p.PermissionSet
		want       p.PermissionMatch
	}{
		"read":                     {p.TelemetryRead, "org_acme", "install_one", p.PermissionSet{p.TelemetryRead: {}}, p.PermissionMatch{Matched: true, Reason: p.ReasonMatched}},
		"ingest":                   {p.TelemetryWrite, "org_acme", "install_one", p.PermissionSet{p.TelemetryWrite: {}}, p.PermissionMatch{Matched: true, Reason: p.ReasonMatched}},
		"sibling install":          {p.TelemetryWrite, "org_acme", "install_two", p.PermissionSet{p.TelemetryWrite: {}}, p.PermissionMatch{Reason: p.ReasonNoGrant}},
		"other tenant":             {p.TelemetryWrite, "org_other", "install_one", p.PermissionSet{p.TelemetryWrite: {}}, p.PermissionMatch{Reason: p.ReasonNoGrant}},
		"read-only credential":     {p.TelemetryWrite, "org_acme", "install_one", p.PermissionSet{p.TelemetryRead: {}}, p.PermissionMatch{Reason: p.ReasonCeilingExceeded}},
		"no implied config access": {p.TelemetryConfigRead, "org_acme", "install_one", p.PermissionSet{p.TelemetryConfigRead: {}}, p.PermissionMatch{Reason: p.ReasonNoGrant}},
	} {
		t.Run(name, func(t *testing.T) {
			got := p.MatchPermissions(p.PermissionCheck{
				Permission: tc.permission,
				Target:     p.ResolvedTarget{Scope: p.Scope{Kind: p.KindTelemetry, OrgID: tc.orgID, ID: tc.installID}, AppID: "app_one", InstallID: tc.installID},
				Grants:     grants,
				Ceiling:    []p.ScopedPermissions{{Scope: p.Scope{Kind: p.KindOrg, OrgID: tc.orgID, ID: tc.orgID}, Permissions: tc.ceiling}},
			})
			require.Equal(t, tc.want, got)
		})
	}
}

func TestPolicyRejectsInvalidDocuments(t *testing.T) {
	tests := map[string]struct{ doc string }{
		"null":                         {`null`},
		"missing kind":                 {`{"api":"ctl-api","binding_kind":"org","permissions":[]}`},
		"unknown kind":                 {`{"kind":"allow_all","api":"ctl-api"}`},
		"missing permissions":          {`{"kind":"bundle","api":"ctl-api","binding_kind":"org"}`},
		"null permissions":             {`{"kind":"bundle","api":"ctl-api","binding_kind":"org","permissions":null}`},
		"null permission entry":        {`{"kind":"bundle","api":"ctl-api","binding_kind":"org","permissions":[null]}`},
		"space separated string":       {`{"kind":"bundle","api":"ctl-api","binding_kind":"org","permissions":"app.read install.read"}`},
		"unknown permission":           {`{"kind":"bundle","api":"ctl-api","binding_kind":"org","permissions":["app.read","app.future"]}`},
		"legacy all":                   {`{"kind":"bundle","api":"ctl-api","binding_kind":"org","permissions":["all"]}`},
		"wrong bundle api":             {`{"kind":"bundle","api":"runner","binding_kind":"org","permissions":[]}`},
		"unknown binding kind":         {`{"kind":"bundle","api":"ctl-api","binding_kind":"runner","permissions":[]}`},
		"unknown field":                {`{"kind":"bundle","api":"ctl-api","binding_kind":"org","permissions":[],"skip_tenant_check":true}`},
		"case variant field":           {`{"kind":"bundle","API":"ctl-api","binding_kind":"org","permissions":[]}`},
		"mixed variants":               {`{"kind":"bundle","api":"ctl-api","binding_kind":"org","permissions":[],"rules":[]}`},
		"duplicate fields":             {`{"kind":"bundle","api":"ctl-api","binding_kind":"org","permissions":[],"permissions":["org.write"]}`},
		"multiple documents":           {`{"kind":"bundle","api":"ctl-api","binding_kind":"org","permissions":[]} {}`},
		"missing rules":                {`{"kind":"resource_rules","api":"runner"}`},
		"null rules":                   {`{"kind":"resource_rules","api":"runner","rules":null}`},
		"unknown api":                  {`{"kind":"resource_rules","api":"future","rules":[]}`},
		"missing scope":                {`{"kind":"resource_rules","api":"runner","rules":[{"permissions":["log_stream.write"]}]}`},
		"missing rule permissions":     {`{"kind":"resource_rules","api":"runner","rules":[{"scope":{"kind":"log_stream","org_id":"org_acme","id":"stream_one"}}]}`},
		"broad workload scope":         {`{"kind":"resource_rules","api":"runner","rules":[{"scope":{"kind":"org","org_id":"org_acme","id":"org_acme"},"permissions":["runner_job.claim"]}]}`},
		"wrong permission resource":    {`{"kind":"resource_rules","api":"runner","rules":[{"scope":{"kind":"log_stream","org_id":"org_acme","id":"stream_one"},"permissions":["runner_job.write"]}]}`},
		"duplicate nested scope field": {`{"kind":"resource_rules","api":"runner","rules":[{"scope":{"kind":"log_stream","org_id":"org_acme","id":"stream_one","id":"stream_two"},"permissions":["log_stream.write"]}]}`},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			policy, err := p.DecodePolicy(3, []byte(tt.doc))
			require.Error(t, err)
			assert.Equal(t, p.Policy{}, policy)
		})
	}
}

func TestPolicyFormatAndCanonicalPermissions(t *testing.T) {
	const doc = `{"kind":"bundle","api":"ctl-api","binding_kind":"install","permissions":["log_stream.read","install.read","install.read"]}`
	for name, tt := range map[string]struct{ version int }{
		"missing": {0}, "legacy v1": {1}, "legacy v2": {2}, "future": {4},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := p.DecodePolicy(tt.version, []byte(doc))
			require.Error(t, err)
		})
	}
	policy, err := p.DecodePolicy(3, []byte(doc))
	require.NoError(t, err)
	assert.Equal(t, p.APICtl, policy.API())
	encoded, err := json.Marshal(policy)
	require.NoError(t, err)
	assert.JSONEq(t, `{"kind":"bundle","api":"ctl-api","binding_kind":"install","permissions":["install.read","log_stream.read"]}`, string(encoded))
	_, err = json.Marshal(p.Policy{})
	require.Error(t, err)
	_, err = (p.Policy{}).Compile("org_acme", p.Scope{Kind: p.KindOrg, OrgID: "org_acme", ID: "org_acme"})
	require.Error(t, err)
}
