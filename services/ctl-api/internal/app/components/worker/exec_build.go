package worker

import (
	"encoding/json"
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/plugins/configs"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/components/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/components/worker/plan"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/controlplanejob"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/job"
)

func (w *Workflows) execBuild(ctx workflow.Context, compID, buildID string, currentApp *app.App, sandboxMode bool) error {
	comp, err := activities.AwaitGetComponent(ctx, activities.GetComponentRequest{
		ComponentID: compID,
	})
	if err != nil {
		w.updateBuildStatus(ctx, buildID, app.ComponentBuildStatusError, "unable to get component")
		return fmt.Errorf("unable to get component: %w", err)
	}

	logStreamID, err := cctx.GetLogStreamIDWorkflow(ctx)
	if err != nil {
		return err
	}

	// Create the runner job early so it appears in the dashboard even if policy evaluation fails
	runnerJob, err := activities.AwaitCreateBuildJob(ctx, &activities.CreateBuildJobRequest{
		BuildID:     buildID,
		Op:          app.RunnerJobOperationTypeBuild,
		Type:        comp.Type.BuildJobType(),
		LogStreamID: logStreamID,
		Metadata: map[string]string{
			"component_id":       comp.ID,
			"component_build_id": buildID,
			"component_name":     comp.Name,
			"app_id":             currentApp.ID,
		},
	})
	if err != nil {
		w.updateBuildStatus(ctx, buildID, app.ComponentBuildStatusError, "unable to create job")
		return fmt.Errorf("unable to create job: %w", err)
	}
	if runnerJob.RunnerID == "" {
		if runnerJob.Executor != app.RunnerJobExecutorControlPlane {
			w.updateBuildStatus(ctx, buildID, app.ComponentBuildStatusError, "no runners available in runner group")
			w.updateJobStatusForPolicyFailure(ctx, runnerJob.ID, "no runners available in runner group")
			return fmt.Errorf("no runners available in runner group for org %s", comp.Org.ID)
		}
	}

	if comp.Type == app.ComponentTypeExternalImage {
		if err := w.evaluateExternalImagePolicy(ctx, buildID, runnerJob.ID, runnerJob.RunnerID, comp.Name); err != nil {
			return err
		}
	}

	runPlan, err := plan.AwaitCreateComponentBuildPlan(ctx, &plan.CreateComponentBuildPlanRequest{
		ComponentID:      comp.ID,
		ComponentBuildID: buildID,
		WorkflowID:       fmt.Sprintf("%s-create-build-plan", workflow.GetInfo(ctx).WorkflowExecution.ID),
		CloudProvider:    w.cfg.CloudProvider,
	})
	if err != nil {
		w.updateBuildStatus(ctx, buildID, app.ComponentBuildStatusError, "unable to get component build plan")
		return errors.Wrap(err, "unable to create plan")
	}

	if runPlan.ContainerImagePullPlan != nil &&
		runPlan.ContainerImagePullPlan.RepoCfg != nil &&
		runPlan.ContainerImagePullPlan.RepoCfg.RegistryType == configs.OCIRegistryTypeGAR {
		garToken, err := activities.AwaitGetGARAccessToken(ctx, &activities.GetGARAccessTokenRequest{
			ServiceAccountEmail:      runPlan.ContainerImagePullPlan.RepoCfg.ServiceAccountEmail,
			WorkloadIdentityProvider: runPlan.ContainerImagePullPlan.RepoCfg.WorkloadIdentityProvider,
		})
		if err != nil {
			return errors.Wrap(err, "unable to get GAR access token")
		}
		runPlan.ContainerImagePullPlan.RepoCfg.OCIAuth = &configs.OCIRegistryAuth{
			Username: garToken.Username,
			Password: garToken.Password,
		}
	}

	planJSON, err := json.Marshal(runPlan)
	if err != nil {
		return errors.Wrap(err, "unable to create json")
	}
	if err != nil {
		w.updateBuildStatus(ctx, buildID, app.ComponentBuildStatusError, "unable to get convert build plan to JSON")
		return fmt.Errorf("unable to convert plan to json: %w", err)
	}

	if err := activities.AwaitSaveRunnerJobPlan(ctx, &activities.SaveRunnerJobPlanRequest{
		JobID:    runnerJob.ID,
		PlanJSON: string(planJSON),
		CompositePlan: plantypes.CompositePlan{
			BuildPlan: runPlan,
		},
	}); err != nil {
		w.updateBuildStatus(ctx, buildID, app.ComponentBuildStatusError, "unable to save job plan")
		return fmt.Errorf("unable to save runner job plan: %w", err)
	}

	// wait for the job
	w.updateBuildStatus(ctx, buildID, app.ComponentBuildStatusBuilding, "building")
	if runnerJob.Executor == app.RunnerJobExecutorControlPlane {
		err = controlplanejob.AwaitExecuteControlPlaneJob(ctx, &controlplanejob.ExecuteRequest{JobID: runnerJob.ID}, &workflow.ChildWorkflowOptions{
			WorkflowID: fmt.Sprintf("control-plane-%s-execute-job-%s", comp.ID, runnerJob.ID),
		})
	} else {
		_, err = job.AwaitExecuteJob(ctx, &job.ExecuteJobRequest{
			RunnerID:   runnerJob.RunnerID,
			JobID:      runnerJob.ID,
			WorkflowID: fmt.Sprintf("queue-signal-%s-execute-job-%s", comp.ID, runnerJob.ID),
		})
	}
	if err != nil {
		w.updateBuildStatus(ctx, buildID, app.ComponentBuildStatusError, "build did not complete successfully")
		return fmt.Errorf("build job failed: %w", err)
	}

	if comp.Type == app.ComponentTypeHelmChart {
		if err := w.evaluateHelmBuildPolicy(ctx, buildID, runnerJob.ID, comp.Name); err != nil {
			return err
		}
	}

	w.updateBuildStatus(ctx, buildID, app.ComponentBuildStatusActive, "build is active and ready to be deployed")
	return nil
}
