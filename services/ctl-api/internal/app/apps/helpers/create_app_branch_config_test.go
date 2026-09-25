package helpers

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestWithDefaultInstallGroup(t *testing.T) {
	t.Run("seeds a default group when none are declared", func(t *testing.T) {
		groups := WithDefaultInstallGroup(nil)
		require.Len(t, groups, 1)
		require.Equal(t, DefaultAppBranchInstallGroupName, groups[0].Name)
		require.True(t, groups[0].Default)
		require.Nil(t, groups[0].LabelSelector)
	})

	t.Run("leaves declared groups untouched", func(t *testing.T) {
		declared := []app.AppBranchInstallGroup{
			{Name: "default", Order: 0, Default: true},
			{Name: "manual", Order: 1},
			{Name: "prod", Order: 2, LabelSelector: &labels.Selector{MatchLabels: labels.Labels{"env": "prod"}}},
		}
		require.Equal(t, declared, WithDefaultInstallGroup(declared))
	})
}
