package worker

import (
	"encoding/json"

	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/actions/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/actions/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/actions/worker/job"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/actions/worker/plan"
	runnersignals "github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/log"
)

// @temporal-gen workflow
// @execution-timeout 1m
// @task-timeout 30s
func (w *Workflows) InstallRun(ctx workflow.Context, sreq signals.RequestSignal) error {
	run, err := activities.AwaitGetInstallActionWorkflowRunByRunID(ctx, sreq.RunID)
	if err != nil {
		return errors.Wrap(err, "unable to get action workflow run")
	}

	ls, err := activities.AwaitCreateLogStreamByActionWorkflowRunID(ctx, sreq.RunID)
	if err != nil {
		return errors.Wrap(err, "unable to create log stream")
	}

	defer func() {
		activities.AwaitCloseLogStreamByLogStreamID(ctx, ls.ID)
	}()
	ctx = cctx.SetLogStreamWorkflowContext(ctx, ls)
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return errors.Wrap(err, "unable to set log stream on context")
	}

	l.Info("creating plan for executing action run")
	runPlan, err := plan.AwaitCreateActionRunPlan(ctx, &plan.CreateActionRunPlanRequest{
		RunID: sreq.RunID,
	})
	if err != nil {
		return errors.Wrap(err, "unable to create plan")
	}

	// execute job
	l.Info("creating runner job to execute action")
	runnerJob, err := activities.AwaitCreateRunnerJob(ctx, &activities.CreateRunnerJobRequest{
		InstallActionWorkflowRunID: sreq.RunID,
		RunnerID:                   run.Install.RunnerID,
	})
	if err != nil {
		return errors.Wrap(err, "unable to create runner job")
	}

	// save runner job plan
	planJSON, err := json.Marshal(runPlan)
	if err != nil {
		return errors.Wrap(err, "unable to convert plan to json")
	}
	if err := activities.AwaitSaveRunnerJobPlan(ctx, &activities.SaveRunnerJobPlanRequest{
		JobID:    runnerJob.ID,
		PlanJSON: string(planJSON),
	}); err != nil {
		return errors.Wrap(err, "unable to save runner job plan")
	}

	// now queue and execute the job
	l.Info("executing runner job")
	// TODO(jm): this signal should be sent via the workflow, as well.
	w.evClient.Send(ctx, run.Install.RunnerID, &runnersignals.Signal{
		Type:  runnersignals.OperationProcessJob,
		JobID: runnerJob.ID,
	})
	_, err = job.AwaitExecuteJob(ctx, &job.ExecuteJobRequest{
		JobID:      runnerJob.ID,
		WorkflowID: "",
	})
	if err != nil {
		return errors.Wrap(err, "runner job failed")
	}

	return nil
}
