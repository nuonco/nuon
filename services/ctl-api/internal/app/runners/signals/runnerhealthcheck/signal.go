package runnerhealthcheck

import (
	"fmt"
	"time"

	"github.com/DataDog/datadog-go/v5/statsd"
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/metrics"
	tmetrics "github.com/nuonco/nuon/pkg/temporal/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/signals/runnerunhealthy"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

const (
	SignalType                signal.SignalType = "runner_healthcheck"
	runnerUnhealthyAlertDelay                   = 15 * time.Minute
)

type Signal struct {
	RunnerID string `json:"runner_id"`

	mw metrics.Writer
	v  *validator.Validate
}

var (
	_ signal.Signal                   = (*Signal)(nil)
	_ signal.SleepAfter               = (*Signal)(nil)
	_ signal.SignalWithMaxInFlightAge = (*Signal)(nil)
	_ signal.SignalWithParams         = (*Signal)(nil)
)

func (s *Signal) WithParams(params *signal.Params) {
	s.mw = params.MW
	s.v = params.V
}

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) SleepAfter() time.Duration {
	return 0
}

func (s *Signal) MaxInFlightAge() time.Duration {
	return 10 * time.Minute
}

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.RunnerID == "" {
		return errors.New("runner_id is required")
	}
	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	tmw, err := tmetrics.New(s.v, tmetrics.WithMetricsWriter(s.mw))
	if err != nil {
		return errors.Wrap(err, "unable to create temporal metrics writer")
	}

	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return errors.Wrap(err, "unable to get logger")
	}

	runner, err := activities.LocalAwaitGet(ctx, activities.GetRequest{RunnerID: s.RunnerID})
	if err != nil {
		return errors.Wrap(err, "unable to get runner")
	}

	tags := map[string]string{
		"runner_id":     s.RunnerID,
		"runner_type":   string(runner.RunnerGroup.Type),
		"runner_status": string(runner.Status),
		"org_id":        runner.OrgID,
		"org_name":      runner.Org.Name,
	}

	if runner.RunnerGroup.OwnerType == "installs" {
		tags["install_id"] = runner.RunnerGroup.OwnerID
		install, err := activities.LocalAwaitGetRunnerInstallByInstallID(ctx, runner.RunnerGroup.OwnerID)
		if err != nil {
			l.Warn("unable to add install name to runner health check metric", zap.Error(err))
		} else {
			tags["install_name"] = install.Name
		}
	}

	if isSkippableStatus(runner.Status) {
		tmw.Incr(ctx, "runner.health_check", metrics.ToTags(tags, metrics.ToTag("result", "skipped"))...)
		return nil
	}

	switch runner.RunnerGroup.Type {
	case app.RunnerGroupTypeOrg:
		return s.checkOrgRunner(ctx, l, tmw, runner, tags)
	case app.RunnerGroupTypeInstall:
		return s.checkInstallRunner(ctx, l, tmw, runner, tags)
	default:
		return nil
	}
}

func (s *Signal) checkOrgRunner(ctx workflow.Context, l *zap.Logger, tmw tmetrics.Writer, runner *app.Runner, tags map[string]string) error {
	_, err := activities.LocalAwaitGetCurrentRunnerProcess(ctx, activities.GetCurrentRunnerProcessRequest{
		RunnerID:    s.RunnerID,
		ProcessType: string(app.RunnerProcessTypeBuild),
	})

	tags["missing_build_process"] = "false"
	if err != nil {
		if isNotFound(err) {
			l.Warn("org runner has no active build process",
				zap.String("runner_id", s.RunnerID),
			)
			tags["missing_build_process"] = "true"
			tmw.Incr(ctx, "runner.health_check", metrics.ToTags(tags, metrics.ToTag("result", "unhealthy"))...)
			return s.handleRunnerOffline(ctx, tmw, runner, "no active build process")
		}
		return errors.Wrap(err, "unable to get current build process")
	}

	tmw.Incr(ctx, "runner.health_check", metrics.ToTags(tags, metrics.ToTag("result", "healthy"))...)
	return s.handleRunnerActive(ctx, runner)
}

func (s *Signal) checkInstallRunner(ctx workflow.Context, l *zap.Logger, tmw tmetrics.Writer, runner *app.Runner, tags map[string]string) error {
	_, err := activities.LocalAwaitGetCurrentRunnerProcess(ctx, activities.GetCurrentRunnerProcessRequest{
		RunnerID:    s.RunnerID,
		ProcessType: string(app.RunnerProcessTypeInstall),
	})

	var (
		status      app.RunnerStatus
		description string
	)

	tags["missing_install_process"] = "false"
	if err != nil {
		if isNotFound(err) {
			l.Warn("install runner has no active install process",
				zap.String("runner_id", s.RunnerID),
			)
			tags["missing_install_process"] = "true"
			status = app.RunnerStatusOffline
			description = "no active install process"
		} else {
			return errors.Wrap(err, "unable to get current install process")
		}
	} else {
		status = app.RunnerStatusActive
		description = "runner healthy"
	}

	missingMngProcess := false
	mngProcessChecked := false
	_, mngErr := activities.LocalAwaitGetCurrentRunnerProcess(ctx, activities.GetCurrentRunnerProcessRequest{
		RunnerID:    s.RunnerID,
		ProcessType: string(app.RunnerProcessTypeMng),
	})

	tags["missing_mng_process"] = "false"
	if mngErr != nil {
		if isNotFound(mngErr) {
			mngProcessChecked = true
			missingMngProcess = true
			l.Warn("install runner missing management process",
				zap.String("runner_id", s.RunnerID),
			)
			tags["missing_mng_process"] = "true"
		} else {
			l.Warn("unable to check management process",
				zap.String("runner_id", s.RunnerID),
				zap.Error(mngErr),
			)
		}
	} else {
		mngProcessChecked = true
	}

	currentMissingMngProcess, hasMissingMngProcess := runner.StatusV2.Metadata["missing_mng_process"].(bool)
	if mngProcessChecked && (!hasMissingMngProcess || currentMissingMngProcess != missingMngProcess) {
		if err := statusactivities.LocalAwaitUpdateRunnerStatusV2Metadata(ctx, statusactivities.UpdateRunnerStatusV2MetadataRequest{
			RunnerID: s.RunnerID,
			Metadata: map[string]any{"missing_mng_process": missingMngProcess},
		}); err != nil {
			return errors.Wrap(err, "unable to update management process status metadata")
		}
	}

	if status == app.RunnerStatusActive {
		tmw.Incr(ctx, "runner.health_check", metrics.ToTags(tags, metrics.ToTag("result", "healthy"))...)
		return s.handleRunnerActive(ctx, runner)
	}

	tmw.Incr(ctx, "runner.health_check", metrics.ToTags(tags, metrics.ToTag("result", "unhealthy"))...)
	return s.handleRunnerOffline(ctx, tmw, runner, description)
}

func (s *Signal) handleRunnerActive(ctx workflow.Context, runner *app.Runner) error {
	if _, ok := runner.StatusV2.Metadata[app.RunnerOfflineTSMetadataKey]; ok {
		if err := statusactivities.LocalAwaitUpdateRunnerStatusV2Metadata(ctx, statusactivities.UpdateRunnerStatusV2MetadataRequest{
			RunnerID: s.RunnerID,
			Metadata: map[string]any{
				app.RunnerOfflineTSMetadataKey: nil,
			},
		}); err != nil {
			return errors.Wrap(err, "unable to clear runner offline metadata")
		}
	}

	return s.updateRunnerStatus(ctx, runner, app.RunnerStatusActive, "runner healthy")
}

func (s *Signal) handleRunnerOffline(ctx workflow.Context, tmw tmetrics.Writer, runner *app.Runner, reason string) error {
	now := workflow.Now(ctx)
	offlineAt, hasOfflineTS := runner.StatusV2.MetadataUnixTime(app.RunnerOfflineTSMetadataKey)

	if !hasOfflineTS {
		if err := statusactivities.LocalAwaitUpdateRunnerStatusV2Metadata(ctx, statusactivities.UpdateRunnerStatusV2MetadataRequest{
			RunnerID: s.RunnerID,
			Metadata: map[string]any{
				app.RunnerOfflineTSMetadataKey: now.Unix(),
			},
		}); err != nil {
			return errors.Wrap(err, "unable to set runner offline metadata")
		}
		offlineAt = now
	}

	if runner.Status != app.RunnerStatusOffline || runner.StatusV2.Status != app.Status(app.RunnerStatusOffline) {
		return s.updateRunnerStatus(ctx, runner, app.RunnerStatusOffline, reason)
	}

	if now.Sub(offlineAt) < runnerUnhealthyAlertDelay {
		return nil
	}

	if err := s.notifyRunnerUnhealthy(ctx, tmw, runner, reason, offlineAt); err != nil {
		return err
	}

	return nil
}

func (s *Signal) runnerOfflineEvent(ctx workflow.Context, runner *app.Runner, reason string) (*statsd.Event, string) {
	eventTags := []string{
		metrics.ToTag("runner_id", s.RunnerID),
		metrics.ToTag("runner_type", string(runner.RunnerGroup.Type)),
		metrics.ToTag("org_id", runner.OrgID),
		metrics.ToTag("org_name", runner.Org.Name),
	}

	title := fmt.Sprintf("Runner went offline (type: %s)", string(runner.RunnerGroup.Type))
	text := fmt.Sprintf(
		"Runner %s (org: %s) transitioned from active to offline.\nReason: %s",
		s.RunnerID, runner.Org.Name, reason,
	)
	ownerName := ""
	if runner.RunnerGroup.OwnerType == "orgs" {
		ownerName = runner.Org.Name
	}

	if runner.RunnerGroup.OwnerType == "installs" {
		install, err := activities.LocalAwaitGetInstall(ctx, activities.GetInstallRequest{
			InstallID: runner.RunnerGroup.OwnerID,
		})
		if err == nil {
			ownerName = install.Name
			eventTags = append(eventTags,
				metrics.ToTag("install_id", install.ID),
				metrics.ToTag("install_name", install.Name),
				metrics.ToTag("app_id", install.AppID),
				metrics.ToTag("app_name", install.App.Name),
				metrics.ToTag("created_by", install.CreatedBy.Email),
			)
			text = fmt.Sprintf(
				"Runner %s (org: %s, app: %s, install: %s) transitioned from active to offline.\nReason: %s\nInstall created by: %s",
				s.RunnerID, runner.Org.Name, install.App.Name, install.Name, reason, install.CreatedBy.Email,
			)
		}
	}

	return &statsd.Event{
		Title:          title,
		Text:           text,
		Tags:           eventTags,
		SourceTypeName: "nuon-runner",
		Priority:       statsd.Normal,
		AlertType:      statsd.Error,
		AggregationKey: "runner-health-check",
	}, ownerName
}

func (s *Signal) notifyRunnerUnhealthy(ctx workflow.Context, tmw tmetrics.Writer, runner *app.Runner, reason string, offlineAt time.Time) error {
	event, ownerName := s.runnerOfflineEvent(ctx, runner, reason)
	resp, err := sharedactivities.AwaitEnqueueSignalToOwner(ctx, &sharedactivities.EnqueueSignalToOwnerRequest{
		OwnerID:         runner.OrgID,
		OwnerType:       "orgs",
		QueueName:       "org-signals",
		SignalOwnerID:   runner.ID,
		SignalOwnerType: "runners",
		IdempotencyKey:  fmt.Sprintf("runner-unhealthy:%s:%d", runner.ID, offlineAt.Unix()),
		Signal: &runnerunhealthy.Signal{
			RunnerID:             runner.ID,
			RunnerName:           runner.DisplayName,
			OrgID:                runner.OrgID,
			OrgName:              runner.Org.Name,
			FromStatus:           app.RunnerStatusActive,
			ToStatus:             app.RunnerStatusOffline,
			Reason:               reason,
			RunnerGroupID:        runner.RunnerGroupID,
			RunnerGroupType:      runner.RunnerGroup.Type,
			RunnerGroupOwnerID:   runner.RunnerGroup.OwnerID,
			RunnerGroupOwnerType: runner.RunnerGroup.OwnerType,
			RunnerGroupOwnerName: ownerName,
		},
	})
	if err != nil {
		return errors.Wrap(err, "unable to enqueue runner unhealthy signal")
	}
	if !resp.Deduplicated {
		tmw.Event(ctx, event)
	}
	return nil
}

func (s *Signal) updateRunnerStatus(ctx workflow.Context, runner *app.Runner, status app.RunnerStatus, description string) error {
	if runner.Status != status {
		if err := activities.LocalAwaitUpdateStatus(ctx, activities.UpdateStatusRequest{
			RunnerID:          s.RunnerID,
			Status:            status,
			StatusDescription: description,
		}); err != nil {
			return errors.Wrap(err, "unable to update runner status")
		}
	}
	if runner.StatusV2.Status == app.Status(status) {
		return nil
	}

	if err := statusactivities.LocalAwaitUpdateRunnerStatusV2(ctx, statusactivities.UpdateRunnerStatusV2Request{
		RunnerID:          s.RunnerID,
		Status:            status,
		StatusDescription: description,
	}); err != nil {
		return errors.Wrap(err, "unable to update runner status v2")
	}

	return nil
}

func isNotFound(err error) bool {
	var appErr *temporal.ApplicationError
	if errors.As(err, &appErr) && appErr.NonRetryable() {
		return true
	}
	return false
}

func isSkippableStatus(status app.RunnerStatus) bool {
	switch status {
	case app.RunnerStatusProvisioning,
		app.RunnerStatusDeprovisioning,
		app.RunnerStatusReprovisioning,
		app.RunnerStatusDeprovisioned,
		app.RunnerStatusPending:
		return true
	}
	return false
}
