package helpers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func prodInstall() *app.Install {
	return &app.Install{
		ID:      "install-1",
		Name:    "prod-1",
		Labeled: labels.Labeled{Labels: labels.Labels{"env": "prod"}},
	}
}

func TestInstallMatchesGroup(t *testing.T) {
	t.Parallel()

	install := prodInstall()

	tests := []struct {
		name  string
		group *app.AppBranchInstallGroup
		want  bool
	}{
		{
			name:  "nil group",
			group: nil,
		},
		{
			name:  "all installs takes everything the branch owns",
			group: &app.AppBranchInstallGroup{AllInstalls: true},
			want:  true,
		},
		{
			name:  "matching selector",
			group: &app.AppBranchInstallGroup{LabelSelector: &labels.Selector{MatchLabels: labels.Labels{"env": "prod"}}},
			want:  true,
		},
		{
			name:  "selector that does not match",
			group: &app.AppBranchInstallGroup{LabelSelector: &labels.Selector{MatchLabels: labels.Labels{"env": "staging"}}},
		},
		{
			name:  "explicit id",
			group: &app.AppBranchInstallGroup{InstallIDs: []string{"install-0", "install-1"}},
			want:  true,
		},
		{
			name:  "explicit id for another install",
			group: &app.AppBranchInstallGroup{InstallIDs: []string{"install-2"}},
		},
		{
			name:  "empty group targets nothing",
			group: &app.AppBranchInstallGroup{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, InstallMatchesGroup(tt.group, install))
		})
	}
}

// An install two groups target would be planned and deployed twice in the same
// branch run, in an order nothing defines.
func TestValidateInstallSingleGroup(t *testing.T) {
	t.Parallel()

	install := prodInstall()

	t.Run("no groups", func(t *testing.T) {
		require.NoError(t, ValidateInstallSingleGroup(nil, install))
	})

	t.Run("one matching group", func(t *testing.T) {
		require.NoError(t, ValidateInstallSingleGroup([]app.AppBranchInstallGroup{
			{Name: "prod", LabelSelector: &labels.Selector{MatchLabels: labels.Labels{"env": "prod"}}},
			{Name: "staging", LabelSelector: &labels.Selector{MatchLabels: labels.Labels{"env": "staging"}}},
		}, install))
	})

	t.Run("selector and explicit id both claim it", func(t *testing.T) {
		err := ValidateInstallSingleGroup([]app.AppBranchInstallGroup{
			{Name: "by-label", LabelSelector: &labels.Selector{MatchLabels: labels.Labels{"env": "prod"}}},
			{Name: "by-id", InstallIDs: []string{"install-1"}},
		}, install)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "by-id")
		assert.Contains(t, err.Error(), "by-label")
	})

	t.Run("all installs alongside a matching selector", func(t *testing.T) {
		err := ValidateInstallSingleGroup([]app.AppBranchInstallGroup{
			{Name: "everything", AllInstalls: true},
			{Name: "prod", LabelSelector: &labels.Selector{MatchLabels: labels.Labels{"env": "prod"}}},
		}, install)
		require.Error(t, err)
	})

	t.Run("all installs alongside a selector that misses", func(t *testing.T) {
		require.NoError(t, ValidateInstallSingleGroup([]app.AppBranchInstallGroup{
			{Name: "everything", AllInstalls: true},
			{Name: "staging", LabelSelector: &labels.Selector{MatchLabels: labels.Labels{"env": "staging"}}},
		}, install))
	})
}

func TestInstallGroupsMatching(t *testing.T) {
	t.Parallel()

	got := InstallGroupsMatching([]app.AppBranchInstallGroup{
		{Name: "everything", AllInstalls: true},
		{Name: "staging", LabelSelector: &labels.Selector{MatchLabels: labels.Labels{"env": "staging"}}},
		{Name: "by-id", InstallIDs: []string{"install-1"}},
	}, prodInstall())

	assert.Equal(t, []string{"everything", "by-id"}, got)
}
