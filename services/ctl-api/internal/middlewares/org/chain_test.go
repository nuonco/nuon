package org

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
)

func paramsFrom(vals map[string]string) chainParams {
	return func(name string) string { return vals[name] }
}

func TestChainFromParams(t *testing.T) {
	const orgID = "org_1"

	for _, tc := range []struct {
		name   string
		params map[string]string
		want   []authz.Link
	}{
		{
			name:   "bare route is org only",
			params: nil,
			want: []authz.Link{
				{Type: app.LevelOrg, ID: orgID},
			},
		},
		{
			name:   "collection route is org only",
			params: map[string]string{},
			want: []authz.Link{
				{Type: app.LevelOrg, ID: orgID},
			},
		},
		{
			name:   "app route",
			params: map[string]string{"app_id": "app_1"},
			want: []authz.Link{
				{Type: app.LevelApp, ID: "app_1"},
				{Type: app.LevelOrg, ID: orgID},
			},
		},
		{
			name:   "flat install route carries an empty app link",
			params: map[string]string{"install_id": "inl_1"},
			want: []authz.Link{
				{Type: app.LevelInstall, ID: "inl_1"},
				{Type: app.LevelApp, ID: ""},
				{Type: app.LevelOrg, ID: orgID},
			},
		},
		{
			name:   "app-nested install route carries the real app id",
			params: map[string]string{"app_id": "app_1", "install_id": "inl_1"},
			want: []authz.Link{
				{Type: app.LevelInstall, ID: "inl_1"},
				{Type: app.LevelApp, ID: "app_1"},
				{Type: app.LevelOrg, ID: orgID},
			},
		},
		{
			name:   "app-nested branch route",
			params: map[string]string{"app_id": "app_1", "app_branch_id": "brn_1"},
			want: []authz.Link{
				{Type: app.LevelAppBranch, ID: "brn_1"},
				{Type: app.LevelApp, ID: "app_1"},
				{Type: app.LevelOrg, ID: orgID},
			},
		},
		{
			name:   "a name in an id param is passed through untouched",
			params: map[string]string{"app_id": "my-app"},
			want: []authz.Link{
				{Type: app.LevelApp, ID: "my-app"},
				{Type: app.LevelOrg, ID: orgID},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, chainFromParams(paramsFrom(tc.params), orgID))
		})
	}
}

// The org link is what keeps every managed org role working on every route, so
// no param combination may omit it or leave it empty.
func TestChainAlwaysEndsInOrg(t *testing.T) {
	const orgID = "org_1"

	for _, params := range []map[string]string{
		nil,
		{"app_id": "app_1"},
		{"install_id": "inl_1"},
		{"app_branch_id": "brn_1"},
		{"app_id": "app_1", "install_id": "inl_1", "app_branch_id": "brn_1"},
		{"component_id": "cmp_1", "workflow_id": "wrk_1"},
	} {
		chain := chainFromParams(paramsFrom(params), orgID)

		require.NotEmpty(t, chain)
		last := chain[len(chain)-1]
		assert.Equal(t, app.LevelOrg, last.Type)
		assert.Equal(t, orgID, last.ID)
	}
}

// Unrecognized leaf params (component, workflow, runner, …) do not create
// links: those tiers have no grant target, so they authorize from their
// ancestors in the URL.
func TestChainIgnoresUnrecognizedParams(t *testing.T) {
	chain := chainFromParams(paramsFrom(map[string]string{
		"install_id":   "inl_1",
		"component_id": "cmp_1",
		"step_id":      "stp_1",
	}), "org_1")

	assert.Equal(t, []authz.Link{
		{Type: app.LevelInstall, ID: "inl_1"},
		{Type: app.LevelApp, ID: ""},
		{Type: app.LevelOrg, ID: "org_1"},
	}, chain)
}
