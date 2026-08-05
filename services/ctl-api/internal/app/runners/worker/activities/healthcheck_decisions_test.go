package activities

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

var decisionNow = time.Date(2026, time.July, 30, 12, 0, 0, 0, time.UTC)

func decisionRunner(groupType app.RunnerGroupType, status app.RunnerStatus, v2Status app.RunnerStatus, metadata map[string]any) *app.Runner {
	return &app.Runner{
		ID:     "rnrtest",
		Status: status,
		StatusV2: app.CompositeStatus{
			Status:   app.Status(v2Status),
			Metadata: metadata,
		},
		RunnerGroup: app.RunnerGroup{Type: groupType},
	}
}

func TestDecideRunnerHealth(t *testing.T) {
	healthyInstall := runnerProcessPresence{HasActiveInstall: true, HasActiveMng: true, MngChecked: true}

	tests := []struct {
		name     string
		runner   *app.Runner
		presence runnerProcessPresence
		want     runnerHealthDecision
	}{
		{
			name:   "skippable status",
			runner: decisionRunner(app.RunnerGroupTypeOrg, app.RunnerStatusProvisioning, app.RunnerStatusProvisioning, nil),
			want:   runnerHealthDecision{Result: "skipped"},
		},
		{
			name:     "healthy active org runner writes nothing",
			runner:   decisionRunner(app.RunnerGroupTypeOrg, app.RunnerStatusActive, app.RunnerStatusActive, nil),
			presence: runnerProcessPresence{HasActiveBuild: true},
			want: runnerHealthDecision{
				Result: "healthy", TargetStatus: app.RunnerStatusActive, Reason: "runner healthy",
			},
		},
		{
			name:   "first failed check arms offline_ts then updates both statuses without alerting",
			runner: decisionRunner(app.RunnerGroupTypeOrg, app.RunnerStatusActive, app.RunnerStatusActive, nil),
			want: runnerHealthDecision{
				Result: "unhealthy", TargetStatus: app.RunnerStatusOffline, Reason: "no active build process",
				SetOfflineTS: true, UpdateLegacy: true, UpdateV2: true,
			},
		},
		{
			name: "offline under alert delay does nothing",
			runner: decisionRunner(app.RunnerGroupTypeOrg, app.RunnerStatusOffline, app.RunnerStatusOffline,
				map[string]any{app.RunnerOfflineTSMetadataKey: decisionNow.Add(-runnerUnhealthyAlertDelay + time.Second).Unix()}),
			want: runnerHealthDecision{
				Result: "unhealthy", TargetStatus: app.RunnerStatusOffline, Reason: "no active build process",
			},
		},
		{
			name: "offline past alert delay alerts with persisted offlineAt",
			runner: decisionRunner(app.RunnerGroupTypeOrg, app.RunnerStatusOffline, app.RunnerStatusOffline,
				map[string]any{app.RunnerOfflineTSMetadataKey: decisionNow.Add(-runnerUnhealthyAlertDelay - time.Minute).Unix()}),
			want: runnerHealthDecision{
				Result: "unhealthy", TargetStatus: app.RunnerStatusOffline, Reason: "no active build process",
				Alert: true, AlertOfflineAt: time.Unix(decisionNow.Add(-runnerUnhealthyAlertDelay-time.Minute).Unix(), 0),
			},
		},
		{
			name:   "offline without timestamp arms only, no alert",
			runner: decisionRunner(app.RunnerGroupTypeOrg, app.RunnerStatusOffline, app.RunnerStatusOffline, nil),
			want: runnerHealthDecision{
				Result: "unhealthy", TargetStatus: app.RunnerStatusOffline, Reason: "no active build process",
				SetOfflineTS: true,
			},
		},
		{
			name: "recovered runner clears stale offline_ts without resetting it",
			runner: decisionRunner(app.RunnerGroupTypeOrg, app.RunnerStatusActive, app.RunnerStatusActive,
				map[string]any{app.RunnerOfflineTSMetadataKey: decisionNow.Add(-time.Hour).Unix()}),
			presence: runnerProcessPresence{HasActiveBuild: true},
			want: runnerHealthDecision{
				Result: "healthy", TargetStatus: app.RunnerStatusActive, Reason: "runner healthy",
				ClearOfflineTS: true,
			},
		},
		{
			name: "legacy correct but v2 stale repairs v2 only",
			runner: decisionRunner(app.RunnerGroupTypeOrg, app.RunnerStatusOffline, app.RunnerStatusActive,
				map[string]any{app.RunnerOfflineTSMetadataKey: decisionNow.Add(-time.Hour).Unix()}),
			want: runnerHealthDecision{
				Result: "unhealthy", TargetStatus: app.RunnerStatusOffline, Reason: "no active build process",
				UpdateV2: true,
			},
		},
		{
			name:     "install runner missing install process is unhealthy",
			runner:   decisionRunner(app.RunnerGroupTypeInstall, app.RunnerStatusActive, app.RunnerStatusActive, nil),
			presence: runnerProcessPresence{HasActiveMng: true, MngChecked: true},
			want: runnerHealthDecision{
				Result: "unhealthy", TargetStatus: app.RunnerStatusOffline, Reason: "no active install process",
				SetOfflineTS: true, UpdateLegacy: true, UpdateV2: true,
				SetMissingMng: generics.ToPtr(false),
			},
		},
		{
			name: "mng metadata unchanged writes nothing",
			runner: decisionRunner(app.RunnerGroupTypeInstall, app.RunnerStatusActive, app.RunnerStatusActive,
				map[string]any{"missing_mng_process": false}),
			presence: healthyInstall,
			want: runnerHealthDecision{
				Result: "healthy", TargetStatus: app.RunnerStatusActive, Reason: "runner healthy",
			},
		},
		{
			name: "mng metadata flip is written",
			runner: decisionRunner(app.RunnerGroupTypeInstall, app.RunnerStatusActive, app.RunnerStatusActive,
				map[string]any{"missing_mng_process": false}),
			presence: runnerProcessPresence{HasActiveInstall: true, HasActiveMng: false, MngChecked: true},
			want: runnerHealthDecision{
				Result: "healthy", TargetStatus: app.RunnerStatusActive, Reason: "runner healthy",
				SetMissingMng: generics.ToPtr(true),
			},
		},
		{
			name:     "mng unchecked never writes mng metadata",
			runner:   decisionRunner(app.RunnerGroupTypeInstall, app.RunnerStatusActive, app.RunnerStatusActive, nil),
			presence: runnerProcessPresence{HasActiveInstall: true},
			want: runnerHealthDecision{
				Result: "healthy", TargetStatus: app.RunnerStatusActive, Reason: "runner healthy",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := decideRunnerHealth(decisionNow, tc.runner, tc.presence)
			require.Equal(t, tc.want, got)
		})
	}
}

func decisionProcess(status app.RunnerProcessStatus, metadata map[string]any) *app.RunnerProcess {
	return &app.RunnerProcess{
		ID: "rnptest",
		CompositeStatus: app.CompositeStatus{
			Status:   app.Status(status),
			Metadata: metadata,
		},
	}
}

func TestDecideProcessHealth(t *testing.T) {
	fresh := decisionNow.Add(-10 * time.Second)
	offlineAge := decisionNow.Add(-90 * time.Second)
	inactiveAge := decisionNow.Add(-6 * time.Minute)

	tests := []struct {
		name        string
		process     *app.RunnerProcess
		heartbeatAt *time.Time
		want        processHealthAction
	}{
		{"pending status is noop", decisionProcess(app.RunnerProcessStatusPendingShutdown, nil), &fresh, processActionNoop},
		{"shutdown requested short-circuits", decisionProcess(app.RunnerProcessStatusActive, map[string]any{"shutdown_requested": true}), &inactiveAge, processActionShutdown},
		{"nil shutdown_requested value is ignored", decisionProcess(app.RunnerProcessStatusActive, map[string]any{"shutdown_requested": nil}), &fresh, processActionActive},
		{"no heartbeat is noop", decisionProcess(app.RunnerProcessStatusActive, nil), nil, processActionNoop},
		{"fresh heartbeat is active", decisionProcess(app.RunnerProcessStatusActive, nil), &fresh, processActionActive},
		{"offline process with fresh heartbeat flips active", decisionProcess(app.RunnerProcessStatusOffline, nil), &fresh, processActionActive},
		{"90s silence is offline", decisionProcess(app.RunnerProcessStatusActive, nil), &offlineAge, processActionOffline},
		{"6m silence is inactive", decisionProcess(app.RunnerProcessStatusActive, nil), &inactiveAge, processActionInactive},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, decideProcessHealth(decisionNow, tc.process, tc.heartbeatAt))
		})
	}
}

func TestDecideVersionWarning(t *testing.T) {
	tests := []struct {
		name        string
		configured  string
		reported    string
		wantWarning bool
		wantEvent   bool
	}{
		{"empty reported", "1.2.3", "", false, false},
		{"match", "1.2.3", "1.2.3", false, false},
		{"empty configured", "", "1.2.3", false, false},
		{"cloud tracks api", "cloud", "1.2.3", false, false},
		{"alias tag warns", "stable", "1.2.3", true, false},
		{"semver mismatch warns", "1.2.4", "1.2.3", true, false},
		{"latest configured emits event", "latest", "1.2.3", true, true},
		{"latest reported emits event", "1.2.3", "latest", true, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			warning, event := decideVersionWarning(tc.configured, tc.reported)
			require.Equal(t, tc.wantWarning, warning != "", warning)
			require.Equal(t, tc.wantEvent, event)
		})
	}
}
