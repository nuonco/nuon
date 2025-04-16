package worker

import (
	"fmt"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/log"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/notifications"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/protos"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/workflows/job"
)

// @temporal-gen workflow
// @execution-timeout 60m
// @task-timeout 30m
func (w *Workflows) ExecuteDeployComponent(ctx workflow.Context, sreq signals.RequestSignal) error {
	install, err := activities.AwaitGetByInstallID(ctx, sreq.ID)
	if err != nil {
		w.updateDeployStatus(ctx, sreq.DeployID, app.InstallDeployStatusError, "unable to get install from database")
		w.writeDeployEvent(ctx, sreq.DeployID, signals.OperationDeploy, app.OperationStatusFailed)
		return fmt.Errorf("unable to get install: %w", err)
	}

	var installDeploy *app.InstallDeploy
	if sreq.ExecuteDeployComponentSubSignal.DeployID != "" {
		installDeploy, err = activities.AwaitGetDeployByDeployID(ctx, sreq.ExecuteDeployComponentSubSignal.DeployID)
		if err != nil {
			return errors.Wrap(err, "unable to get deploy")
		}
	} else {
		componentBuild, err := activities.AwaitGetComponentLatestBuildByComponentID(ctx, sreq.ExecuteDeployComponentSubSignal.ComponentID)
		if err != nil {
			return fmt.Errorf("unable to get component build: %w", err)
		}

		installDeploy, err = activities.AwaitCreateInstallDeploy(ctx, activities.CreateInstallDeployRequest{
			InstallID:   install.ID,
			ComponentID: sreq.ExecuteDeployComponentSubSignal.ComponentID,
			BuildID:     componentBuild.ID,
		})
		if err != nil {
			return fmt.Errorf("unable to create install deploy: %w", err)
		}
		sreq.DeployID = installDeploy.ID
	}

	if err := activities.AwaitUpdateInstallWorkflowStepTarget(ctx, activities.UpdateInstallWorkflowStepTargetRequest{
		StepID:         sreq.WorkflowStepID,
		StepTargetID:   installDeploy.ID,
		StepTargetType: plugins.TableName(w.db, installDeploy),
	}); err != nil {
		return errors.Wrap(err, "unable to update install workflow")
	}

	defer func() {
		if pan := recover(); pan != nil {
			w.updateDeployStatus(ctx, sreq.DeployID, app.InstallDeployStatusError, "internal error")
			w.writeDeployEvent(ctx, sreq.DeployID, signals.OperationDeploy, app.OperationStatusFailed)
			panic(pan)
		}
	}()

	logStream, err := activities.AwaitCreateLogStream(ctx, activities.CreateLogStreamRequest{
		DeployID: sreq.DeployID,
	})
	if err != nil {
		return errors.Wrap(err, "unable to create log stream")
	}
	defer func() {
		activities.AwaitCloseLogStreamByLogStreamID(ctx, logStream.ID)
	}()

	ctx = cctx.SetLogStreamWorkflowContext(ctx, logStream)
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}

	l.Info("performing deploy")
	err = w.doDeploy(ctx, sreq, install)
	if err != nil {
		w.writeDeployEvent(ctx, sreq.DeployID, signals.OperationDeploy, app.OperationStatusFailed)
		w.sendNotification(ctx, notifications.NotificationsTypeDeployFailed, install.AppID, map[string]string{
			"install_name": install.Name,
			"app_name":     install.App.Name,
			"created_by":   install.CreatedBy.Email,
		})
	}
	return err
}

func (w *Workflows) execDeploy(ctx workflow.Context, install *app.Install, installDeploy *app.InstallDeploy, sandboxMode bool) error {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}

	w.updateDeployStatus(ctx, installDeploy.ID, app.InstallDeployStatusPlanning, "creating deploy plan")
	l.Info("planning and executing deploy")

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
		Metadata: map[string]string{
			"install_id":           install.ID,
			"deploy_id":            installDeploy.ID,
			"install_component_id": installDeploy.InstallComponentID,
			"component_id":         build.ComponentConfigConnection.ComponentID,
			"component_name":       build.ComponentConfigConnection.Component.Name,
		},
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
	deployImagePlanWorkflowID := fmt.Sprintf("%s-deploy-plan-%s", install.ID, installDeploy.ID)
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
		l.Error("unable to create plan", zap.Error(err))
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

	if err := activities.AwaitSaveIntermediateData(ctx, &activities.SaveIntermediateDataRequest{
		InstallID:   install.ID,
		RunnerJobID: runnerJob.ID,
		PlanJSON:    string(planJSON),
	}); err != nil {
		return errors.Wrap(err, "unable to save install intermediate data")
	}

	w.updateDeployStatus(ctx, installDeploy.ID, app.InstallDeployStatusExecuting, "executing deploy")
	_, err = job.AwaitExecuteJob(ctx, &job.ExecuteJobRequest{
		RunnerID:   install.RunnerGroup.Runners[0].ID,
		JobID:      runnerJob.ID,
		WorkflowID: fmt.Sprintf("event-loop-%s-execute-job-%s", install.ID, runnerJob.ID),
	})
	if err != nil {
		w.updateDeployStatus(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to execute runner job")
		w.writeDeployEvent(ctx, installDeploy.ID, signals.OperationDeploy, app.OperationStatusFailed)
		l.Error("job did not succeed", zap.Error(err))
		return fmt.Errorf("unable to get install: %w", err)
	}

	return nil
}
