package components

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/plan"
	pkgplan "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/plan"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/job"
)

func (w *Workflows) execApplyPlan(ctx workflow.Context, install *app.Install, installDeploy *app.InstallDeploy, installWorkflowStepID string, sandboxMode bool) error {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}

	w.updateDeployStatus(ctx, installDeploy.ID, app.InstallDeployStatusPlanning, "creating deploy plan")
	l.Info("creating execution plan")

	build, err := activities.AwaitGetComponentBuildByComponentBuildID(ctx, installDeploy.ComponentBuildID)
	if err != nil {
		w.updateDeployStatus(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to get component build")
		return fmt.Errorf("unable to get build: %w", err)
	}

	// get previous job
	operation := app.RunnerJobOperationTypeCreateApplyPlan
	if installDeploy.Type == app.InstallDeployTypeTeardown {
		operation = app.RunnerJobOperationTypeCreateTeardownPlan
	}
	runnerJobType := build.ComponentConfigConnection.Type.DeployJobType()
	planJob, err := activities.AwaitGetLatestJob(ctx, &activities.GetLatestJobRequest{
		OwnerID:   installDeploy.ID,
		Operation: operation,
		Group:     runnerJobType.Group(),
		Type:      runnerJobType,
	})
	if err != nil {
		return errors.Wrap(err, "unable to plan runner job for current apply job")
	}

	logStreamID, err := cctx.GetLogStreamIDWorkflow(ctx)
	if err != nil {
		return err
	}

	defer func() {
		activities.AwaitCloseLogStreamByLogStreamID(ctx, logStreamID)
	}()

	op := app.RunnerJobOperationTypeApplyPlan
	jobTyp := build.ComponentConfigConnection.Type.DeployJobType()

	runnerJob, err := activities.AwaitCreateDeployJob(ctx, &activities.CreateDeployJobRequest{
		RunnerID:        install.RunnerID,
		DeployID:        installDeploy.ID,
		Op:              op,
		Type:            jobTyp,
		LogStreamID:     logStreamID,
		TimeoutDuration: build.ComponentConfigConnection.GetDeployTimeout(),
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

	// NOTE(jm): this is probably going to need to be refactored
	plan, err := plan.AwaitCreateDeployPlan(ctx, &plan.CreateDeployPlanRequest{
		InstallDeployID: installDeploy.ID,
		InstallID:       install.ID,
		WorkflowID:      fmt.Sprintf("%s-create-apply-plan", workflow.GetInfo(ctx).WorkflowExecution.ID),
	})
	if err != nil {
		w.updateDeployStatus(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to create deploy plan")
		return errors.Wrap(err, "unable to create deploy plan")
	}

	if err := activities.AwaitUpdateInstallWorkflowStepTarget(ctx, activities.UpdateInstallWorkflowStepTargetRequest{
		StepID:         installWorkflowStepID,
		StepTargetID:   installDeploy.ID,
		StepTargetType: plugins.TableName(w.db, installDeploy),
	}); err != nil {
		return errors.Wrap(err, "unable to update install workflow")
	}

	// Add Plan contents from the result to the plan
	if runnerJob.Type == app.RunnerJobTypeJobNOOPDeploy {
		plan.ApplyPlanContents = ""
		plan.ApplyPlanDisplay = ""
	} else if len(planJob.Execution.Result.Contents) > 0 {
		l.Info("using the legacy contents from the runner job execution result")
		plan.ApplyPlanContents = planJob.Execution.Result.Contents
		plan.ApplyPlanDisplay = string(planJob.Execution.Result.ContentsDisplay)
	} else if len(planJob.Execution.Result.ContentsGzip) > 0 {
		l.Info("using the compressed contents from the runner job execution result")
		applyPlanContents, err := planJob.Execution.Result.GetContentsB64String()
		if err != nil {
			return errors.Wrap(err, "unable to get contents string")
		}
		plan.ApplyPlanContents = applyPlanContents
		applyPlanContentsDisplay, err := planJob.Execution.Result.GetContentsDisplayString()
		if err != nil {
			return errors.Wrap(err, "unable to get contents display string")
		}
		plan.ApplyPlanDisplay = applyPlanContentsDisplay
	}

	planJSON, err := json.Marshal(plan)
	if err != nil {
		w.updateDeployStatus(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to create json from deploy plan")
		return errors.Wrap(err, "unable to create json from plan")
	}

	// Get app config for role selection
	appConfig, err := activities.AwaitGetAppConfigByID(ctx, install.AppConfigID)
	if err != nil {
		w.updateDeployStatus(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to get app config")
		return fmt.Errorf("unable to get app config: %w", err)
	}

	// Get install stack for auth configuration
	stack, err := activities.AwaitGetInstallStackByInstallID(ctx, install.ID)
	if err != nil {
		w.updateDeployStatus(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to get install stack")
		return errors.Wrap(err, "unable to get install stack")
	}

	// Get install state for role name rendering
	installState, err := activities.AwaitGetInstallState(ctx, &activities.GetInstallStateRequest{
		InstallID: install.ID,
	})
	if err != nil {
		w.updateDeployStatus(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to get install state")
		return fmt.Errorf("unable to get install state: %w", err)
	}

	// Create composite plan
	compositePlan := plantypes.CompositePlan{
		DeployPlan: plan,
	}

	// Select role for this deploy operation
	roleSelection, opn, err := w.getRoleForDeploy(
		l,
		appConfig,
		installDeploy,
		build,
		&build.ComponentConfigConnection.Component,
		stack,
		installState,
	)
	if err != nil {
		l.Error("unable to evaluate role for component operation", zap.Error(err))
		w.updateDeployStatus(
			ctx,
			installDeploy.ID,
			app.InstallDeployStatusError,
			"unable to select role",
		)
		return errors.Wrap(err, "unable to evaluate role for component deploy")
	}

	l.Info("selected role for component deploy apply",
		zap.String("component_name", build.ComponentConfigConnection.Component.Name),
		zap.String("role_name", roleSelection.RoleName),
		zap.String("role_arn", roleSelection.RoleARN),
		zap.String("source", string(roleSelection.Source)),
		zap.String("operation", string(opn)),
		zap.String("deploy_type", string(installDeploy.Type)),
	)

	// Create auth configuration for the plan
	planAuth, err := pkgplan.CreatePlanAuth(
		stack.InstallStackOutputs,
		roleSelection.RoleARN,
		fmt.Sprintf("install-deploy-%s", installDeploy.ID),
	)
	if err != nil {
		l.Error("unable to build plan auth for component operation", zap.Error(err))
		w.updateDeployStatus(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to create auth config")
		return errors.Wrap(err, "unable to create plan auth")
	}
	compositePlan.Auth = planAuth

	if err := activities.AwaitSaveRunnerJobPlan(ctx, &activities.SaveRunnerJobPlanRequest{
		JobID:         runnerJob.ID,
		PlanJSON:      string(planJSON),
		CompositePlan: compositePlan,
	}); err != nil {
		w.updateDeployStatus(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to store runner job plan")
		return fmt.Errorf("unable to get install: %w", err)
	}

	planJSON = nil
	plan = nil

	w.updateDeployStatus(ctx, installDeploy.ID, app.InstallDeployStatusExecuting, "executing deploy plan")
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

	return nil
}
