package worker

import (
	"fmt"

	"github.com/powertoolsdev/mono/pkg/workflows/types/executors"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/protos"
	"go.temporal.io/sdk/workflow"
)

func (w *Workflows) executeDeprovisionOrgRunner(ctx workflow.Context, runnerID string, sandboxMode bool) error {
	req := &executors.DeprovisionRunnerRequest{
		RunnerID: runnerID,
	}
	var resp executors.ProvisionRunnerResponse
	err := w.execChildWorkflow(ctx, runnerID, executors.DeprovisionRunnerWorkflowName, sandboxMode, req, &resp)
	if err != nil {
		w.updateStatus(ctx, runnerID, app.RunnerStatusError, "unable to deprovision runner")
		return fmt.Errorf("unable to deprovision runner: %w", err)
	}

	return nil
}

func (w *Workflows) executeDeprovisionInstallRunner(ctx workflow.Context, runnerID string, sandboxMode bool) error {
	runner, err := activities.AwaitGet(ctx, activities.GetRequest{
		RunnerID: runnerID,
	})
	if err != nil {
		w.updateStatus(ctx, runnerID, app.RunnerStatusError, "unable to get runner")
		return fmt.Errorf("unable to get runner: %w", err)
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
	runnerJob, err := activities.AwaitCreateJob(ctx, &activities.CreateJobRequest{
		RunnerID:  runner.Org.RunnerGroup.Runners[0].ID,
		OwnerType: "runners",
		OwnerID:   runnerID,
		Op:        app.RunnerJobOperationTypeDestroy,
		Type:      runner.RunnerGroup.Platform.JobType(),
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
		Type: signals.OperationJobQueued,
	})
	// wait for the job

	w.updateStatus(ctx, runnerID, app.RunnerStatusActive, "runner is active and ready to process jobs")
	return nil
}
