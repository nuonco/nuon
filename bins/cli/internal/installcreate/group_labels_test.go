package installcreate

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func labelGroup(id string, matchLabels map[string]string) *models.AppAppBranchInstallGroup {
	return &models.AppAppBranchInstallGroup{
		ID: id,
		LabelSelector: &models.GithubComNuoncoNuonPkgLabelsSelector{
			MatchLabels: matchLabels,
		},
	}
}

func TestGroupLabels(t *testing.T) {
	t.Run("returns concrete match labels", func(t *testing.T) {
		got, err := GroupLabels(labelGroup("group-id", map[string]string{"env": "staging", "tier": "api"}))
		require.NoError(t, err)
		assert.Equal(t, map[string]string{"env": "staging", "tier": "api"}, got)
	})

	t.Run("rejects static groups", func(t *testing.T) {
		_, err := GroupLabels(&models.AppAppBranchInstallGroup{ID: "group-id"})
		require.Error(t, err)
	})

	t.Run("rejects wildcard selectors", func(t *testing.T) {
		_, err := GroupLabels(labelGroup("group-id", map[string]string{"env": "*"}))
		require.EqualError(t, err, `install group label "env" uses a wildcard and cannot be applied during creation`)
	})
}

func TestEligibleGroups(t *testing.T) {
	groups := []*models.AppAppBranchInstallGroup{
		labelGroup("concrete", map[string]string{"env": "staging"}),
		labelGroup("wildcard", map[string]string{"env": "*"}),
		{ID: "static"},
		nil,
	}

	eligible := EligibleGroups(groups)
	require.Len(t, eligible, 1)
	assert.Equal(t, "concrete", eligible[0].ID)
}

func TestMergeGroupLabels(t *testing.T) {
	t.Run("applies group labels", func(t *testing.T) {
		into := map[string]string{"owner": "team-a"}
		require.NoError(t, MergeGroupLabels(into, map[string]string{"env": "staging"}))
		assert.Equal(t, map[string]string{"owner": "team-a", "env": "staging"}, into)
	})

	t.Run("rejects conflicting user labels", func(t *testing.T) {
		into := map[string]string{"env": "prod"}
		err := MergeGroupLabels(into, map[string]string{"env": "staging"})
		require.EqualError(t, err, `label "env" is "prod", but the selected install group requires "staging"`)
	})

	t.Run("allows matching values", func(t *testing.T) {
		into := map[string]string{"env": "staging"}
		require.NoError(t, MergeGroupLabels(into, map[string]string{"env": "staging"}))
	})
}
