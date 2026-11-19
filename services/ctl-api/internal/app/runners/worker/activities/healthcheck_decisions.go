package activities

import (
	"fmt"
	"time"

	"github.com/Masterminds/semver/v3"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

const (
	processOfflineTimeout       = 1 * time.Minute
	processInactiveTimeout      = 5 * time.Minute
	processFirstHeartbeatGrace  = 10 * time.Minute
	processShutdownStaleTimeout = 1 * time.Hour

	runnerUnhealthyAlertDelay = 15 * time.Minute

	runnerHealthResultSkipped   = "skipped"
	runnerHealthResultHealthy   = "healthy"
	runnerHealthResultUnhealthy = "unhealthy"
)

// skippableRunnerStatuses mirrors the statuses the runner healthcheck never
// acts on (provisioning lifecycle states, plus runners the install stack
// disabled or that intentionally run without a runner, which never report
// health by design).
var skippableRunnerStatuses = []app.RunnerStatus{
	app.RunnerStatusProvisioning,
	app.RunnerStatusDeprovisioning,
	app.RunnerStatusReprovisioning,
	app.RunnerStatusDeprovisioned,
	app.RunnerStatusPending,
	app.RunnerStatusDisabled,
	app.RunnerStatusNotUsed,
}

func isSkippableRunnerStatus(status app.RunnerStatus) bool {
	for _, s := range skippableRunnerStatuses {
		if s == status {
			return true
		}
	}
	return false
}

type runnerProcessPresence struct {
	HasActiveInstall bool
	HasActiveMng     bool
	MngChecked       bool
}

type runnerHealthDecision struct {
	Result       string
	TargetStatus app.RunnerStatus
	Reason       string

	SetMissingMng *bool

	ClearOfflineTS bool
	SetOfflineTS   bool

	UpdateLegacy bool
	UpdateV2     bool

	Alert          bool
	AlertOfflineAt time.Time

	InstallCronToggleDecision *InstallCronState
}

// decideRunnerHealth encodes runnerhealthcheck.Signal.Execute's branch logic:
// arm offline_ts before status writes, alert only on a tick after the offline
// transition once the delay has elapsed, guard every write on current state.
func decideRunnerHealth(now time.Time, runner *app.Runner, presence runnerProcessPresence) runnerHealthDecision {
	var d runnerHealthDecision

	if runner.Status == app.RunnerStatusDisabled || runner.Status == app.RunnerStatusNotUsed {
		d.Result = runnerHealthResultSkipped
		if runner.RunnerGroup.Type == app.RunnerGroupTypeInstall {
			state := InstallCronsDisabled
			d.InstallCronToggleDecision = &state
		}
		return d
	}

	if isSkippableRunnerStatus(runner.Status) {
		d.Result = runnerHealthResultSkipped
		return d
	}

	var healthy bool
	switch runner.RunnerGroup.Type {
	case app.RunnerGroupTypeInstall:
		healthy = presence.HasActiveInstall
		d.Reason = "no active install process"
		if presence.MngChecked {
			missing := !presence.HasActiveMng
			current, has := runner.StatusV2.Metadata["missing_mng_process"].(bool)
			if !has || current != missing {
				d.SetMissingMng = &missing
			}
		}
	default:
		return d
	}

	if !healthy && runner.HealthcheckPending(now) {
		return runnerHealthDecision{Result: runnerHealthResultSkipped}
	}

	if healthy {
		d.Result = runnerHealthResultHealthy
		d.TargetStatus = app.RunnerStatusActive
		d.Reason = "runner healthy"
		_, hasOfflineTS := runner.StatusV2.Metadata[app.RunnerOfflineTSMetadataKey]
		d.ClearOfflineTS = hasOfflineTS
		d.UpdateLegacy = runner.Status != app.RunnerStatusActive
		d.UpdateV2 = runner.StatusV2.Status != app.Status(app.RunnerStatusActive)
		if runner.RunnerGroup.Type == app.RunnerGroupTypeInstall {
			state := InstallCronsEnabled
			d.InstallCronToggleDecision = &state
		}
		return d
	}

	d.Result = runnerHealthResultUnhealthy
	d.TargetStatus = app.RunnerStatusOffline

	offlineAt, hasOfflineTS := runner.StatusV2.MetadataUnixTime(app.RunnerOfflineTSMetadataKey)
	if !hasOfflineTS {
		d.SetOfflineTS = true
		offlineAt = now
	}

	if runner.Status != app.RunnerStatusOffline || runner.StatusV2.Status != app.Status(app.RunnerStatusOffline) {
		d.UpdateLegacy = runner.Status != app.RunnerStatusOffline
		d.UpdateV2 = runner.StatusV2.Status != app.Status(app.RunnerStatusOffline)
		return d
	}

	if now.Sub(offlineAt) < runnerUnhealthyAlertDelay {
		return d
	}

	d.Alert = true
	d.AlertOfflineAt = offlineAt
	if runner.RunnerGroup.Type == app.RunnerGroupTypeInstall {
		state := InstallCronsDisabled
		d.InstallCronToggleDecision = &state
	}
	return d
}

type processHealthAction int

const (
	processActionNoop processHealthAction = iota
	processActionShutdown
	processActionInactive
	processActionOffline
	processActionActive
)

// decideProcessHealth differs from processhealthcheck.Signal.Execute in two
// ways: a process that stopped heartbeating goes inactive before a pending
// shutdown_requested is honoured, since a dead process never acts on it, and a
// process stuck mid-shutdown goes inactive after processShutdownStaleTimeout
// without a heartbeat.
func decideProcessHealth(now time.Time, process *app.RunnerProcess, heartbeatAt *time.Time) processHealthAction {
	switch process.ProcessStatus() {
	case app.RunnerProcessStatusActive, app.RunnerProcessStatusOffline:
	case app.RunnerProcessStatusPendingShutdown, app.RunnerProcessStatusShuttingDown:
		if heartbeatStale(now, process.StartedAt, heartbeatAt, processShutdownStaleTimeout, processShutdownStaleTimeout) {
			return processActionInactive
		}
		return processActionNoop
	default:
		return processActionNoop
	}

	if heartbeatStale(now, process.StartedAt, heartbeatAt, processFirstHeartbeatGrace, processInactiveTimeout) {
		return processActionInactive
	}

	if val, ok := process.CompositeStatus.Metadata["shutdown_requested"]; ok && val != nil {
		return processActionShutdown
	}

	if heartbeatAt == nil {
		return processActionNoop
	}

	if now.Sub(*heartbeatAt) >= processOfflineTimeout {
		return processActionOffline
	}
	return processActionActive
}

func isShuttingDownProcessStatus(status app.RunnerProcessStatus) bool {
	return status == app.RunnerProcessStatusPendingShutdown || status == app.RunnerProcessStatusShuttingDown
}

func heartbeatStale(now time.Time, startedAt, heartbeatAt *time.Time, firstHeartbeatGrace, timeout time.Duration) bool {
	if heartbeatAt == nil {
		return startedAt != nil && now.Sub(*startedAt) >= firstHeartbeatGrace
	}
	return now.Sub(*heartbeatAt) >= timeout
}

// decideVersionWarning lifts processhealthcheck.Signal.checkVersionMismatch's
// warning derivation. emitLatestEvent flags the 'latest' tag Datadog event.
func decideVersionWarning(configured, reported string) (warning string, emitLatestEvent bool) {
	if reported == "" {
		return "", false
	}

	emitLatestEvent = configured == "latest" || reported == "latest"

	isAliasTag := configured != "" && func() bool {
		_, err := semver.NewVersion(configured)
		return err != nil
	}()

	switch {
	case configured == "" || configured == reported:
	case configured == "cloud":
	case isAliasTag:
		warning = fmt.Sprintf(
			"Runner is configured with alias tag (%s). We recommend pinning a specific version to avoid drift.",
			configured,
		)
	default:
		warning = fmt.Sprintf(
			"Reported runner version (%s) does not match configured version (%s). Please update the runner to the correct version.",
			reported, configured,
		)
	}
	return warning, emitLatestEvent
}
