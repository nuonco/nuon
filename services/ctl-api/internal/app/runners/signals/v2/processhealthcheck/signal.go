package processhealthcheck

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	installactivities "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/signals/v2/oninactive"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/worker/activities"
	dbgenerics "github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
)

const SignalType signal.SignalType = "process_healthcheck"

const (
	offlineTimeout  = 1 * time.Minute
	inactiveTimeout = 5 * time.Minute
)

type Signal struct {
	RunnerID  string `json:"runner_id"`
	ProcessID string `json:"process_id"`

	mw metrics.Writer
}

var (
	_ signal.Signal           = (*Signal)(nil)
	_ signal.SleepAfter       = (*Signal)(nil)
	_ signal.SignalWithParams = (*Signal)(nil)
)

func (s *Signal) WithParams(params *signal.Params) {
	s.mw = params.MW
}

func (s *Signal) SleepAfter() time.Duration {
	return 0
}

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.RunnerID == "" {
		return errors.New("runner_id is required")
	}
	if s.ProcessID == "" {
		return errors.New("process_id is required")
	}

	return nil
}

func (s *Signal) Execute(ctx workflow.Context) (err error) {
	var startMs int64
	_ = workflow.SideEffect(ctx, func(workflow.Context) interface{} {
		return time.Now().UnixMilli()
	}).Get(&startMs)

	var ownerLabels map[string]string
	var runnerType string
	var runnerStatus string
	var runnerVersion string
	var processType string
	var orgID string
	var installID string
	defer func() {
		status := "ok"
		if err != nil {
			status = "error"
		}
		tagMap := make(map[string]string, len(ownerLabels)+8)
		for k, v := range ownerLabels {
			tagMap[k] = v
		}
		tagMap["status"] = status
		tagMap["runner_id"] = s.RunnerID
		tagMap["runner_type"] = runnerType
		tagMap["runner_status"] = runnerStatus
		tagMap["runner_version"] = runnerVersion
		tagMap["process_type"] = processType
		tagMap["org_id"] = orgID
		if installID != "" {
			tagMap["install_id"] = installID
		}
		tags := metrics.ToTags(tagMap)

		var endMs int64
		_ = workflow.SideEffect(ctx, func(workflow.Context) interface{} {
			return time.Now().UnixMilli()
		}).Get(&endMs)
		latency := time.Duration(endMs-startMs) * time.Millisecond

		s.mw.Timing("runner.health_check.latency", latency, tags)
	}()

	runner, err := activities.AwaitGet(ctx, activities.GetRequest{RunnerID: s.RunnerID})
	if err != nil {
		return errors.Wrap(err, "unable to get runner")
	}
	runnerType = string(runner.RunnerGroup.Type)
	runnerStatus = string(runner.Status)
	orgID = runner.OrgID
	switch runner.RunnerGroup.OwnerType {
	case "installs":
		installID = runner.RunnerGroup.OwnerID
		install, err := installactivities.AwaitGetByInstallID(ctx, runner.RunnerGroup.OwnerID)
		if err != nil {
			return errors.Wrap(err, "unable to get install for runner")
		}
		ownerLabels = install.Labels
	case "orgs":
		ownerLabels = runner.Org.Labels
	}

	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return errors.Wrap(err, "unable to get logger")
	}

	// Only run health checks for active or offline processes; noop for any other status
	process, err := activities.AwaitGetRunnerProcessByProcessID(ctx, s.ProcessID)
	if err != nil {
		return dbgenerics.TemporalGormError(err, "runner process not found")
	}
	processType = string(process.Type)

	switch process.ProcessStatus() {
	case app.RunnerProcessStatusActive, app.RunnerProcessStatusOffline:
		// continue with health check
	default:
		return nil
	}

	// If a promotion requested shutdown, create the shutdown record and clear
	// the flag. The runner's shutdown poller will pick it up. Because health
	// check emitters run per-process on a 1-minute cron, shutdowns are
	// staggered naturally across all runners.
	if val, ok := process.CompositeStatus.Metadata["shutdown_requested"]; ok && val != nil {
		l.Info("shutdown requested via metadata, creating shutdown record",
			zap.String("runner_id", s.RunnerID),
			zap.String("process_id", s.ProcessID),
		)

		_, err := activities.AwaitCreateRunnerProcessShutdown(ctx, activities.CreateRunnerProcessShutdownRequest{
			RunnerProcessID: s.ProcessID,
			Type:            app.RunnerProcessShutdownTypeGraceful,
			CompositeStatus: app.CompositeStatus{
				Status:                 app.Status(app.RunnerProcessShutdownStatusRequested),
				StatusHumanDescription: "shutdown requested via promotion",
				CreatedAtTS:            workflow.Now(ctx).Unix(),
			},
		})
		if err != nil {
			return errors.Wrap(err, "unable to create shutdown for process")
		}

		// Clear the flag so we don't create duplicate shutdowns on the next health check.
		if clearErr := activities.AwaitClearProcessShutdownRequested(ctx, activities.ClearProcessShutdownRequestedRequest{
			ProcessID: s.ProcessID,
		}); clearErr != nil {
			l.Warn("unable to clear shutdown_requested metadata",
				zap.String("process_id", s.ProcessID),
				zap.Error(clearErr),
			)
		}

		return nil
	}

	heartbeat, err := activities.AwaitGetMostRecentHeartBeatByProcess(ctx, activities.GetMostRecentHeartBeatByProcessRequest{
		RunnerID:  s.RunnerID,
		ProcessID: s.ProcessID,
	})
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.Wrap(err, "unable to get heartbeat")
	}
	if heartbeat != nil {
		runnerVersion = heartbeat.Version
	}

	now := workflow.Now(ctx)
	heartbeatAge := s.heartbeatAge(now, heartbeat)

	// Tier 1: no heartbeat for 5 minutes → mark inactive and stop the queue
	if heartbeatAge >= inactiveTimeout {
		l.Warn("process inactive - no heartbeat for 5 minutes, stopping queue",
			zap.String("runner_id", s.RunnerID),
			zap.String("process_id", s.ProcessID),
		)

		_, err = activities.AwaitUpdateRunnerProcessStatus(ctx, activities.UpdateRunnerProcessStatusRequest{
			ProcessID:         s.ProcessID,
			Status:            app.RunnerProcessStatusInactive,
			StatusDescription: "no heartbeat received for 5 minutes",
		})
		if err != nil {
			return errors.Wrap(err, "unable to update process status to inactive")
		}

		stopTagMap := make(map[string]string, len(ownerLabels)+8)
		for k, v := range ownerLabels {
			stopTagMap[k] = v
		}
		stopTagMap["runner_id"] = s.RunnerID
		stopTagMap["runner_type"] = runnerType
		stopTagMap["runner_status"] = runnerStatus
		stopTagMap["runner_version"] = runnerVersion
		stopTagMap["process_id"] = s.ProcessID
		stopTagMap["process_type"] = string(process.Type)
		stopTagMap["org_id"] = orgID
		if installID != "" {
			stopTagMap["install_id"] = installID
		}
		stopTags := metrics.ToTags(stopTagMap)
		s.mw.Incr("runner.process.stop", stopTags)
		if process.StartedAt != nil {
			s.mw.Timing("runner.process.latency", workflow.Now(ctx).Sub(*process.StartedAt), stopTags)
		}

		// Enqueue on_inactive signal before stopping the queue
		_, err = sharedactivities.AwaitEnqueueSignalToOwner(ctx, &sharedactivities.EnqueueSignalToOwnerRequest{
			OwnerID:   s.RunnerID,
			OwnerType: "runners",
			QueueName: fmt.Sprintf("runner-process-%s", s.ProcessID),
			Signal: &oninactive.Signal{
				RunnerID:  s.RunnerID,
				ProcessID: s.ProcessID,
				Reason:    "offline",
			},
		})
		if err != nil {
			l.Warn("unable to enqueue on_inactive signal",
				zap.String("process_id", s.ProcessID),
				zap.Error(err),
			)
		}

		// Stop the process queue (terminates the cron emitter)
		err = activities.AwaitStopProcessQueue(ctx, activities.StopProcessQueueRequest{
			RunnerID:  s.RunnerID,
			ProcessID: s.ProcessID,
		})
		if err != nil {
			l.Warn("unable to stop process queue",
				zap.String("process_id", s.ProcessID),
				zap.Error(err),
			)
		}

		return nil
	}

	// Tier 2: no heartbeat for 1 minute → mark offline
	if heartbeatAge >= offlineTimeout {
		if process.ProcessStatus() != app.RunnerProcessStatusOffline {
			l.Warn("process offline - no heartbeat for 1 minute",
				zap.String("runner_id", s.RunnerID),
				zap.String("process_id", s.ProcessID),
			)

			_, err = activities.AwaitUpdateRunnerProcessStatus(ctx, activities.UpdateRunnerProcessStatusRequest{
				ProcessID:         s.ProcessID,
				Status:            app.RunnerProcessStatusOffline,
				StatusDescription: "Runner is offline and will be marked inactive in 5 minutes",
			})
			if err != nil {
				return errors.Wrap(err, "unable to update process status to offline")
			}
		}

		// Create red health check while offline
		_, err = activities.AwaitCreateHealthCheck(ctx, activities.CreateHealthCheckRequest{
			RunnerID:  s.RunnerID,
			ProcessID: s.ProcessID,
			Status:    app.RunnerStatusError,
		})
		if err != nil {
			l.Warn("unable to create offline health check",
				zap.String("process_id", s.ProcessID),
				zap.Error(err),
			)
		}

		return nil
	}

	// Heartbeat is fresh — ensure process is active
	if process.ProcessStatus() == app.RunnerProcessStatusOffline {
		l.Info("process back online",
			zap.String("runner_id", s.RunnerID),
			zap.String("process_id", s.ProcessID),
		)

		_, err = activities.AwaitUpdateRunnerProcessStatus(ctx, activities.UpdateRunnerProcessStatusRequest{
			ProcessID:         s.ProcessID,
			Status:            app.RunnerProcessStatusActive,
			StatusDescription: "heartbeat received",
		})
		if err != nil {
			return errors.Wrap(err, "unable to update process status to active")
		}
	}

	// Create health check record
	_, err = activities.AwaitCreateHealthCheck(ctx, activities.CreateHealthCheckRequest{
		RunnerID:  s.RunnerID,
		ProcessID: s.ProcessID,
		Status:    app.RunnerStatusActive,
	})
	if err != nil {
		return errors.Wrap(err, "unable to create process health check")
	}

	// Version mismatch check: compare configured version to runner's reported version
	if heartbeat != nil && heartbeat.Version != "" {
		runner, err := activities.AwaitGet(ctx, activities.GetRequest{RunnerID: s.RunnerID})
		if err != nil {
			l.Warn("unable to get runner for version comparison",
				zap.String("runner_id", s.RunnerID),
				zap.Error(err),
			)
		} else {
			settings := runner.RunnerGroup.Settings
			var configuredVersion string
			switch process.Type {
			case app.RunnerProcessTypeMng:
				configuredVersion = settings.BinaryVersion
			default:
				configuredVersion = settings.ContainerImageTag
			}

			var metadata map[string]any
			if configuredVersion != "" && configuredVersion != heartbeat.Version {
				metadata = map[string]any{
					"version_warning": fmt.Sprintf(
						"Reported runner version (%s) does not match configured version (%s). Please update the runner to the correct version.",
						heartbeat.Version, configuredVersion,
					),
				}
			} else {
				metadata = map[string]any{
					"version_warning": "",
				}
			}

			_, err = activities.AwaitUpdateRunnerProcessStatus(ctx, activities.UpdateRunnerProcessStatusRequest{
				ProcessID:         s.ProcessID,
				Status:            app.RunnerProcessStatusActive,
				StatusDescription: "",
				Metadata:          metadata,
			})
			if err != nil {
				l.Warn("unable to update version warning metadata",
					zap.String("process_id", s.ProcessID),
					zap.Error(err),
				)
			}
		}
	}

	return nil
}

func (s *Signal) heartbeatAge(now time.Time, heartbeat *app.RunnerHeartBeat) time.Duration {
	if heartbeat == nil {
		return inactiveTimeout
	}
	return now.Sub(heartbeat.CreatedAt)
}
