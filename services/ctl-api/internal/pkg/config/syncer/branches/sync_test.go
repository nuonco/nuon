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
				InstallNames:                 []string{"canary-install"},
				AutoApproveOnPoliciesPassing: generics.ToPtr(true),
			},
			{
				Name:                         "prod",
				Order:                        1,
				LabelSelector:                map[string]string{"env": "prod"},
				AutoApproveOnPoliciesPassing: generics.ToPtr(false),
			},
			{
				Name:  "manual",
				Order: 2,
				// Omitted in the TOML — stays nil so the getter defaults it off.
			},
		},
	}

	groups, err := buildInstallGroups(branchCfg, map[string]string{"canary-install": "install-1"})
	require.NoError(t, err)
	require.Len(t, groups, 3)

	require.Equal(t, []string{"install-1"}, []string(groups[0].InstallIDs))
	require.NotNil(t, groups[0].AutoApproveOnPoliciesPassing)
	require.True(t, groups[0].GetAutoApproveOnPoliciesPassing())

	require.NotNil(t, groups[1].AutoApproveOnPoliciesPassing)
	require.False(t, groups[1].GetAutoApproveOnPoliciesPassing())

	require.Nil(t, groups[2].AutoApproveOnPoliciesPassing)
	require.False(t, groups[2].GetAutoApproveOnPoliciesPassing())
}
