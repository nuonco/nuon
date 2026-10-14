package runnerdemand_test

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/catalog"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/job"
)

const preparedJobsSignalType signal.SignalType = "test-runner-demand-prepared-jobs"

type preparedJobsSignal struct {
	JobIDs         []string `json:"job_ids"`
	HoldBeforePark bool     `json:"hold_before_park,omitempty"`
}

type phaseCheckpoint struct {
	JobID    string `json:"job_id"`
	Finished bool   `json:"finished"`
}

func init() {
	catalog.Register(preparedJobsSignalType, func() signal.Signal { return &preparedJobsSignal{} })
}

func (*preparedJobsSignal) Type() signal.SignalType         { return preparedJobsSignalType }
func (*preparedJobsSignal) Validate(workflow.Context) error { return nil }
func (*preparedJobsSignal) SleepAfter() time.Duration       { return 0 }

func (s *preparedJobsSignal) Execute(ctx workflow.Context) error {
	for i, jobID := range s.JobIDs {
		phase := fmt.Sprintf("job-%d", i)
		var checkpoint phaseCheckpoint
		found, err := job.LoadCheckpoint(ctx, phase, &checkpoint)
		if err != nil {
			return err
		}
		if checkpoint.Finished {
			continue
		}
		if !found {
			checkpoint.JobID = jobID
			if err := job.SaveCheckpoint(ctx, phase, checkpoint); err != nil {
				return err
			}
		}
		if _, err := job.Execute(ctx, &job.ExecuteJobRequest{JobID: checkpoint.JobID}, nil); err != nil {
			if job.IsSuspended(err) && s.HoldBeforePark {
				var release bool
				workflow.GetSignalChannel(ctx, "allow-park").Receive(ctx, &release)
			}
			return err
		}
		checkpoint.Finished = true
		if err := job.SaveCheckpoint(ctx, phase, checkpoint); err != nil {
			return err
		}
	}
	return nil
}
