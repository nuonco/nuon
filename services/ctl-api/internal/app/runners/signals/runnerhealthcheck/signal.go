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
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

const SignalType signal.SignalType = "runner_healthcheck"

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

	runner, err := activities.AwaitGet(ctx, activities.GetRequest{RunnerID: s.RunnerID})
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
	_, err := activities.AwaitGetCurrentRunnerProcess(ctx, activities.GetCurrentRunnerProcessRequest{
		RunnerID:    s.RunnerID,
		ProcessType: string(app.RunnerProcessTypeOrg),
	})

	tags["missing_org_process"] = "false"
	if err != nil {
		if isNotFound(err) {
			l.Warn("org runner has no active org process",
				zap.String("runner_id", s.RunnerID),
			)
			tags["missing_org_process"] = "true"
			tmw.Incr(ctx, "runner.health_check", metrics.ToTags(tags, metrics.ToTag("result", "unhealthy"))...)
			if runner.Status == app.RunnerStatusActive {
				s.emitOfflineEvent(ctx, runner, "no active org process")
			}
			return s.updateRunnerStatus(ctx, runner, app.RunnerStatusOffline, "no active org process", nil)
		}
		return errors.Wrap(err, "unable to get current org process")
	}

	tmw.Incr(ctx, "runner.health_check", metrics.ToTags(tags, metrics.ToTag("result", "healthy"))...)
	return s.updateRunnerStatus(ctx, runner, app.RunnerStatusActive, "runner healthy", nil)
}

func (s *Signal) checkInstallRunner(ctx workflow.Context, l *zap.Logger, tmw tmetrics.Writer, runner *app.Runner, tags map[string]string) error {
	_, err := activities.AwaitGetCurrentRunnerProcess(ctx, activities.GetCurrentRunnerProcessRequest{
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

	var metadata map[string]any
	_, mngErr := activities.AwaitGetCurrentRunnerProcess(ctx, activities.GetCurrentRunnerProcessRequest{
		RunnerID:    s.RunnerID,
		ProcessType: string(app.RunnerProcessTypeMng),
	})

	tags["missing_mng_process"] = "false"
	if mngErr != nil {
		if isNotFound(mngErr) {
			l.Warn("install runner missing management process",
				zap.String("runner_id", s.RunnerID),
			)
			tags["missing_mng_process"] = "true"
			metadata = map[string]any{"missing_mng_process": true}
		} else {
			l.Warn("unable to check management process",
				zap.String("runner_id", s.RunnerID),
				zap.Error(mngErr),
			)
		}
	} else {
		metadata = map[string]any{"missing_mng_process": false}
	}

	if status == app.RunnerStatusActive {
		tmw.Incr(ctx, "runner.health_check", metrics.ToTags(tags, metrics.ToTag("result", "healthy"))...)
	} else {
		tmw.Incr(ctx, "runner.health_check", metrics.ToTags(tags, metrics.ToTag("result", "unhealthy"))...)
		if runner.Status == app.RunnerStatusActive {
			s.emitOfflineEvent(ctx, runner, description)
		}
	}

	return s.updateRunnerStatus(ctx, runner, status, description, metadata)
}

func (s *Signal) emitOfflineEvent(ctx workflow.Context, runner *app.Runner, reason string) {
	eventTags := []string{
		metrics.ToTag("runner_id", s.RunnerID),
		metrics.ToTag("runner_type", string(runner.RunnerGroup.Type)),
		metrics.ToTag("org_id", runner.OrgID),
		metrics.ToTag("org_name", runner.Org.Name),
	}

	title := fmt.Sprintf("Runner %s went offline", runner.DisplayName)
	text := fmt.Sprintf(
		"Runner %s (org: %s) transitioned from active to offline.\nReason: %s",
		s.RunnerID, runner.Org.Name, reason,
	)

	if runner.RunnerGroup.OwnerType == "installs" {
		install, err := activities.AwaitGetInstall(ctx, activities.GetInstallRequest{
			InstallID: runner.RunnerGroup.OwnerID,
		})
		if err == nil {
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

	s.mw.Event(&statsd.Event{
		Title:          title,
		Text:           text,
		Tags:           eventTags,
		SourceTypeName: "nuon-runner",
		Priority:       statsd.Normal,
		AlertType:      statsd.Error,
		AggregationKey: "runner-health-check",
	})
}

func (s *Signal) updateRunnerStatus(ctx workflow.Context, runner *app.Runner, status app.RunnerStatus, description string, metadata map[string]any) error {
	if runner.Status != status {
		if err := activities.AwaitUpdateStatus(ctx, activities.UpdateStatusRequest{
			RunnerID:          s.RunnerID,
			Status:            status,
			StatusDescription: description,
		}); err != nil {
			return errors.Wrap(err, "unable to update runner status")
		}
	}

	statusactivities.AwaitUpdateRunnerStatusV2(ctx, statusactivities.UpdateRunnerStatusV2Request{
		RunnerID:          s.RunnerID,
		Status:            status,
		StatusDescription: description,
		Metadata:          metadata,
	})

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
