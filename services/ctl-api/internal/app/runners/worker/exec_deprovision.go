package worker

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/pkg/workflows/types/executors"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/log"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/protos"
)

func (w *Workflows) executeDeprovisionOrgRunner(ctx workflow.Context, runnerID string, sandboxMode bool) error {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}

	runner, err := activities.AwaitGet(ctx, activities.GetRequest{
		RunnerID: runnerID,
	})
	if err != nil {
		w.updateStatus(ctx, runnerID, app.RunnerStatusError, "unable to get runner")
		return fmt.Errorf("unable to get runner: %w", err)
	}

	if runner.RunnerGroup.Platform == app.AppRunnerTypeLocal {
		l.Info("skipping local runner")
		return nil
	}
	if runner.Org.OrgType == app.OrgTypeIntegration {
		return nil
	}

	req := &executors.DeprovisionRunnerRequest{
		RunnerID: runnerID,
	}
	var resp executors.DeprovisionRunnerResponse
	err = w.execChildWorkflow(ctx, runnerID, executors.DeprovisionRunnerWorkflowName, sandboxMode, req, &resp)
	if err != nil {
		w.updateStatus(ctx, runnerID, app.RunnerStatusError, "unable to deprovision runner")
		return fmt.Errorf("unable to deprovision runner: %w", err)
	}

	return nil
}

func (w *Workflows) executeDeprovisionInstallRunner(ctx workflow.Context, runnerID string, sandboxMode bool) error {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}

	l.Info("deprovisioning install runner")
	runner, err := activities.AwaitGet(ctx, activities.GetRequest{
		RunnerID: runnerID,
	})
	if err != nil {
		w.updateStatus(ctx, runnerID, app.RunnerStatusError, "unable to get runner")
		return fmt.Errorf("unable to get runner: %w", err)
	}

	if runner.RunnerGroup.Platform == app.AppRunnerTypeLocal {
		l.Info("skipping local runner")
		return nil
	}
	if runner.Org.OrgType == app.OrgTypeIntegration {
		l.Info("skipping local runner for integration org")
		return nil
	}

	if runner.Status != app.RunnerStatusDeprovisioning && !runner.Status.IsHealthy() {
		l.Warn("runner was not successfully provisioned, so skipping and letting the sandbox deprovision remove it")
		return nil
	}

	install, err := activities.AwaitGetInstall(ctx, activities.GetInstallRequest{
		InstallID: runner.RunnerGroup.OwnerID,
	})
	if err != nil {
		return fmt.Errorf("unable to get runner install: %w", err)
	}

	// create the sandbox plan request
	planWorkflowID := fmt.Sprintf("%s-runner-deprovision", runnerID)
	planReq, err := w.protos.ToRunnerInstallPlanRequest(runner, install, "")
	if err != nil {
		w.updateStatus(ctx, runnerID, app.RunnerStatusError, "unable to create runner plan request")
		return fmt.Errorf("unable to create runner plan: %w", err)
	}

	planResp, err := w.execCreatePlanWorkflow(ctx, sandboxMode, planWorkflowID, planReq)
	if err != nil {
		w.updateStatus(ctx, runnerID, app.RunnerStatusError, "unable to create provision plan")
		return fmt.Errorf("unable to create runner plan: %w", err)
	}

	// create the job
	logStream, err := cctx.GetLogStreamWorkflow(ctx)
	if err != nil {
		return errors.Wrap(err, "no log stream found")
	}
	runnerJob, err := activities.AwaitCreateJob(ctx, &activities.CreateJobRequest{
		RunnerID:    runner.Org.RunnerGroup.Runners[0].ID,
		OwnerType:   "runners",
		OwnerID:     runnerID,
		Op:          app.RunnerJobOperationTypeDestroy,
		Type:        runner.RunnerGroup.Platform.JobType(),
		LogStreamID: logStream.ID,
	})
	if err != nil {
		w.updateStatus(ctx, runnerID, app.RunnerStatusError, "unable to create deprovision job")
		return fmt.Errorf("unable to create job: %w", err)
	}

	// store the plan in the db
	planJSON, err := protos.ToJSON(planResp.Plan)
	if err != nil {
		return fmt.Errorf("unable to convert plan to json: %w", err)
	}
	if err := activities.AwaitSaveRunnerJobPlan(ctx, &activities.SaveRunnerJobPlanRequest{
		JobID:    runnerJob.ID,
		PlanJSON: string(planJSON),
	}); err != nil {
		w.updateStatus(ctx, runnerID, app.RunnerStatusError, "unable to create plan json")
		return fmt.Errorf("unable to save runner job plan: %w", err)
	}

	// queue job
	w.evClient.Send(ctx, runner.Org.RunnerGroup.Runners[0].ID, &signals.Signal{
		Type:  signals.OperationProcessJob,
		JobID: runnerJob.ID,
	})
	if err := w.pollJob(ctx, runnerJob.ID); err != nil {
		w.updateStatus(ctx, runnerID, app.RunnerStatusError, "unable to poll runner job to completion")
		return fmt.Errorf("unable to poll runner job to completion: %w", err)
	}

	w.updateStatus(ctx, runnerID, app.RunnerStatusDeprovisioned, "runner is deprovisioned")
	return nil
}
