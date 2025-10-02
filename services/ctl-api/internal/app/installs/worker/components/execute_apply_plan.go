package components

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	plantypes "github.com/powertoolsdev/mono/pkg/plans/types"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/plan"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/log"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/workflows/job"
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

	op := app.RunnerJobOperationTypeApplyPlan
	jobTyp := build.ComponentConfigConnection.Type.DeployJobType()

	runnerJob, err := activities.AwaitCreateDeployJob(ctx, &activities.CreateDeployJobRequest{
		RunnerID:    install.RunnerID,
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

	if err := activities.AwaitSaveRunnerJobPlan(ctx, &activities.SaveRunnerJobPlanRequest{
		JobID:    runnerJob.ID,
		PlanJSON: string(planJSON),
		CompositePlan: plantypes.CompositePlan{
			DeployPlan: plan,
		},
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
