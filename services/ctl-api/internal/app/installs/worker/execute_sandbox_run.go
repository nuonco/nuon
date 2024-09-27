package worker

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/pkg/workflows/types/executors"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
	runnersignals "github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/protos"
)

func (w *Workflows) executeSandboxRun(ctx workflow.Context, install *app.Install, installRun *app.InstallSandboxRun, op app.RunnerJobOperationType, sandboxMode bool) error {
	// create an install provision request
	req, err := w.protos.ToInstallProvisionRequest(install, installRun.ID)
	if err != nil {
		w.updateRunStatus(ctx, installRun.ID, app.SandboxRunStatusError, "unable to create install provision request")
		return fmt.Errorf("unable to get install provision request: %w", err)
	}

	// check permissions
	var resp executors.CheckPermissionsResponse
	if err := w.execChildWorkflow(ctx, install.ID, executors.CheckPermissionsWorkflowName, sandboxMode, req, &resp); err != nil {
		w.updateRunStatus(ctx, installRun.ID, app.SandboxRunStatusAccessError, "unable to validate credentials before install "+err.Error())
		return fmt.Errorf("unable to validate credentials before install: %w", err)
	}

	// create the sandbox plan request
	planReq, err := w.protos.ToInstallPlanRequest(install, installRun.ID)
	if err != nil {
		w.updateRunStatus(ctx, installRun.ID, app.SandboxRunStatusError, "unable to create install plan request")
		return fmt.Errorf("unable to get install plan request: %w", err)
	}

	// create the sandbox plan
	planWorkflowID := fmt.Sprintf("%s-plan", installRun.ID)
	planResp, err := w.execCreatePlanWorkflow(ctx, sandboxMode, planWorkflowID, planReq)
	if err != nil {
		w.updateRunStatus(ctx, installRun.ID, app.SandboxRunStatusError, "unable to create install plan")
		return fmt.Errorf("unable to create install plan: %w", err)
	}

	// create the job
	runnerJob, err := activities.AwaitCreateSandboxJob(ctx, &activities.CreateSandboxJobRequest{
		InstallID: install.ID,
		RunnerID:  install.Org.RunnerGroup.Runners[0].ID,
		OwnerType: "runners",
		OwnerID:   install.Org.RunnerGroup.Runners[0].ID,
		Op:        op,
	})
	if err != nil {
		w.updateRunStatus(ctx, installRun.ID, app.SandboxRunStatusError, "unable to create runner job")
		return fmt.Errorf("unable to create runner job: %w", err)
	}

	// store the plan in the db
	planJSON, err := protos.ToJSON(planResp.Plan)
	if err != nil {
		w.updateRunStatus(ctx, installRun.ID, app.SandboxRunStatusError, "unable to convert plan to json")
		return fmt.Errorf("unable to convert plan to json: %w", err)
	}

	if err := activities.AwaitSaveRunnerJobPlan(ctx, &activities.SaveRunnerJobPlanRequest{
		JobID:    runnerJob.ID,
		PlanJSON: string(planJSON),
	}); err != nil {
		w.updateRunStatus(ctx, installRun.ID, app.SandboxRunStatusError, "unable to save plan")
		return fmt.Errorf("unable to get install: %w", err)
	}

	// queue job
	w.evClient.Send(ctx, install.Org.RunnerGroup.Runners[0].ID, &runnersignals.Signal{
		Type:  runnersignals.OperationJobQueued,
		JobID: runnerJob.ID,
	})
	if err := w.pollJob(ctx, runnerJob.ID); err != nil {
		return fmt.Errorf("unable to poll job: %w", err)
	}

	return nil
}
