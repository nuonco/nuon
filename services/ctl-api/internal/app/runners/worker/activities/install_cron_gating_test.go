package activities

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestDecideInstallCronToggle(t *testing.T) {
	tests := []struct {
		name      string
		healthy   bool
		requested InstallCronState
		current   InstallCronState
		want      *InstallCronState
	}{
		{
			name:      "offline past the alert delay switches crons off",
			requested: InstallCronsDisabled,
			current:   InstallCronsEnabled,
			want:      installCronStatePtr(InstallCronsDisabled),
		},
		{
			name:    "offline but not yet past the alert delay is left alone",
			current: InstallCronsEnabled,
		},
		{
			name:      "already disabled is not disabled again",
			requested: InstallCronsDisabled,
			current:   InstallCronsDisabled,
		},
		{
			name:    "a healthy runner brings disabled crons back",
			healthy: true,
			current: InstallCronsDisabled,
			want:    installCronStatePtr(InstallCronsEnabled),
		},
		{
			name:      "healthy sibling outvotes a disable candidate",
			healthy:   true,
			requested: InstallCronsDisabled,
			current:   InstallCronsDisabled,
			want:      installCronStatePtr(InstallCronsEnabled),
		},
		{
			name:      "healthy sibling with crons already on is a noop",
			healthy:   true,
			requested: InstallCronsDisabled,
			current:   InstallCronsEnabled,
		},
		{
			name:    "steady-state healthy install is a noop",
			healthy: true,
			current: InstallCronsEnabled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, decideInstallCronToggle(tt.healthy, tt.requested, tt.current))
		})
	}
}

func TestDecideRunnerHealthInstallCronToggle(t *testing.T) {
	now := time.Now()
	offlineFor := func(d time.Duration) map[string]any {
		return map[string]any{app.RunnerOfflineTSMetadataKey: float64(now.Add(-d).Unix())}
	}

	runner := func(groupType app.RunnerGroupType, status app.RunnerStatus, metadata map[string]any) *app.Runner {
		return &app.Runner{
			ID:     "rnrtest",
			Status: status,
			StatusV2: app.CompositeStatus{
				Status:   app.Status(status),
				Metadata: metadata,
			},
			RunnerGroup: app.RunnerGroup{Type: groupType},
		}
	}

	tests := []struct {
		name    string
		runner  *app.Runner
		want    *InstallCronState
		wantMsg string
	}{
		{
			name:    "install runner offline past the alert delay",
			runner:  runner(app.RunnerGroupTypeInstall, app.RunnerStatusOffline, offlineFor(runnerUnhealthyAlertDelay+time.Minute)),
			want:    installCronStatePtr(InstallCronsDisabled),
			wantMsg: "should be a disable candidate",
		},
		{
			name:    "install runner offline but inside the alert delay",
			runner:  runner(app.RunnerGroupTypeInstall, app.RunnerStatusOffline, offlineFor(time.Minute)),
			want:    nil,
			wantMsg: "should wait out the alert delay",
		},
		{
			name:    "install runner that just transitioned to offline",
			runner:  runner(app.RunnerGroupTypeInstall, app.RunnerStatusActive, nil),
			want:    nil,
			wantMsg: "should only arm offline_ts on the transition tick",
		},
		{
			name:    "intentionally disabled runner disables crons",
			runner:  runner(app.RunnerGroupTypeInstall, app.RunnerStatusDisabled, offlineFor(runnerUnhealthyAlertDelay+time.Minute)),
			want:    installCronStatePtr(InstallCronsDisabled),
			wantMsg: "disabled runners should turn off install crons",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := decideRunnerHealth(now, tt.runner, runnerProcessPresence{MngChecked: true})
			require.Equal(t, tt.want, d.InstallCronToggleDecision, tt.wantMsg)
		})
	}
}

func TestDecideRunnerHealthHealthyRunnerEnablesCrons(t *testing.T) {
	now := time.Now()
	r := &app.Runner{
		ID:          "rnrtest",
		Status:      app.RunnerStatusOffline,
		StatusV2:    app.CompositeStatus{Status: app.Status(app.RunnerStatusOffline)},
		RunnerGroup: app.RunnerGroup{Type: app.RunnerGroupTypeInstall},
	}

	d := decideRunnerHealth(now, r, runnerProcessPresence{HasActiveInstall: true, HasActiveMng: true, MngChecked: true})

	require.Equal(t, runnerHealthResultHealthy, d.Result)
	require.Equal(t, installCronStatePtr(InstallCronsEnabled), d.InstallCronToggleDecision)
}

func installCronStatePtr(value InstallCronState) *InstallCronState {
	return &value
}
