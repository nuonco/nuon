package installs

import (
	"testing"

	"github.com/nuonco/nuon/sdks/nuon-go/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequireInstallRegion(t *testing.T) {
	for _, cloud := range []string{"AWS", "Azure", "GCP"} {
		t.Run(cloud, func(t *testing.T) {
			require.EqualError(t, requireInstallRegion("", cloud), "--region is required for "+cloud+" installs")
			require.NoError(t, requireInstallRegion("us-west-2", cloud))
		})
	}
}

func TestParseInstallInputs(t *testing.T) {
	inputs, err := parseInstallInputs([]string{"token=part=two", "empty="})
	require.NoError(t, err)
	assert.Equal(t, map[string]string{"token": "part=two", "empty": ""}, inputs)

	for _, input := range []string{"missing-separator", "=missing-name"} {
		t.Run(input, func(t *testing.T) {
			_, err := parseInstallInputs([]string{input})
			require.EqualError(t, err, `invalid input "`+input+`": expected name=value`)
		})
	}
}

func TestCreateInstallGroupLabels(t *testing.T) {
	t.Run("returns concrete match labels", func(t *testing.T) {
		group := &models.AppAppBranchInstallGroup{
			ID: "group-id",
			LabelSelector: &models.GithubComNuoncoNuonPkgLabelsSelector{
				MatchLabels: map[string]string{"env": "staging", "tier": "api"},
			},
		}

		got, err := createInstallGroupLabels(group)
		require.NoError(t, err)
		assert.Equal(t, map[string]string{"env": "staging", "tier": "api"}, got)
	})

	t.Run("rejects static groups", func(t *testing.T) {
		_, err := createInstallGroupLabels(&models.AppAppBranchInstallGroup{ID: "group-id"})
		require.Error(t, err)
	})

	t.Run("rejects wildcard selectors", func(t *testing.T) {
		group := &models.AppAppBranchInstallGroup{
			ID: "group-id",
			LabelSelector: &models.GithubComNuoncoNuonPkgLabelsSelector{
				MatchLabels: map[string]string{"env": "*"},
			},
		}

		_, err := createInstallGroupLabels(group)
		require.EqualError(t, err, `install group label "env" uses a wildcard and cannot be applied during creation`)
	})
}
