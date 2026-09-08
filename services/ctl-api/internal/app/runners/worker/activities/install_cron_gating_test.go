package activities

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestDecideCronGate(t *testing.T) {
	tests := []struct {
		name              string
		healthy           bool
		disableCandidate  bool
		currentlyDisabled bool
		want              cronGateAction
	}{
		{
			name:             "offline past the alert delay switches crons off",
			disableCandidate: true,
			want:             cronGateDisable,
		},
		{
			name: "offline but not yet past the alert delay is left alone",
			want: cronGateNoop,
		},
		{
			name:              "already disabled is not disabled again",
			disableCandidate:  true,
			currentlyDisabled: true,
			want:              cronGateNoop,
		},
		{
			name:              "a healthy runner brings disabled crons back",
			healthy:           true,
			currentlyDisabled: true,
			want:              cronGateEnable,
		},
		{
			// A group where one runner is stuck offline but another is up must
			// keep running: the disable candidate loses to the healthy sibling.
			name:              "healthy sibling outvotes a disable candidate",
			healthy:           true,
			disableCandidate:  true,
			currentlyDisabled: true,
			want:              cronGateEnable,
		},
		{
			name:             "healthy sibling with crons already on is a noop",
			healthy:          true,
			disableCandidate: true,
			want:             cronGateNoop,
		},
		{
			name:    "steady-state healthy install is a noop",
			healthy: true,
			want:    cronGateNoop,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, decideCronGate(tt.healthy, tt.disableCandidate, tt.currentlyDisabled))
		})
	}
}

func TestDecideRunnerHealthDisableInstallCrons(t *testing.T) {
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
		want    bool
		wantMsg string
	}{
		{
			name:    "install runner offline past the alert delay",
			runner:  runner(app.RunnerGroupTypeInstall, app.RunnerStatusOffline, offlineFor(runnerUnhealthyAlertDelay+time.Minute)),
			want:    true,
			wantMsg: "should be a disable candidate",
		},
		{
			name:    "install runner offline but inside the alert delay",
			runner:  runner(app.RunnerGroupTypeInstall, app.RunnerStatusOffline, offlineFor(time.Minute)),
			want:    false,
			wantMsg: "should wait out the alert delay",
		},
		{
			name:    "install runner that just transitioned to offline",
			runner:  runner(app.RunnerGroupTypeInstall, app.RunnerStatusActive, nil),
			want:    false,
			wantMsg: "should only arm offline_ts on the transition tick",
		},
		{
			name:    "org runner never gates install crons",
			runner:  runner(app.RunnerGroupTypeOrg, app.RunnerStatusOffline, offlineFor(runnerUnhealthyAlertDelay+time.Minute)),
			want:    false,
			wantMsg: "org runners have no install",
		},
		{
			name:    "intentionally disabled runner is skipped",
			runner:  runner(app.RunnerGroupTypeInstall, app.RunnerStatusDisabled, offlineFor(runnerUnhealthyAlertDelay+time.Minute)),
			want:    false,
			wantMsg: "disabled runners never report health",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := decideRunnerHealth(now, tt.runner, runnerProcessPresence{MngChecked: true})
			require.Equal(t, tt.want, d.DisableInstallCrons, tt.wantMsg)
		})
	}
}

func TestDecideRunnerHealthHealthyRunnerIsNoDisableCandidate(t *testing.T) {
	now := time.Now()
	r := &app.Runner{
		ID:          "rnrtest",
		Status:      app.RunnerStatusOffline,
		StatusV2:    app.CompositeStatus{Status: app.Status(app.RunnerStatusOffline)},
		RunnerGroup: app.RunnerGroup{Type: app.RunnerGroupTypeInstall},
	}

	d := decideRunnerHealth(now, r, runnerProcessPresence{HasActiveInstall: true, HasActiveMng: true, MngChecked: true})

	require.Equal(t, "healthy", d.Result)
	require.False(t, d.DisableInstallCrons)
}
