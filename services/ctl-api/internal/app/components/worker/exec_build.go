package worker

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/components/worker/activities"
	runnersignals "github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/protos"
)

func (w *Workflows) execBuild(ctx workflow.Context, compID, buildID string, currentApp *app.App, sandboxMode bool) error {
	buildCfg, err := activities.AwaitGetComponentConfig(ctx, activities.GetRequest{
		BuildID: buildID,
	})
	if err != nil {
		w.updateBuildStatus(ctx, buildID, app.ComponentBuildStatusError, "unable to get component config")
		return fmt.Errorf("unable to get build component config: %w", err)
	}

	// create the sandbox plan request
	build, err := activities.AwaitGetComponentBuildByID(ctx, buildID)
	if err != nil {
		w.updateBuildStatus(ctx, buildID, app.ComponentBuildStatusError, "unable to get component build")
		return fmt.Errorf("unable to get component build: %w", err)
	}

	buildPlanWorkflowID := fmt.Sprintf("%s-build-plan-%s", compID, buildID)
	planReq := w.protos.ToBuildPlanRequest(build, buildCfg)
	planResp, err := w.execCreatePlanWorkflow(ctx, sandboxMode, buildPlanWorkflowID, planReq)
	if err != nil {
		w.updateBuildStatus(ctx, buildID, app.ComponentBuildStatusError, "unable to get component config")
		return fmt.Errorf("unable to create build plan: %w", err)
	}

	comp, err := activities.AwaitGetComponent(ctx, activities.GetComponentRequest{
		ComponentID: compID,
	})
	if err != nil {
		w.updateBuildStatus(ctx, buildID, app.ComponentBuildStatusError, "unable to get component")
		return fmt.Errorf("unable to get component: %w", err)
	}

	// create the job
	runnerJob, err := activities.AwaitCreateBuildJob(ctx, &activities.CreateBuildJobRequest{
		RunnerID: comp.Org.RunnerGroup.Runners[0].ID,
		BuildID:  buildID,
		Op:       app.RunnerJobOperationTypeBuild,
		Type:     comp.Type.BuildJobType(),
	})
	if err != nil {
		w.updateBuildStatus(ctx, buildID, app.ComponentBuildStatusError, "unable to create job")
		return fmt.Errorf("unable to create job: %w", err)
	}

	// store the plan in the db
	planJSON, err := protos.ToJSON(planResp.Plan)
	if err != nil {
		return fmt.Errorf("unable to convert plan to json: %w", err)
	}

	if err := activities.AwaitSaveRunnerJobPlan(ctx, &activities.SaveRunnerJobPlanRequest{
		JobID:    runnerJob.ID,
		PlanJSON: string(planJSON),
	}); err != nil {
		w.updateBuildStatus(ctx, buildID, app.ComponentBuildStatusError, "unable to save job plan")
		return fmt.Errorf("unable to save runner job plan: %w", err)
	}

	// queue job
	w.evClient.Send(ctx, comp.Org.RunnerGroup.Runners[0].ID, &runnersignals.Signal{
		JobID: runnerJob.ID,
		Type:  runnersignals.OperationJobQueued,
	})

	// wait for the job
	w.updateBuildStatus(ctx, buildID, app.ComponentBuildStatusBuilding, "building")
	if err := w.pollJob(ctx, runnerJob.ID); err != nil {
		w.updateBuildStatus(ctx, buildID, app.ComponentBuildStatusError, "build did not complete successfully")
		return fmt.Errorf("build job failed: %w", err)
	}

	w.updateBuildStatus(ctx, buildID, app.ComponentBuildStatusActive, "build is active and ready to be deployed")
	return nil
}
