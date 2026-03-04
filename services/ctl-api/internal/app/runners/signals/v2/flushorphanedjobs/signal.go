package flushorphanedjobs

import (
	"time"

	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const (
	SignalType signal.SignalType = "flush_orphaned_jobs"

	// any job over 12 hours old that is _still_ queued will be automatically flushed
	orphanedJobsThreshold time.Duration = time.Hour * 12
)

type Signal struct {
	RunnerID string `json:"runner_id"`
}

var _ signal.Signal = (*Signal)(nil)

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.RunnerID == "" {
		return errors.New("runner_id is required")
	}

	// Validate runner exists in database
	_, err := activities.AwaitGetByRunnerID(ctx, s.RunnerID)
	if err != nil {
		return errors.Wrap(err, "runner not found")
	}

	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	// Calculate threshold time (12 hours ago from workflow time)
	ts := workflow.Now(ctx)
	threshold := ts.Add(-orphanedJobsThreshold)

	// Flush orphaned jobs older than threshold
	if err := activities.AwaitFlushOrphanedJobs(ctx, activities.FlushOrphanedJobsRequest{
		RunnerID:  s.RunnerID,
		Threshold: threshold,
	}); err != nil {
		return errors.Wrap(err, "unable to flush orphaned jobs")
	}

	return nil
}
