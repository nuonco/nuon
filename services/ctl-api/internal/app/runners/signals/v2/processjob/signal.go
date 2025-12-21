package processjob

import (
	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "process_job"

type Signal struct {
	RunnerID string `json:"runner_id"`
	JobID    string `json:"job_id"`
}

var _ signal.Signal = (*Signal)(nil)

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.RunnerID == "" {
		return errors.New("runner_id is required")
	}
	if s.JobID == "" {
		return errors.New("job_id is required")
	}

	// Validate runner exists and is healthy
	runner, err := activities.AwaitGet(ctx, activities.GetRequest{RunnerID: s.RunnerID})
	if err != nil {
		return errors.Wrap(err, "runner not found")
	}

	if !runner.Status.IsHealthy() {
		return errors.Errorf("runner is not healthy: %s", runner.Status)
	}

	// Validate job exists
	_, err = activities.AwaitGetJob(ctx, activities.GetJobRequest{ID: s.JobID})
	return errors.Wrap(err, "job not found")
}

func (s *Signal) Execute(ctx workflow.Context) error {
	// TODO: Implement full process_job logic
	// This is a complex signal with ~533 lines of logic across multiple files.
	// For now, this is a stub that validates inputs.
	// Full implementation requires:
	// 1. Sandbox mode handling (needs config access via activity)
	// 2. Queue timeout checking
	// 3. Retry logic with multiple execution attempts
	// 4. Job execution monitoring with polling
	// 5. Metrics emission (needs metrics wrapper via activity)
	// 6. Complex timeout handling (available, execution, overall)

	return errors.New("process_job signal not yet fully implemented")
}
