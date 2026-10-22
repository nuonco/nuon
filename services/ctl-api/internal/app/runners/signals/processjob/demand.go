package processjob

import (
	"time"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/worker/activities"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
)

func (s *Signal) executeDemand(ctx workflow.Context) error {
	actCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{StartToCloseTimeout: time.Minute})
	if s.Resume {
		return workflow.ExecuteActivity(actCtx, (*queueclient.Client).ResumeWorkflowForRunnerJob, s.JobID).Get(ctx, nil)
	}
	resume := signaldb.SignalData{Signal: &Signal{RunnerID: s.RunnerID, JobID: s.JobID, Demand: true, Resume: true, ExecutionID: s.ExecutionID}}
	ch := workflow.GetSignalChannel(ctx, TerminalSignalName(s.JobID))
	for attempt := 0; ; attempt++ {
		var finished bool
		if err := workflow.ExecuteActivity(actCtx, (*activities.Activities).PollDemandRunnerJob, s.JobID, s.ExecutionID, resume).Get(ctx, &finished); err != nil {
			return err
		}
		if finished {
			return nil
		}
		sleepOrSignal(ctx, ch, pollPeriod(attempt))
		if ctx.Err() != nil {
			return ctx.Err()
		}
	}
}
