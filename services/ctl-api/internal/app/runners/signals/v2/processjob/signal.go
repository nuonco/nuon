package processjob

import (
	"time"

	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/signals"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/eventloop"
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
	childCtx := workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
		WorkflowExecutionTimeout: 70 * time.Minute,
	})

	// RequestSignal uses EventLoopRequest.ID for the runner ID and Signal.JobID for the job ID.
	req := signals.NewRequestSignal(
		eventloop.EventLoopRequest{ID: s.RunnerID},
		&signals.Signal{
			Type:  signals.OperationProcessJob,
			JobID: s.JobID,
		},
	)

	return workflow.ExecuteChildWorkflow(childCtx, "ProcessJob", req).Get(ctx, nil)
}
