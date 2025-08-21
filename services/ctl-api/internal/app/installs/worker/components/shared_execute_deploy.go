package components

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/plan"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/log"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/workflows/job"
)

func (w *Workflows) execPlan(ctx workflow.Context, install *app.Install, installDeploy *app.InstallDeploy, stepID string, sandboxMode bool) error {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}

	w.updateDeployStatus(ctx, installDeploy.ID, app.InstallDeployStatusPlanning, "creating deploy plan")
	l.Info("planning and executing deploy")

	build, err := activities.AwaitGetComponentBuildByComponentBuildID(ctx, installDeploy.ComponentBuildID)
	if err != nil {
		w.updateDeployStatus(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to get component build")
		return fmt.Errorf("unable to get build: %w", err)
	}

	comp, err := activities.AwaitGetComponentByComponentID(ctx, build.ComponentID)
	if err != nil {
		return errors.Wrap(err, "unable to get component")
	}

	logStreamID, err := cctx.GetLogStreamIDWorkflow(ctx)
	if err != nil {
		return err
	}

	op := app.RunnerJobOperationTypeCreateTeardownPlan
	if installDeploy.Type == app.InstallDeployTypeApply {
		op = app.RunnerJobOperationTypeCreateApplyPlan
	}

	jobTyp := build.ComponentConfigConnection.Type.DeployPlanJobType()
	runnerJob, err := activities.AwaitCreateDeployJob(ctx, &activities.CreateDeployJobRequest{
		RunnerID:    install.RunnerGroup.Runners[0].ID,
		DeployID:    installDeploy.ID,
		Op:          op,
		Type:        jobTyp,
		LogStreamID: logStreamID,
		Metadata: map[string]string{
			"install_id":           install.ID,
			"deploy_id":            installDeploy.ID,
			"install_component_id": installDeploy.InstallComponentID,
			"component_id":         build.ComponentConfigConnection.ComponentID,
			"component_name":       build.ComponentConfigConnection.Component.Name,
			"operation":            string(op),
		},
	})
	if err != nil {
		w.updateDeployStatus(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to create runner job")
		return fmt.Errorf("unable to create runner job: %w", err)
	}

	plan, err := plan.AwaitCreateDeployPlan(ctx, &plan.CreateDeployPlanRequest{
		InstallDeployID: installDeploy.ID,
		InstallID:       install.ID,
		WorkflowID:      fmt.Sprintf("%s-create-deploy-plan", workflow.GetInfo(ctx).WorkflowExecution.ID),
	})
	if err != nil {
		w.updateDeployStatus(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to create deploy plan")
		return errors.Wrap(err, "unable to create deploy plan")
	}

	planJSON, err := json.Marshal(plan)
	if err != nil {
		w.updateDeployStatus(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to create json from deploy plan")
		return errors.Wrap(err, "unable to create json from plan")
	}

	if err := activities.AwaitSaveRunnerJobPlan(ctx, &activities.SaveRunnerJobPlanRequest{
		JobID:    runnerJob.ID,
		PlanJSON: string(planJSON),
	}); err != nil {
		w.updateDeployStatus(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to store runner job plan")
		return fmt.Errorf("unable to get install: %w", err)
	}

	planJSON = nil
	plan = nil

	w.updateDeployStatus(ctx, installDeploy.ID, app.InstallDeployStatusExecuting, "creating plan")
	_, err = job.AwaitExecuteJob(ctx, &job.ExecuteJobRequest{
		RunnerID:   install.RunnerID,
		JobID:      runnerJob.ID,
		WorkflowID: fmt.Sprintf("event-loop-%s-execute-job-%s", install.ID, runnerJob.ID),
	})
	if err != nil {
		w.updateDeployStatus(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to execute runner job")
		l.Error("job did not succeed", zap.Error(err))
		return fmt.Errorf("unable to get install: %w", err)
	}

	job, err := activities.AwaitGetJobByID(ctx, runnerJob.ID)
	if err != nil {
		return errors.Wrap(err, "unable to get job")
	}

	var approvalTyp app.WorkflowStepApprovalType
	switch comp.Type {
	case app.ComponentTypeHelmChart:
		approvalTyp = app.HelmApprovalApprovalType
	case app.ComponentTypeTerraformModule:
		approvalTyp = app.TerraformPlanApprovalType
	case app.ComponentTypeKubernetesManifest:
		approvalTyp = app.KubernetesManifestApprovalType
	default:
		approvalTyp = app.NoopApprovalType
	}

	// all component types require a plan EXCEPT fro docker builds
	if job.Execution.Result == nil || (len(job.Execution.Result.Contents) < 1 && len(job.Execution.Result.ContentsGzip) < 1) {
		if runnerJob.Type != app.RunnerJobTypeJobNOOPDeploy {
			return errors.New("no plan returned from job")
		}
	}

	var contentsDisplay string
	if runnerJob.Type == app.RunnerJobTypeJobNOOPDeploy {
		l.Debug("job is a noop - no plan is expected")
		contentsDisplay = "{\"op\": \"noop\"}"
	} else {
		l.Debug(fmt.Sprintf("getting job execution result from result:%s", job.Execution.Result.ID))
		contentsDisplay, err = job.Execution.Result.GetContentsDisplayString()
		if err != nil {
			return errors.Wrap(err, "unable to get content display")
		}
	}

	if _, err := activities.AwaitCreateStepApproval(ctx, &activities.CreateStepApprovalRequest{
		OwnerID:     installDeploy.ID,
		OwnerType:   "install_deploys",
		RunnerJobID: job.ID,
		StepID:      stepID,
		Plan:        contentsDisplay,
		Type:        approvalTyp,
	}); err != nil {
		return errors.Wrap(err, "unable to create approval")
	}

	return nil
}
