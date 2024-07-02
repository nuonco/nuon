package worker

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	componentsv1 "github.com/powertoolsdev/mono/pkg/types/components/component/v1"
	execv1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/execute/v1"
	planv1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/plan/v1"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/components/worker/activities"
)

func (w *Workflows) build(ctx workflow.Context, cmpID, buildID string, dryRun bool) error {
	w.updateBuildStatus(ctx, buildID, app.ComponentBuildStatusPlanning, "creating build plan")

	var currentApp app.App
	if err := w.defaultExecGetActivity(ctx, w.acts.GetComponentApp, activities.GetComponentAppRequest{
		ComponentID: cmpID,
	}, &currentApp); err != nil {
		w.updateBuildStatus(ctx, buildID, app.ComponentBuildStatusError, "unable to get component app")
		return fmt.Errorf("unable to get component app: %w", err)
	}

	var buildCfg componentsv1.Component
	if err := w.defaultExecGetActivity(ctx, w.acts.GetComponentConfig, activities.GetRequest{
		BuildID: buildID,
	}, &buildCfg); err != nil {
		w.updateBuildStatus(ctx, buildID, app.ComponentBuildStatusError, "unable to get component config")
		return fmt.Errorf("unable to get build component config: %w", err)
	}

	// execute the plan phase here
	buildPlanWorkflowID := fmt.Sprintf("%s-build-plan-%s", cmpID, buildID)
	planResp, err := w.execCreatePlanWorkflow(ctx, dryRun, buildPlanWorkflowID, &planv1.CreatePlanRequest{
		Input: &planv1.CreatePlanRequest_Component{
			Component: &planv1.ComponentInput{
				OrgId:     currentApp.OrgID,
				AppId:     currentApp.ID,
				BuildId:   buildID,
				Component: &buildCfg,
				Type:      planv1.ComponentInputType_COMPONENT_INPUT_TYPE_WAYPOINT_BUILD,
				Context:   w.protos.BuildContext(),
			},
		},
	})
	if err != nil {
		w.updateBuildStatus(ctx, buildID, app.ComponentBuildStatusError, "unable to create build plan")
		return fmt.Errorf("unable to execute build plan: %w", err)
	}

	// update status with response
	w.updateBuildStatus(ctx, buildID, app.ComponentBuildStatusBuilding, "executing build plan")

	// execute the exec phase here
	buildExecuteWorkflowID := fmt.Sprintf("%s-build-execute-%s", cmpID, buildID)
	_, err = w.execExecPlanWorkflow(ctx, dryRun, buildExecuteWorkflowID, &execv1.ExecutePlanRequest{
		Plan: planResp.Plan,
	})
	if err != nil {
		w.updateBuildStatus(ctx, buildID, app.ComponentBuildStatusError, "unable to execute build plan")
		return fmt.Errorf("unable to execute build plan: %w", err)
	}

	w.updateBuildStatus(ctx, buildID, app.ComponentBuildStatusActive, "build is active and ready to be deployed")
	return nil
}
