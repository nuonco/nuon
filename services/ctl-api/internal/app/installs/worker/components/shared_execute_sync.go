package components

import (
	"encoding/json"
	"fmt"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"

	plantypes "github.com/powertoolsdev/mono/pkg/plans/types"
	"github.com/powertoolsdev/mono/pkg/types/state"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/plan"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/log"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/workflows/job"
)

func (w *Workflows) execSync(ctx workflow.Context, install *app.Install, installDeploy *app.InstallDeploy, sandboxMode bool) error {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}

	l.Info("syncing image into install OCI repository")
	w.updateDeployStatusWithoutStatusSync(ctx, installDeploy.ID, app.InstallDeployStatusPlanning, "creating sync plan")

	build, err := activities.AwaitGetComponentBuildByComponentBuildID(ctx, installDeploy.ComponentBuildID)
	if err != nil {
		w.updateDeployStatusWithoutStatusSync(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to get component build")
		return fmt.Errorf("unable to get build: %w", err)
	}

	logStreamID, err := cctx.GetLogStreamIDWorkflow(ctx)
	if err != nil {
		return err
	}

	runnerJob, err := activities.AwaitCreateSyncJob(ctx, &activities.CreateSyncJobRequest{
		DeployID:    installDeploy.ID,
		RunnerID:    install.RunnerID,
		Op:          app.RunnerJobOperationTypeExec,
		Type:        build.ComponentConfigConnection.Type.SyncJobType(),
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
		w.updateDeployStatusWithoutStatusSync(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to create runner job")
		return fmt.Errorf("unable to create runner job: %w", err)
	}

	// create the plan request
	runPlan, err := plan.AwaitCreateSyncPlan(ctx, &plan.CreateSyncPlanRequest{
		InstallID:       install.ID,
		InstallDeployID: installDeploy.ID,
		WorkflowID:      fmt.Sprintf("%s-create-oci-sync-plan", workflow.GetInfo(ctx).WorkflowExecution.ID),
	})
	if err != nil {
		w.updateDeployStatusWithoutStatusSync(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to store runner job plan")
		return errors.Wrap(err, "unable to create plan")
	}

	planJSON, err := json.Marshal(runPlan)
	if err != nil {
		return errors.Wrap(err, "unable to create json")
	}

	// Deprecated: for now we dual write both the plan json and the composite plan
	if err := activities.AwaitSaveRunnerJobPlan(ctx, &activities.SaveRunnerJobPlanRequest{
		JobID:    runnerJob.ID,
		PlanJSON: string(planJSON),
		CompositePlan: plantypes.CompositePlan{
			SyncOCIPlan: runPlan,
		},
	}); err != nil {
		w.updateDeployStatusWithoutStatusSync(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to store runner job plan")
		return fmt.Errorf("unable to get install: %w", err)
	}

	// queue job
	w.updateDeployStatusWithoutStatusSync(ctx, installDeploy.ID, app.InstallDeployStatusSyncing, "executing sync plan")
	_, err = job.AwaitExecuteJob(ctx, &job.ExecuteJobRequest{
		RunnerID:   install.RunnerID,
		JobID:      runnerJob.ID,
		WorkflowID: fmt.Sprintf("%s-execute-job", workflow.GetInfo(ctx).WorkflowExecution.ID),
	})
	if err != nil {
		w.updateDeployStatusWithoutStatusSync(ctx, installDeploy.ID, app.InstallDeployStatusError, "unable to poll job")
		l.Error("error polling sync image job", zap.Error(err))
		return fmt.Errorf("unable to poll job: %w", err)
	}
	l.Info("sync image job was successfully completed")

	// parse outputs
	job, err := activities.AwaitGetJobByID(ctx, runnerJob.ID)
	if err != nil {
		return errors.Wrap(err, "unable to get runner job")
	}

	var ociArtOutputs state.OCIArtifactOutputs
	if err := mapstructure.Decode(job.ParsedOutputs["image"], &ociArtOutputs); err != nil {
		l.Error("error parsing oci artifact outputs", zap.Error(err))
		return errors.Wrap(err, "unable to parse oci artifact outputs")
	}

	if _, err := activities.AwaitCreateOCIArtifact(ctx, activities.CreateOCIArtifactRequest{
		OwnerID:   installDeploy.ID,
		OwnerType: plugins.TableName(w.db, installDeploy),
		Outputs:   ociArtOutputs,
	}); err != nil {
		return errors.Wrap(err, "unable to create oci artifact")
	}

	return nil
}
