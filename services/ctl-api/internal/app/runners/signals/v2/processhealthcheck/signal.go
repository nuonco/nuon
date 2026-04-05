package processhealthcheck

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/signals/v2/oninactive"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/worker/activities"
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
}

var _ signal.Signal = (*Signal)(nil)

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

func (s *Signal) Execute(ctx workflow.Context) error {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return errors.Wrap(err, "unable to get logger")
	}

	// Only run health checks for active or offline processes; noop for any other status
	process, err := activities.AwaitGetRunnerProcessByProcessID(ctx, s.ProcessID)
	if err != nil {
		return nil
	}

	switch process.ProcessStatus() {
	case app.RunnerProcessStatusActive, app.RunnerProcessStatusOffline:
		// continue with health check
	default:
		return nil
	}

	heartbeat, err := activities.AwaitGetMostRecentHeartBeatByProcess(ctx, activities.GetMostRecentHeartBeatByProcessRequest{
		RunnerID:  s.RunnerID,
		ProcessID: s.ProcessID,
	})
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.Wrap(err, "unable to get heartbeat")
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

	// Version mismatch check: compare API version to runner's reported version
	if heartbeat != nil && heartbeat.Version != "" {
		apiVersion, err := activities.AwaitGetAPIVersion(ctx)
		if err != nil {
			l.Warn("unable to get API version for comparison",
				zap.String("process_id", s.ProcessID),
				zap.Error(err),
			)
		} else {
			var metadata map[string]any
			if apiVersion != heartbeat.Version {
				metadata = map[string]any{
					"version_warning": "Reported runner version does not match running API version and could cause issues. Please update the tag to the same version as the control plane (" + apiVersion + ")",
				}
			} else {
				// Clear the warning if versions now match
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
