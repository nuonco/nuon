package v2

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestToggleTransitionFor(t *testing.T) {
	component := func(status app.InstallComponentStatus) *app.InstallComponent {
		return &app.InstallComponent{Status: status}
	}

	for name, tc := range map[string]struct {
		effEnabled bool
		ic         *app.InstallComponent
		toggled    bool
		want       toggleTransition
	}{
		"enable never deployed":    {effEnabled: true, ic: component(""), want: toggleEnable},
		"enable missing":           {effEnabled: true, ic: nil, want: toggleEnable},
		"enable failed":            {effEnabled: true, ic: component(app.InstallComponentStatusError), want: toggleEnable},
		"enable already active":    {effEnabled: true, ic: component(app.InstallComponentStatusActive), want: toggleNone},
		"disable active":           {ic: component(app.InstallComponentStatusActive), toggled: true, want: toggleDisable},
		"disable failed deploy":    {ic: component(app.InstallComponentStatusError), toggled: true, want: toggleDisable},
		"disable noop":             {ic: component(app.InstallComponentStatusNoop), toggled: true, want: toggleDisable},
		"skip never deployed":      {ic: component(""), toggled: true, want: toggleSkip},
		"skip torn down":           {ic: component(app.InstallComponentStatusDeleted), toggled: true, want: toggleSkip},
		"untoggled never deployed": {ic: component(""), want: toggleNone},
	} {
		t.Run(name, func(t *testing.T) {
			require.Equal(t, tc.want, toggleTransitionFor(tc.effEnabled, tc.ic, tc.toggled))
		})
	}
}
