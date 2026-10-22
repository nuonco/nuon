package runnerdemand_test

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/callback"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/catalog"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/job"
)

const preparedJobsSignalType signal.SignalType = "test-runner-demand-prepared-jobs"

type preparedJobsSignal struct {
	JobIDs         []string `json:"job_ids"`
	HoldBeforePark bool     `json:"hold_before_park,omitempty"`
}

func init() {
	catalog.Register(preparedJobsSignalType, func() signal.Signal { return &preparedJobsSignal{} })
}

func (*preparedJobsSignal) Type() signal.SignalType         { return preparedJobsSignalType }
func (*preparedJobsSignal) Validate(workflow.Context) error { return nil }
func (*preparedJobsSignal) SleepAfter() time.Duration       { return 0 }

func (s *preparedJobsSignal) Execute(ctx workflow.Context) error {
	phases := make([]job.Phase, 0, len(s.JobIDs))
	for i, jobID := range s.JobIDs {
		phases = append(phases, job.Phase{
			Key: fmt.Sprintf("job-%d", i),
			Prepare: func(workflow.Context) (string, error) {
				return jobID, nil
			},
		})
	}
	err := (job.Sequence{Phases: phases}).Run(ctx)
	if callback.IsSuspended(err) && s.HoldBeforePark {
		var release bool
		workflow.GetSignalChannel(ctx, "allow-park").Receive(ctx, &release)
	}
	return err
}
