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
			name:  "matching selector",
			group: &app.AppBranchInstallGroup{LabelSelector: &labels.Selector{MatchLabels: labels.Labels{"env": "prod"}}},
			want:  true,
		},
		{
			name:  "selector that does not match",
			group: &app.AppBranchInstallGroup{LabelSelector: &labels.Selector{MatchLabels: labels.Labels{"env": "staging"}}},
		},
		{
			name:  "default does not match directly",
			group: &app.AppBranchInstallGroup{Default: true},
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

	t.Run("two selectors claim it", func(t *testing.T) {
		err := ValidateInstallSingleGroup([]app.AppBranchInstallGroup{
			{Name: "by-label", LabelSelector: &labels.Selector{MatchLabels: labels.Labels{"env": "prod"}}},
			{Name: "also-by-label", LabelSelector: &labels.Selector{MatchLabels: labels.Labels{"env": "prod"}}},
		}, install)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "also-by-label")
		assert.Contains(t, err.Error(), "by-label")
	})

	t.Run("label selector precedes default", func(t *testing.T) {
		groups := []app.AppBranchInstallGroup{
			{Name: "default", Default: true},
			{Name: "prod", LabelSelector: &labels.Selector{MatchLabels: labels.Labels{"env": "prod"}}},
		}
		resolved, err := ResolveInstallGroup(groups, install)
		require.NoError(t, err)
		require.NotNil(t, resolved)
		assert.Equal(t, "prod", resolved.Name)

		unmatched := prodInstall()
		unmatched.Labels = labels.Labels{"env": "dev"}
		resolved, err = ResolveInstallGroup(groups, unmatched)
		require.NoError(t, err)
		require.NotNil(t, resolved)
		assert.Equal(t, "default", resolved.Name)
	})

	t.Run("explicit group overrides matching selector", func(t *testing.T) {
		explicit := prodInstall()
		explicit.AppBranchGroup = "canary"
		resolved, err := ResolveInstallGroup([]app.AppBranchInstallGroup{
			{Name: "canary", Default: true},
			{Name: "prod", LabelSelector: &labels.Selector{MatchLabels: labels.Labels{"env": "prod"}}},
		}, explicit)
		require.NoError(t, err)
		require.NotNil(t, resolved)
		assert.Equal(t, "canary", resolved.Name)
	})
}

func TestInstallGroupsMatching(t *testing.T) {
	t.Parallel()

	got := InstallGroupsMatching([]app.AppBranchInstallGroup{
		{Name: "default", Default: true},
		{Name: "prod", LabelSelector: &labels.Selector{MatchLabels: labels.Labels{"env": "prod"}}},
		{Name: "staging", LabelSelector: &labels.Selector{MatchLabels: labels.Labels{"env": "staging"}}},
	}, prodInstall())

	assert.Equal(t, []string{"prod"}, got)
}
