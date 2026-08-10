package activities

import (
	"fmt"
	"time"

	"github.com/Masterminds/semver/v3"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

const (
	processOfflineTimeout  = 1 * time.Minute
	processInactiveTimeout = 5 * time.Minute

	runnerUnhealthyAlertDelay = 15 * time.Minute
)

// skippableRunnerStatuses mirrors the statuses the runner healthcheck never
// acts on (provisioning lifecycle states).
var skippableRunnerStatuses = []app.RunnerStatus{
	app.RunnerStatusProvisioning,
	app.RunnerStatusDeprovisioning,
	app.RunnerStatusReprovisioning,
	app.RunnerStatusDeprovisioned,
	app.RunnerStatusPending,
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
	HasActiveBuild   bool
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
}

// decideRunnerHealth encodes runnerhealthcheck.Signal.Execute's branch logic:
// arm offline_ts before status writes, alert only on a tick after the offline
// transition once the delay has elapsed, guard every write on current state.
func decideRunnerHealth(now time.Time, runner *app.Runner, presence runnerProcessPresence) runnerHealthDecision {
	var d runnerHealthDecision

	if isSkippableRunnerStatus(runner.Status) {
		d.Result = "skipped"
		return d
	}

	var healthy bool
	switch runner.RunnerGroup.Type {
	case app.RunnerGroupTypeOrg:
		healthy = presence.HasActiveBuild
		d.Reason = "no active build process"
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

	if healthy {
		d.Result = "healthy"
		d.TargetStatus = app.RunnerStatusActive
		d.Reason = "runner healthy"
		_, hasOfflineTS := runner.StatusV2.Metadata[app.RunnerOfflineTSMetadataKey]
		d.ClearOfflineTS = hasOfflineTS
		d.UpdateLegacy = runner.Status != app.RunnerStatusActive
		d.UpdateV2 = runner.StatusV2.Status != app.Status(app.RunnerStatusActive)
		return d
	}

	d.Result = "unhealthy"
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

// decideProcessHealth encodes processhealthcheck.Signal.Execute's branch
// order: status gate, shutdown_requested short-circuit, nil-heartbeat no-op,
// then heartbeat-age tiers.
func decideProcessHealth(now time.Time, process *app.RunnerProcess, heartbeatAt *time.Time) processHealthAction {
	switch process.ProcessStatus() {
	case app.RunnerProcessStatusActive, app.RunnerProcessStatusOffline:
	default:
		return processActionNoop
	}

	if val, ok := process.CompositeStatus.Metadata["shutdown_requested"]; ok && val != nil {
		return processActionShutdown
	}

	if heartbeatAt == nil {
		return processActionNoop
	}

	switch age := now.Sub(*heartbeatAt); {
	case age >= processInactiveTimeout:
		return processActionInactive
	case age >= processOfflineTimeout:
		return processActionOffline
	}
	return processActionActive
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
