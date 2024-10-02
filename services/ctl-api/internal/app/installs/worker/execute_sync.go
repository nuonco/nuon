package worker

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	execv1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/execute/v1"
	planv1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/plan/v1"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
	runnersignals "github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/protos"
)

func (w *Workflows) execSyncLegacy(ctx workflow.Context, install *app.Install, installDeploy *app.InstallDeploy, sandboxMode bool) error {
	w.updateDeployStatus(ctx, installDeploy.ID, app.InstallDeployStatusPlanning, "creating sync plan")

	deployCfg, err := activities.AwaitGetComponentConfig(ctx, activities.GetComponentConfigRequest{
		DeployID: installDeploy.ID,
	})
	if err != nil {
		w.updateDeployStatus(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to get component config")
		w.writeDeployEvent(ctx, installDeploy.ID, signals.OperationDeploy, app.OperationStatusFailed)
		return fmt.Errorf("unable to get deploy component config: %w", err)
	}

	// execute the plan phase here
	syncImagePlanWorkflowID := fmt.Sprintf("%s-sync-plan-%s", install.ID, installDeploy.ID)
	planResp, err := w.execCreatePlanWorkflow(ctx, sandboxMode, syncImagePlanWorkflowID, &planv1.CreatePlanRequest{
		Input: &planv1.CreatePlanRequest_Component{
			Component: &planv1.ComponentInput{
				OrgId:     install.App.OrgID,
				AppId:     install.App.ID,
				BuildId:   installDeploy.ComponentBuildID,
				InstallId: install.ID,
				DeployId:  installDeploy.ID,
				Component: deployCfg,
				Type:      planv1.ComponentInputType_COMPONENT_INPUT_TYPE_WAYPOINT_SYNC_IMAGE,
				Context:   w.protos.InstallContext(install),
			},
		},
	})
	if err != nil {
		w.updateDeployStatus(ctx, installDeploy.ID, app.InstallDeployStatusError, fmt.Sprintf("unable to create sync plan: %s", err))
		w.writeDeployEvent(ctx, installDeploy.ID, signals.OperationDeploy, app.OperationStatusFailed)
		return fmt.Errorf("unable to create sync plan: %w", err)
	}

	w.updateDeployStatus(ctx, installDeploy.ID, app.InstallDeployStatusSyncing, "executing sync plan")

	syncExecuteWorkflowID := fmt.Sprintf("%s-sync-execute-%s", install.ID, installDeploy.ID)
	_, err = w.execExecPlanWorkflow(ctx, sandboxMode, syncExecuteWorkflowID, &execv1.ExecutePlanRequest{
		Plan: planResp.Ref,
	})
	if err != nil {
		w.updateDeployStatus(ctx, installDeploy.ID, app.InstallDeployStatusError, fmt.Sprintf("unable to execute sync plan: %s", err))
		w.writeDeployEvent(ctx, installDeploy.ID, signals.OperationDeploy, app.OperationStatusFailed)
		return fmt.Errorf("unable to execute sync plan: %w", err)
	}

	return nil
}

func (w *Workflows) execSync(ctx workflow.Context, install *app.Install, installDeploy *app.InstallDeploy, sandboxMode bool) error {
	w.updateDeployStatus(ctx, installDeploy.ID, app.InstallDeployStatusPlanning, "creating sync plan")

	build, err := activities.AwaitGetComponentBuildByComponentBuildID(ctx, installDeploy.ComponentBuildID)
	if err != nil {
		w.updateDeployStatus(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to get component build")
		w.writeDeployEvent(ctx, installDeploy.ID, signals.OperationDeploy, app.OperationStatusFailed)
		return fmt.Errorf("unable to get build: %w", err)
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
	syncImagePlanWorkflowID := fmt.Sprintf("%s-sync-plan-%s", install.ID, installDeploy.ID)
	planResp, err := w.execCreatePlanWorkflow(ctx, sandboxMode, syncImagePlanWorkflowID, planReq)
	if err != nil {
		w.updateDeployStatus(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to create sync plan")
		w.writeDeployEvent(ctx, installDeploy.ID, signals.OperationDeploy, app.OperationStatusFailed)
		return fmt.Errorf("unable to create plan: %w", err)
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
		Type:  runnersignals.OperationJobQueued,
		JobID: runnerJob.ID,
	})
	if err := w.pollJob(ctx, runnerJob.ID); err != nil {
		w.updateDeployStatus(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to poll job")
		w.writeDeployEvent(ctx, installDeploy.ID, signals.OperationDeploy, app.OperationStatusFailed)
		return fmt.Errorf("unable to poll job: %w", err)
	}

	return nil
}
