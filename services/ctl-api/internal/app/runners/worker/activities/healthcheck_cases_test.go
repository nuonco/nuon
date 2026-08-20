package activities

import (
	"time"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

var corpusNow = time.Date(2026, time.August, 6, 12, 0, 0, 0, time.UTC)

type runnerHealthWant struct {
	result         string
	reason         string
	setMissingMng  *bool
	armOfflineTS   bool
	clearOfflineTS bool
	legacyStatus   *app.RunnerStatus
	v2Status       *app.RunnerStatus
	alert          bool
	alertOfflineAt int64
}

type runnerHealthCase struct {
	name          string
	groupType     app.RunnerGroupType
	status        app.RunnerStatus
	v2Status      app.RunnerStatus
	metadata      map[string]any
	activeBuild   bool
	activeInstall bool
	activeMng     bool
	mngChecked    bool
	want          runnerHealthWant
}

func runnerStatusPtr(s app.RunnerStatus) *app.RunnerStatus { return &s }

func runnerHealthCases() []runnerHealthCase {
	offlineFresh := corpusNow.Add(-15*time.Minute + time.Second).Unix()
	offlineExact := corpusNow.Add(-15 * time.Minute).Unix()
	offlineStale := corpusNow.Add(-time.Hour).Unix()

	healthyOrg := runnerHealthWant{result: "healthy", reason: "runner healthy"}
	unhealthyOrgReason := "no active build process"
	unhealthyInstallReason := "no active install process"

	cases := []runnerHealthCase{
		{
			name: "unknown group type does nothing", groupType: "", status: app.RunnerStatusActive, v2Status: app.RunnerStatusActive,
			want: runnerHealthWant{},
		},
		{
			name: "org healthy already active writes nothing", groupType: app.RunnerGroupTypeOrg,
			status: app.RunnerStatusActive, v2Status: app.RunnerStatusActive, activeBuild: true,
			want: healthyOrg,
		},
		{
			name: "org healthy clears stale offline_ts without resetting", groupType: app.RunnerGroupTypeOrg,
			status: app.RunnerStatusActive, v2Status: app.RunnerStatusActive, activeBuild: true,
			metadata: map[string]any{app.RunnerOfflineTSMetadataKey: offlineStale},
			want:     runnerHealthWant{result: "healthy", reason: "runner healthy", clearOfflineTS: true},
		},
		{
			name: "org recovery flips both statuses and clears offline_ts", groupType: app.RunnerGroupTypeOrg,
			status: app.RunnerStatusOffline, v2Status: app.RunnerStatusOffline, activeBuild: true,
			metadata: map[string]any{app.RunnerOfflineTSMetadataKey: offlineStale},
			want: runnerHealthWant{
				result: "healthy", reason: "runner healthy", clearOfflineTS: true,
				legacyStatus: runnerStatusPtr(app.RunnerStatusActive), v2Status: runnerStatusPtr(app.RunnerStatusActive),
			},
		},
		{
			name: "org first failed check arms and transitions without alert", groupType: app.RunnerGroupTypeOrg,
			status: app.RunnerStatusActive, v2Status: app.RunnerStatusActive,
			want: runnerHealthWant{
				result: "unhealthy", reason: unhealthyOrgReason, armOfflineTS: true,
				legacyStatus: runnerStatusPtr(app.RunnerStatusOffline), v2Status: runnerStatusPtr(app.RunnerStatusOffline),
			},
		},
		{
			name: "org unhealthy with existing offline_ts still transitioning does not re-arm", groupType: app.RunnerGroupTypeOrg,
			status: app.RunnerStatusActive, v2Status: app.RunnerStatusActive,
			metadata: map[string]any{app.RunnerOfflineTSMetadataKey: offlineStale},
			want: runnerHealthWant{
				result: "unhealthy", reason: unhealthyOrgReason,
				legacyStatus: runnerStatusPtr(app.RunnerStatusOffline), v2Status: runnerStatusPtr(app.RunnerStatusOffline),
			},
		},
		{
			name: "org offline under alert delay does nothing", groupType: app.RunnerGroupTypeOrg,
			status: app.RunnerStatusOffline, v2Status: app.RunnerStatusOffline,
			metadata: map[string]any{app.RunnerOfflineTSMetadataKey: offlineFresh},
			want:     runnerHealthWant{result: "unhealthy", reason: unhealthyOrgReason},
		},
		{
			name: "org offline exactly at alert delay alerts", groupType: app.RunnerGroupTypeOrg,
			status: app.RunnerStatusOffline, v2Status: app.RunnerStatusOffline,
			metadata: map[string]any{app.RunnerOfflineTSMetadataKey: offlineExact},
			want:     runnerHealthWant{result: "unhealthy", reason: unhealthyOrgReason, alert: true, alertOfflineAt: offlineExact},
		},
		{
			name: "org offline past alert delay alerts with persisted offlineAt", groupType: app.RunnerGroupTypeOrg,
			status: app.RunnerStatusOffline, v2Status: app.RunnerStatusOffline,
			metadata: map[string]any{app.RunnerOfflineTSMetadataKey: offlineStale},
			want:     runnerHealthWant{result: "unhealthy", reason: unhealthyOrgReason, alert: true, alertOfflineAt: offlineStale},
		},
		{
			name: "org offline without timestamp arms only", groupType: app.RunnerGroupTypeOrg,
			status: app.RunnerStatusOffline, v2Status: app.RunnerStatusOffline,
			want: runnerHealthWant{result: "unhealthy", reason: unhealthyOrgReason, armOfflineTS: true},
		},
		{
			name: "legacy offline but v2 stale repairs v2 only", groupType: app.RunnerGroupTypeOrg,
			status: app.RunnerStatusOffline, v2Status: app.RunnerStatusActive,
			metadata: map[string]any{app.RunnerOfflineTSMetadataKey: offlineStale},
			want: runnerHealthWant{
				result: "unhealthy", reason: unhealthyOrgReason,
				v2Status: runnerStatusPtr(app.RunnerStatusOffline),
			},
		},
		{
			name: "v2 offline but legacy stale repairs legacy only", groupType: app.RunnerGroupTypeOrg,
			status: app.RunnerStatusActive, v2Status: app.RunnerStatusOffline,
			metadata: map[string]any{app.RunnerOfflineTSMetadataKey: offlineStale},
			want: runnerHealthWant{
				result: "unhealthy", reason: unhealthyOrgReason,
				legacyStatus: runnerStatusPtr(app.RunnerStatusOffline),
			},
		},
		{
			name: "install healthy with mng writes missing_mng=false when metadata absent", groupType: app.RunnerGroupTypeInstall,
			status: app.RunnerStatusActive, v2Status: app.RunnerStatusActive,
			activeInstall: true, activeMng: true, mngChecked: true,
			want: runnerHealthWant{result: "healthy", reason: "runner healthy", setMissingMng: generics.ToPtr(false)},
		},
		{
			name: "install healthy mng metadata unchanged writes nothing", groupType: app.RunnerGroupTypeInstall,
			status: app.RunnerStatusActive, v2Status: app.RunnerStatusActive,
			metadata:      map[string]any{"missing_mng_process": false},
			activeInstall: true, activeMng: true, mngChecked: true,
			want: runnerHealthWant{result: "healthy", reason: "runner healthy"},
		},
		{
			name: "install healthy mng went missing writes flip", groupType: app.RunnerGroupTypeInstall,
			status: app.RunnerStatusActive, v2Status: app.RunnerStatusActive,
			metadata:      map[string]any{"missing_mng_process": false},
			activeInstall: true, activeMng: false, mngChecked: true,
			want: runnerHealthWant{result: "healthy", reason: "runner healthy", setMissingMng: generics.ToPtr(true)},
		},
		{
			name: "install healthy non-bool mng metadata is rewritten", groupType: app.RunnerGroupTypeInstall,
			status: app.RunnerStatusActive, v2Status: app.RunnerStatusActive,
			metadata:      map[string]any{"missing_mng_process": "yes"},
			activeInstall: true, activeMng: true, mngChecked: true,
			want: runnerHealthWant{result: "healthy", reason: "runner healthy", setMissingMng: generics.ToPtr(false)},
		},
		{
			name: "install mng unchecked never writes mng metadata", groupType: app.RunnerGroupTypeInstall,
			status: app.RunnerStatusActive, v2Status: app.RunnerStatusActive,
			activeInstall: true, activeMng: false, mngChecked: false,
			want: runnerHealthWant{result: "healthy", reason: "runner healthy"},
		},
		{
			name: "install missing install process is unhealthy", groupType: app.RunnerGroupTypeInstall,
			status: app.RunnerStatusActive, v2Status: app.RunnerStatusActive,
			activeMng: true, mngChecked: true,
			want: runnerHealthWant{
				result: "unhealthy", reason: unhealthyInstallReason, armOfflineTS: true,
				setMissingMng: generics.ToPtr(false),
				legacyStatus:  runnerStatusPtr(app.RunnerStatusOffline), v2Status: runnerStatusPtr(app.RunnerStatusOffline),
			},
		},
		{
			name: "install offline past delay alerts with install reason", groupType: app.RunnerGroupTypeInstall,
			status: app.RunnerStatusOffline, v2Status: app.RunnerStatusOffline,
			metadata:  map[string]any{app.RunnerOfflineTSMetadataKey: offlineStale, "missing_mng_process": true},
			activeMng: false, mngChecked: true,
			want: runnerHealthWant{result: "unhealthy", reason: unhealthyInstallReason, alert: true, alertOfflineAt: offlineStale},
		},
	}

	// A disabled runner has no processes at all, which must not arm offline_ts
	// or raise an unhealthy alert the way a genuinely silent runner does.
	cases = append(cases, runnerHealthCase{
		name:      "disabled install runner with no processes is skipped",
		groupType: app.RunnerGroupTypeInstall,
		status:    app.RunnerStatusDisabled, v2Status: app.RunnerStatusDisabled,
		mngChecked: true,
		want:       runnerHealthWant{result: "skipped"},
	})

	for _, status := range []app.RunnerStatus{
		app.RunnerStatusProvisioning,
		app.RunnerStatusDeprovisioning,
		app.RunnerStatusReprovisioning,
		app.RunnerStatusDeprovisioned,
		app.RunnerStatusPending,
		app.RunnerStatusDisabled,
	} {
		cases = append(cases, runnerHealthCase{
			name:      "skippable status " + string(status),
			groupType: app.RunnerGroupTypeOrg,
			status:    status, v2Status: status,
			activeBuild: true,
			want:        runnerHealthWant{result: "skipped"},
		})
	}

	return cases
}

type processHealthCase struct {
	name              string
	status            app.RunnerProcessStatus
	shutdownRequested any
	hasShutdownKey    bool
	heartbeatAge      *time.Duration
	want              processHealthAction
}

func processAge(d time.Duration) *time.Duration { return &d }

func processHealthCases() []processHealthCase {
	cases := []processHealthCase{
		{name: "shutdown requested wins over stale heartbeat", status: app.RunnerProcessStatusActive,
			shutdownRequested: true, hasShutdownKey: true, heartbeatAge: processAge(6 * time.Minute), want: processActionShutdown},
		{name: "shutdown requested wins over fresh heartbeat on offline process", status: app.RunnerProcessStatusOffline,
			shutdownRequested: true, hasShutdownKey: true, heartbeatAge: processAge(10 * time.Second), want: processActionShutdown},
		{name: "nil shutdown_requested value is ignored", status: app.RunnerProcessStatusActive,
			shutdownRequested: nil, hasShutdownKey: true, heartbeatAge: processAge(10 * time.Second), want: processActionActive},
		{name: "no heartbeat is noop", status: app.RunnerProcessStatusActive, want: processActionNoop},
		{name: "no heartbeat on offline process is noop", status: app.RunnerProcessStatusOffline, want: processActionNoop},
		{name: "10s heartbeat is active", status: app.RunnerProcessStatusActive, heartbeatAge: processAge(10 * time.Second), want: processActionActive},
		{name: "59s heartbeat is active", status: app.RunnerProcessStatusActive, heartbeatAge: processAge(59 * time.Second), want: processActionActive},
		{name: "60s heartbeat is offline", status: app.RunnerProcessStatusActive, heartbeatAge: processAge(60 * time.Second), want: processActionOffline},
		{name: "61s heartbeat is offline", status: app.RunnerProcessStatusActive, heartbeatAge: processAge(61 * time.Second), want: processActionOffline},
		{name: "299s heartbeat is offline", status: app.RunnerProcessStatusActive, heartbeatAge: processAge(299 * time.Second), want: processActionOffline},
		{name: "300s heartbeat is inactive", status: app.RunnerProcessStatusActive, heartbeatAge: processAge(300 * time.Second), want: processActionInactive},
		{name: "301s heartbeat is inactive", status: app.RunnerProcessStatusActive, heartbeatAge: processAge(301 * time.Second), want: processActionInactive},
		{name: "6m heartbeat is inactive", status: app.RunnerProcessStatusActive, heartbeatAge: processAge(6 * time.Minute), want: processActionInactive},
		{name: "offline process with fresh heartbeat flips active", status: app.RunnerProcessStatusOffline, heartbeatAge: processAge(10 * time.Second), want: processActionActive},
		{name: "offline process still silent stays offline action", status: app.RunnerProcessStatusOffline, heartbeatAge: processAge(2 * time.Minute), want: processActionOffline},
	}

	for _, status := range []app.RunnerProcessStatus{
		app.RunnerProcessStatusInactive,
		app.RunnerProcessStatusPendingShutdown,
		app.RunnerProcessStatusShuttingDown,
		app.RunnerProcessStatusShutDown,
		app.RunnerProcessStatusError,
		app.RunnerProcessStatusUnknown,
	} {
		cases = append(cases, processHealthCase{
			name:   "status " + string(status) + " is noop",
			status: status, heartbeatAge: processAge(10 * time.Second),
			want: processActionNoop,
		})
	}

	return cases
}

type versionWarningCase struct {
	name            string
	configured      string
	reported        string
	wantWarning     bool
	wantLatestEvent bool
}

func versionWarningCases() []versionWarningCase {
	return []versionWarningCase{
		{name: "versions match", configured: "1.2.3", reported: "1.2.3"},
		{name: "empty reported version", configured: "1.2.3", reported: ""},
		{name: "empty configured version", configured: "", reported: "1.2.3"},
		{name: "cloud tracks api version", configured: "cloud", reported: "1.2.3"},
		{name: "alias tag warns", configured: "stable", reported: "1.2.3", wantWarning: true},
		{name: "semver mismatch warns", configured: "1.2.4", reported: "1.2.3", wantWarning: true},
		{name: "latest configured warns and emits event", configured: "latest", reported: "1.2.3", wantWarning: true, wantLatestEvent: true},
		{name: "latest reported warns and emits event", configured: "1.2.3", reported: "latest", wantWarning: true, wantLatestEvent: true},
	}
}
