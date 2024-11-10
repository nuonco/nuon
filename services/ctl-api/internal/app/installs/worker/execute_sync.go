package worker

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	logv1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/log/v1"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
	runnersignals "github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/protos"
)

func (w *Workflows) execSync(ctx workflow.Context, install *app.Install, installDeploy *app.InstallDeploy, sandboxMode bool) error {
	w.updateDeployStatus(ctx, installDeploy.ID, app.InstallDeployStatusPlanning, "creating sync plan")

	token, err := activities.AwaitCreateJobLogTokenByRunnerID(ctx, install.RunnerGroup.Runners[0].ID)
	if err != nil {
		return fmt.Errorf("unable to create job log token: %w", err)
	}

	build, err := activities.AwaitGetComponentBuildByComponentBuildID(ctx, installDeploy.ComponentBuildID)
	if err != nil {
		w.updateDeployStatus(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to get component build")
		w.writeDeployEvent(ctx, installDeploy.ID, signals.OperationDeploy, app.OperationStatusFailed)
		return fmt.Errorf("unable to get build: %w", err)
	}

	// create the job
	runnerJob, err := activities.AwaitCreateSyncJob(ctx, &activities.CreateSyncJobRequest{
		DeployID: installDeploy.ID,
		RunnerID: install.RunnerGroup.Runners[0].ID,
		Op:       installDeploy.Type.RunnerJobOperationType(),
		Type:     build.ComponentConfigConnection.Type.SyncJobType(),
	})
	if err != nil {
		w.updateDeployStatus(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to create runner job")
		w.writeDeployEvent(ctx, installDeploy.ID, signals.OperationDeploy, app.OperationStatusFailed)
		return fmt.Errorf("unable to create runner job: %w", err)
	}

	deployCfg, err := activities.AwaitGetComponentConfig(ctx, activities.GetComponentConfigRequest{
		DeployID: installDeploy.ID,
	})
	if err != nil {
		w.updateDeployStatus(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to get component config")
		w.writeDeployEvent(ctx, installDeploy.ID, signals.OperationDeploy, app.OperationStatusFailed)
		return fmt.Errorf("unable to get deploy component config: %w", err)
	}

	// create the sandbox plan request
	planReq := w.protos.ToSyncPlanRequest(install, installDeploy, deployCfg)
	planReq.LogConfiguration = &logv1.LogConfiguration{
		RunnerId:       install.RunnerGroup.Runners[0].ID,
		RunnerApiToken: token.Token,
		RunnerApiUrl:   w.cfg.RunnerAPIURL,
		RunnerJobId:    runnerJob.ID,
		Attrs:          logv1.NewAttrs(generics.ToStringMap(install.RunnerGroup.Settings.Metadata)),
	}

	syncImagePlanWorkflowID := fmt.Sprintf("%s-sync-plan-%s", install.ID, installDeploy.ID)
	planResp, err := w.execCreatePlanWorkflow(ctx, sandboxMode, syncImagePlanWorkflowID, planReq)
	if err != nil {
		w.updateDeployStatus(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to create sync plan")
		w.writeDeployEvent(ctx, installDeploy.ID, signals.OperationDeploy, app.OperationStatusFailed)
		return fmt.Errorf("unable to create plan: %w", err)
	}

	// store the plan in the db
	planJSON, err := protos.ToJSON(planResp.Plan)
	if err != nil {
		w.updateDeployStatus(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to store runner job plan")
		w.writeDeployEvent(ctx, installDeploy.ID, signals.OperationDeploy, app.OperationStatusFailed)
		return fmt.Errorf("unable to convert plan to json: %w", err)
	}
	if err := activities.AwaitSaveRunnerJobPlan(ctx, &activities.SaveRunnerJobPlanRequest{
		JobID:    runnerJob.ID,
		PlanJSON: string(planJSON),
	}); err != nil {
		w.updateDeployStatus(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to store runner job plan")
		w.writeDeployEvent(ctx, installDeploy.ID, signals.OperationDeploy, app.OperationStatusFailed)
		return fmt.Errorf("unable to get install: %w", err)
	}

	// queue job
	w.evClient.Send(ctx, install.RunnerGroup.Runners[0].ID, &runnersignals.Signal{
		Type:  runnersignals.OperationProcessJob,
		JobID: runnerJob.ID,
	})
	if err := w.pollJob(ctx, runnerJob.ID); err != nil {
		w.updateDeployStatus(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to poll job")
		w.writeDeployEvent(ctx, installDeploy.ID, signals.OperationDeploy, app.OperationStatusFailed)
		return fmt.Errorf("unable to poll job: %w", err)
	}

	return nil
}
