package processhealthcheck

import (
	"time"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
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

	// Check if the process is still active or offline; noop for any other status
	process, err := activities.AwaitGetRunnerProcessByProcessID(ctx, s.ProcessID)
	if err != nil {
		return nil
	}
	if process.Status != app.RunnerProcessStatusActive && process.Status != app.RunnerProcessStatusOffline {
		l.Info("skipping process health check - process not active/offline",
			zap.String("process_id", s.ProcessID),
			zap.String("status", string(process.Status)),
		)
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
		if process.Status != app.RunnerProcessStatusOffline {
			l.Warn("process offline - no heartbeat for 1 minute",
				zap.String("runner_id", s.RunnerID),
				zap.String("process_id", s.ProcessID),
			)

			_, err = activities.AwaitUpdateRunnerProcessStatus(ctx, activities.UpdateRunnerProcessStatusRequest{
				ProcessID:         s.ProcessID,
				Status:            app.RunnerProcessStatusOffline,
				StatusDescription: "no heartbeat received for 1 minute",
			})
			if err != nil {
				return errors.Wrap(err, "unable to update process status to offline")
			}
		}

		return nil
	}

	// Heartbeat is fresh — ensure process is active
	if process.Status == app.RunnerProcessStatusOffline {
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

	return nil
}

func (s *Signal) heartbeatAge(now time.Time, heartbeat *app.RunnerHeartBeat) time.Duration {
	if heartbeat == nil {
		return inactiveTimeout
	}
	return now.Sub(heartbeat.CreatedAt)
}
