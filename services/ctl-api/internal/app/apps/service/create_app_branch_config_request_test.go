package service

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

func TestInstallGroupsFromRequest(t *testing.T) {
	groups := installGroupsFromRequest([]InstallGroupRequest{
		{
			Name:                         "canary",
			Order:                        0,
			Default:                      true,
			LabelSelector:                &labels.Selector{MatchLabels: labels.Labels{"type": "push"}},
			AutoApproveOnPoliciesPassing: generics.ToPtr(true),
		},
		{
			Name:                         "prod",
			Order:                        1,
			LabelSelector:                &labels.Selector{MatchLabels: labels.Labels{"env": "prod"}},
			AutoApproveOnPoliciesPassing: generics.ToPtr(false),
		},
		{
			Name:          "empty-selector",
			Order:         2,
			LabelSelector: &labels.Selector{},
		},
	})

	require.Len(t, groups, 3)

	require.True(t, groups[0].GetAutoApproveOnPoliciesPassing())
	require.True(t, groups[0].Default)
	require.Equal(t, labels.Labels{"type": "push"}, groups[0].LabelSelector.MatchLabels)

	require.NotNil(t, groups[1].AutoApproveOnPoliciesPassing)
	require.False(t, groups[1].GetAutoApproveOnPoliciesPassing())
	require.Equal(t, labels.Labels{"env": "prod"}, groups[1].LabelSelector.MatchLabels)

	require.Nil(t, groups[2].LabelSelector)
}

func TestCreateAppBranchConfigRequestValidateInstallGroups(t *testing.T) {
	tests := map[string]struct {
		groups  []InstallGroupRequest
		wantErr string
	}{
		"no groups is allowed": {
			groups: nil,
		},
		"pinned only group needs neither labels nor default": {
			groups: []InstallGroupRequest{
				{Name: "default", Order: 0, Default: true},
				{Name: "manual", Order: 1},
			},
		},
		"default group may also carry labels": {
			groups: []InstallGroupRequest{
				{Name: "push", Order: 0, Default: true, LabelSelector: &labels.Selector{MatchLabels: labels.Labels{"type": "push"}}},
			},
		},
		"no default group": {
			groups: []InstallGroupRequest{
				{Name: "prod", Order: 0, LabelSelector: &labels.Selector{MatchLabels: labels.Labels{"env": "prod"}}},
			},
			wantErr: "one install group must be default",
		},
		"pinned only group without a default": {
			groups: []InstallGroupRequest{
				{Name: "manual", Order: 0},
			},
			wantErr: "one install group must be default",
		},
		"two default groups": {
			groups: []InstallGroupRequest{
				{Name: "a", Order: 0, Default: true},
				{Name: "b", Order: 1, Default: true},
			},
			wantErr: "only one install group can be default",
		},
		"duplicate label selectors": {
			groups: []InstallGroupRequest{
				{Name: "a", Order: 0, Default: true, LabelSelector: &labels.Selector{MatchLabels: labels.Labels{"env": "prod", "tier": "web"}}},
				{Name: "b", Order: 1, LabelSelector: &labels.Selector{MatchLabels: labels.Labels{"tier": "web", "env": "prod"}}},
			},
			wantErr: `install groups "a" and "b" have the same label selector`,
		},
		"distinct label selectors": {
			groups: []InstallGroupRequest{
				{Name: "a", Order: 0, Default: true, LabelSelector: &labels.Selector{MatchLabels: labels.Labels{"env": "prod"}}},
				{Name: "b", Order: 1, LabelSelector: &labels.Selector{MatchLabels: labels.Labels{"env": "staging"}}},
			},
		},
		"empty selectors on pinned groups are allowed": {
			groups: []InstallGroupRequest{
				{Name: "default", Order: 0, Default: true},
				{Name: "manual-a", Order: 1},
				{Name: "manual-b", Order: 2},
			},
		},
	}

	v := validator.New()
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			req := &CreateAppBranchConfigRequest{InstallGroups: tc.groups}
			err := req.Validate(v)
			if tc.wantErr == "" {
				require.NoError(t, err)
				return
			}
			require.Error(t, err)
			var userErr stderr.ErrUser
			require.ErrorAs(t, err, &userErr)
			require.Equal(t, tc.wantErr, userErr.Description)
		})
	}
}
