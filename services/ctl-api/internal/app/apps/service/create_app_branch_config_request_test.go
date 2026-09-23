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
			Default:                      true,
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

	require.NotNil(t, groups[1].AutoApproveOnPoliciesPassing)
	require.False(t, groups[1].GetAutoApproveOnPoliciesPassing())
	require.Equal(t, labels.Labels{"env": "prod"}, groups[1].LabelSelector.MatchLabels)

	require.Nil(t, groups[2].LabelSelector)
}
