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

func (w *Workflows) execDeployLegacy(ctx workflow.Context, install *app.Install, installDeploy *app.InstallDeploy, sandboxMode bool) error {
	w.updateDeployStatus(ctx, installDeploy.ID, app.InstallDeployStatusPlanning, "creating deploy plan")

	deployCfg, err := activities.AwaitGetComponentConfig(ctx, activities.GetComponentConfigRequest{
		DeployID: installDeploy.ID,
	})
	if err != nil {
		w.updateDeployStatus(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to get component config")
		w.writeDeployEvent(ctx, installDeploy.ID, signals.OperationDeploy, app.OperationStatusFailed)
		return fmt.Errorf("unable to get deploy component config: %w", err)
	}

	// execute the plan phase here
	deployPlanWorkflowID := fmt.Sprintf("%s-deploy-plan-%s", install.ID, installDeploy.ID)
	deployPlanTyp := planv1.ComponentInputType_COMPONENT_INPUT_TYPE_WAYPOINT_DEPLOY
	if installDeploy.Type == app.InstallDeployTypeTeardown {
		deployPlanTyp = planv1.ComponentInputType_COMPONENT_INPUT_TYPE_WAYPOINT_DESTROY
	}
	planResp, err := w.execCreatePlanWorkflow(ctx, sandboxMode, deployPlanWorkflowID, &planv1.CreatePlanRequest{
		Input: &planv1.CreatePlanRequest_Component{
			Component: &planv1.ComponentInput{
				OrgId:     install.App.OrgID,
				AppId:     install.App.ID,
				InstallId: install.ID,
				BuildId:   installDeploy.ComponentBuildID,
				DeployId:  installDeploy.ID,
				Component: deployCfg,
				Type:      deployPlanTyp,
				Context:   w.protos.InstallContext(install),
			},
		},
	})
	if err != nil {
		w.updateDeployStatus(ctx, installDeploy.ID, app.InstallDeployStatusError, fmt.Sprintf("unable to create deploy plan: %s", err))
		w.writeDeployEvent(ctx, installDeploy.ID, signals.OperationDeploy, app.OperationStatusFailed)
		return fmt.Errorf("unable to create deploy plan: %w", err)
	}

	w.updateDeployStatus(ctx, installDeploy.ID, app.InstallDeployStatusExecuting, "executing deploy plan")

	deployExecuteWorkflowID := fmt.Sprintf("%s-deploy-execute-%s", install.ID, installDeploy.ID)
	_, err = w.execExecPlanWorkflow(ctx, sandboxMode, deployExecuteWorkflowID, &execv1.ExecutePlanRequest{
		Plan: planResp.Ref,
	})
	if err != nil {
		w.updateDeployStatus(ctx, installDeploy.ID, app.InstallDeployStatusError, fmt.Sprintf("unable to execute deploy plan: %s", err))
		w.writeDeployEvent(ctx, installDeploy.ID, signals.OperationDeploy, app.OperationStatusFailed)
		return fmt.Errorf("unable to execute deploy plan: %w", err)
	}

	return nil
}

func (w *Workflows) execDeploy(ctx workflow.Context, install *app.Install, installDeploy *app.InstallDeploy, sandboxMode bool) error {
	w.updateDeployStatus(ctx, installDeploy.ID, app.InstallDeployStatusPlanning, "deploying")

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
	planReq := w.protos.ToDeployPlanRequest(install, installDeploy, deployCfg)
	deployImagePlanWorkflowID := fmt.Sprintf("%s-deploy-%s", install.ID, installDeploy.ID)
	planResp, err := w.execCreatePlanWorkflow(ctx, sandboxMode, deployImagePlanWorkflowID, planReq)
	if err != nil {
		w.updateDeployStatus(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to create deploy plan")
		w.writeDeployEvent(ctx, installDeploy.ID, signals.OperationDeploy, app.OperationStatusFailed)
		return fmt.Errorf("unable to create plan: %w", err)
	}

	// create the job
	runnerJob, err := activities.AwaitCreateDeployJob(ctx, &activities.CreateDeployJobRequest{
		RunnerID: install.RunnerGroup.Runners[0].ID,
		DeployID: installDeploy.ID,
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

	w.evClient.Send(ctx, install.RunnerGroup.Runners[0].ID, &runnersignals.Signal{
		Type: runnersignals.OperationJobQueued,
	})
	if err := w.pollJob(ctx, runnerJob.ID); err != nil {
		w.updateDeployStatus(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to execute runner job")
		w.writeDeployEvent(ctx, installDeploy.ID, signals.OperationDeploy, app.OperationStatusFailed)
		return fmt.Errorf("unable to get install: %w", err)
	}

	return nil
}
