package worker

import (
	"fmt"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
	runnersignals "github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/log"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/protos"
)

func (w *Workflows) execDeploy(ctx workflow.Context, install *app.Install, installDeploy *app.InstallDeploy, sandboxMode bool) error {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}

	w.updateDeployStatus(ctx, installDeploy.ID, app.InstallDeployStatusPlanning, "deploying")
	l.Info("executing deploy")

	build, err := activities.AwaitGetComponentBuildByComponentBuildID(ctx, installDeploy.ComponentBuildID)
	if err != nil {
		w.updateDeployStatus(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to get component build")
		w.writeDeployEvent(ctx, installDeploy.ID, signals.OperationDeploy, app.OperationStatusFailed)
		return fmt.Errorf("unable to get build: %w", err)
	}

	logStreamID, err := cctx.GetLogStreamIDWorkflow(ctx)
	if err != nil {
		return err
	}
	runnerJob, err := activities.AwaitCreateDeployJob(ctx, &activities.CreateDeployJobRequest{
		RunnerID:    install.RunnerGroup.Runners[0].ID,
		DeployID:    installDeploy.ID,
		Op:          installDeploy.Type.RunnerJobOperationType(),
		Type:        build.ComponentConfigConnection.Type.DeployJobType(),
		LogStreamID: logStreamID,
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
	planReq := w.protos.ToDeployPlanRequest(install, installDeploy, deployCfg)

	l.Info("creating deploy plan")
	deployImagePlanWorkflowID := fmt.Sprintf("%s-deploy-%s", install.ID, installDeploy.ID)
	planResp, err := w.execCreatePlanWorkflow(ctx, sandboxMode, deployImagePlanWorkflowID, planReq)
	if err != nil {
		w.updateDeployStatus(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to create deploy plan")
		w.writeDeployEvent(ctx, installDeploy.ID, signals.OperationDeploy, app.OperationStatusFailed)
		l.Error("error creating deploy plan", zap.Error(err))
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

	w.evClient.Send(ctx, install.RunnerGroup.Runners[0].ID, &runnersignals.Signal{
		JobID: runnerJob.ID,
		Type:  runnersignals.OperationProcessJob,
	})
	if err := w.pollJob(ctx, runnerJob.ID); err != nil {
		w.updateDeployStatus(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to execute runner job")
		w.writeDeployEvent(ctx, installDeploy.ID, signals.OperationDeploy, app.OperationStatusFailed)
		l.Error("job did not succeed", zap.Error(err))
		return fmt.Errorf("unable to get install: %w", err)
	}

	return nil
}
