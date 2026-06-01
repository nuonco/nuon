package updateversion

import (
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/signals/processjob"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

const SignalType signal.SignalType = "update_version"

type Signal struct {
	RunnerID      string `json:"runner_id"`
	HealthCheckID string `json:"health_check_id"`
}

var _ signal.Signal = (*Signal)(nil)

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.RunnerID == "" {
		return errors.New("runner_id is required")
	}
	if s.HealthCheckID == "" {
		return errors.New("health_check_id is required")
	}

	// Validate runner exists
	_, err := activities.AwaitGetByRunnerID(ctx, s.RunnerID)
	if err != nil {
		return errors.Wrap(err, "runner not found")
	}

	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	runner, err := activities.AwaitGetByRunnerID(ctx, s.RunnerID)
	if err != nil {
		return errors.Wrap(err, "unable to get runner")
	}

	// Create a logstream tied to the healthcheck
	logStream, err := activities.AwaitCreateLogStreamByOperationID(ctx, s.HealthCheckID)
	if err != nil {
		return errors.Wrap(err, "unable to create log stream for health check")
	}
	ctx = cctx.SetLogStreamWorkflowContext(ctx, logStream)

	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return errors.Wrap(err, "could not get logger")
	}

	// Create update-version job
	runnerJob, err := activities.AwaitCreateUpdateVersionJob(ctx, &activities.CreateUpdateVersionJobRequest{
		RunnerID:    runner.Org.RunnerGroup.Runners[0].ID,
		OwnerID:     s.HealthCheckID,
		LogStreamID: logStream.ID,
	})
	if err != nil {
		if updateErr := activities.AwaitUpdateStatus(ctx, activities.UpdateStatusRequest{
			RunnerID:          s.RunnerID,
			Status:            app.RunnerStatusError,
			StatusDescription: "unable to create update-version job",
		}); updateErr != nil {
			l.Error("failed to update runner status", zap.Error(updateErr))
		}
		statusactivities.AwaitUpdateRunnerStatusV2(ctx, statusactivities.UpdateRunnerStatusV2Request{
			RunnerID:          s.RunnerID,
			Status:            app.RunnerStatusError,
			StatusDescription: "unable to create update-version job",
		})
		return errors.Wrap(err, "unable to create job")
	}

	// Update the healthcheck with the job it caused to happen
	err = activities.AwaitSetHealthCheckRunnerJob(ctx, activities.SetHealthCheckRunnerJobRequest{
		HealthCheckID: s.HealthCheckID,
		RunnerJobID:   runnerJob.ID,
	})
	if err != nil {
		return errors.Wrap(err, "unable to set runner job on health check")
	}

	l.Info("dispatching job to update runner version",
		zap.String("runner_id", runner.Org.RunnerGroup.Runners[0].ID),
		zap.String("runner_type", string(runner.RunnerGroup.Type)),
		zap.String("expected_version", runner.RunnerGroup.Settings.ExpectedVersion),
	)

	// Send process_job signal to the runner's queue (cross-namespace)
	// This is fire-and-forget - we don't wait for the job to complete
	_, err = sharedactivities.AwaitEnqueueSignalToOwner(ctx, &sharedactivities.EnqueueSignalToOwnerRequest{
		OwnerID:   runner.Org.RunnerGroup.Runners[0].ID,
		OwnerType: "runners",
		Signal: &processjob.Signal{
			RunnerID: runner.Org.RunnerGroup.Runners[0].ID,
			JobID:    runnerJob.ID,
		},
	})
	if err != nil {
		return errors.Wrap(err, "unable to send process_job signal")
	}

	return nil
}
