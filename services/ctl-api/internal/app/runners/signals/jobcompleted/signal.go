package jobcompleted

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "job-completed"

type Signal struct {
	JobID    string `json:"job_id"`
	RunnerID string `json:"runner_id"`
}

var _ signal.Signal = (*Signal)(nil)

func (s *Signal) Type() signal.SignalType { return SignalType }

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.JobID == "" {
		return fmt.Errorf("job_id is required")
	}
	job, err := activities.AwaitGetJobByID(ctx, s.JobID)
	if err != nil {
		return fmt.Errorf("load job: %w", err)
	}
	if !job.Status.IsTerminal() {
		return fmt.Errorf("job %s is %s", s.JobID, job.Status)
	}
	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	if err := activities.AwaitDeleteJobCredentials(ctx, activities.DeleteJobCredentialsRequest{JobID: s.JobID}); err != nil {
		return fmt.Errorf("delete job credentials: %w", err)
	}
	return nil
}
