package branches

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/generics"
)

func TestBuildInstallGroupsThreadsAutoApproveOnPoliciesPassing(t *testing.T) {
	branchCfg := &config.AppBranchConfig{
		Name: "main",
		InstallGroups: []config.AppBranchInstallGroupConfig{
			{
				Name:                         "canary",
				Order:                        0,
				Default:                      true,
				AutoApproveOnPoliciesPassing: generics.ToPtr(true),
			},
			{
				Name:                         "prod",
				Order:                        1,
				LabelSelector:                map[string]string{"env": "prod"},
				AutoApproveOnPoliciesPassing: generics.ToPtr(false),
			},
			{
				Name:          "manual",
				Order:         2,
				LabelSelector: map[string]string{"tier": "manual"},
				// Omitted in the TOML — stays nil so the getter defaults it off.
			},
		},
	}

	groups := buildInstallGroups(branchCfg)
	require.Len(t, groups, 3)

	require.True(t, groups[0].Default)
	require.NotNil(t, groups[0].AutoApproveOnPoliciesPassing)
	require.True(t, groups[0].GetAutoApproveOnPoliciesPassing())

	require.NotNil(t, groups[1].AutoApproveOnPoliciesPassing)
	require.False(t, groups[1].GetAutoApproveOnPoliciesPassing())

	require.Nil(t, groups[2].AutoApproveOnPoliciesPassing)
	require.False(t, groups[2].GetAutoApproveOnPoliciesPassing())
}
