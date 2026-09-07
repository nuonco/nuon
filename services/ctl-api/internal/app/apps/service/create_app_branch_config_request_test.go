package service

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/pkg/labels"
)

func TestInstallGroupsFromRequest(t *testing.T) {
	groups := installGroupsFromRequest([]InstallGroupRequest{
		{
			Name:                         "canary",
			Order:                        0,
			InstallIDs:                   []string{"install-1"},
			AutoApproveOnPoliciesPassing: generics.ToPtr(true),
		},
		{
			Name:                         "prod",
			Order:                        1,
			LabelSelector:                &labels.Selector{MatchLabels: labels.Labels{"env": "prod"}},
			AutoApproveOnPoliciesPassing: generics.ToPtr(false),
		},
		{
			Name:        "everything",
			Order:       2,
			AllInstalls: true,
		},
		{
			Name:          "empty-selector",
			Order:         3,
			InstallIDs:    []string{"install-2"},
			LabelSelector: &labels.Selector{},
		},
	})

	require.Len(t, groups, 4)

	require.True(t, groups[0].GetAutoApproveOnPoliciesPassing())
	require.Equal(t, []string{"install-1"}, []string(groups[0].InstallIDs))

	require.NotNil(t, groups[1].AutoApproveOnPoliciesPassing)
	require.False(t, groups[1].GetAutoApproveOnPoliciesPassing())
	require.Equal(t, labels.Labels{"env": "prod"}, groups[1].LabelSelector.MatchLabels)

	require.Nil(t, groups[2].AutoApproveOnPoliciesPassing)
	require.False(t, groups[2].GetAutoApproveOnPoliciesPassing())
	require.True(t, groups[2].AllInstalls)

	require.Nil(t, groups[3].LabelSelector, "an empty selector must not shadow install_ids")
}
