package helpers

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// A branch that took the install through an all-installs group claimed
// everything unclaimed rather than this install, so it hands it over. Anything
// naming the install keeps it.
func TestClaimIsAllInstallsOnly(t *testing.T) {
	t.Parallel()

	install := &app.Install{
		ID:      "install-1",
		Labeled: labels.Labeled{Labels: labels.Labels{"env": "prod"}},
	}

	tests := []struct {
		name   string
		groups []app.AppBranchInstallGroup
		weak   bool
	}{
		{
			name:   "all installs group only",
			groups: []app.AppBranchInstallGroup{{AllInstalls: true}},
			weak:   true,
		},
		{
			name: "explicit install id",
			groups: []app.AppBranchInstallGroup{
				{AllInstalls: true},
				{InstallIDs: []string{"install-1"}},
			},
			weak: false,
		},
		{
			name: "matching selector",
			groups: []app.AppBranchInstallGroup{
				{AllInstalls: true},
				{LabelSelector: &labels.Selector{MatchLabels: labels.Labels{"env": "prod"}}},
			},
			weak: false,
		},
		{
			name: "selector that does not match",
			groups: []app.AppBranchInstallGroup{
				{AllInstalls: true},
				{LabelSelector: &labels.Selector{MatchLabels: labels.Labels{"env": "staging"}}},
			},
			weak: true,
		},
		{
			name: "explicit id for another install",
			groups: []app.AppBranchInstallGroup{
				{AllInstalls: true},
				{InstallIDs: []string{"install-2"}},
			},
			weak: true,
		},
		{
			name:   "no groups at all",
			groups: nil,
			weak:   false,
		},
		{
			name:   "groups but no all installs",
			groups: []app.AppBranchInstallGroup{{InstallIDs: []string{"install-2"}}},
			weak:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.weak, claimIsAllInstallsOnly(tt.groups, install))
		})
	}
}
