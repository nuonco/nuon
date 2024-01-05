package worker

import (
	"fmt"

	componentsv1 "github.com/powertoolsdev/mono/pkg/types/components/component/v1"
	execv1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/execute/v1"
	planv1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/plan/v1"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
	"go.temporal.io/sdk/workflow"
)

func (w *Workflows) deploy(ctx workflow.Context, installID, deployID string, sandboxMode bool) error {
	var install app.Install
	if err := w.defaultExecGetActivity(ctx, w.acts.Get, activities.GetRequest{
		InstallID: installID,
	}, &install); err != nil {
		w.updateDeployStatus(ctx, deployID, StatusError, "unable to get install from database")
		return fmt.Errorf("unable to get install: %w", err)
	}

	var installDeploy app.InstallDeploy
	if err := w.defaultExecGetActivity(ctx, w.acts.GetDeploy, activities.GetDeployRequest{
		DeployID: deployID,
	}, &installDeploy); err != nil {
		w.updateDeployStatus(ctx, deployID, StatusError, "unable to get install deploy from database")
		return fmt.Errorf("unable to get install deploy: %w", err)
	}

	var deployCfg componentsv1.Component
	if err := w.defaultExecGetActivity(ctx, w.acts.GetComponentConfig, activities.GetComponentConfigRequest{
		DeployID: deployID,
	}, &deployCfg); err != nil {
		w.updateDeployStatus(ctx, deployID, StatusError, "unable to get component config")
		return fmt.Errorf("unable to get deploy component config: %w", err)
	}

	w.updateDeployStatus(ctx, deployID, StatusPlanning, "creating sync plan")

	// execute the plan phase here
	syncImagePlanWorkflowID := fmt.Sprintf("%s-sync-plan-%s", installID, deployID)
	planResp, err := w.execCreatePlanWorkflow(ctx, sandboxMode, syncImagePlanWorkflowID, &planv1.CreatePlanRequest{
		Input: &planv1.CreatePlanRequest_Component{
			Component: &planv1.ComponentInput{
				OrgId:     install.App.OrgID,
				AppId:     install.App.ID,
				BuildId:   installDeploy.ComponentBuildID,
				InstallId: install.ID,
				DeployId:  deployID,
				Component: &deployCfg,
				Type:      planv1.ComponentInputType_COMPONENT_INPUT_TYPE_WAYPOINT_SYNC_IMAGE,
			},
		},
	})
	if err != nil {
		w.updateDeployStatus(ctx, deployID, StatusError, fmt.Sprintf("unable to create sync plan: %s", err))
		return fmt.Errorf("unable to create sync plan: %w", err)
	}

	w.updateDeployStatus(ctx, deployID, StatusSyncing, "executing sync plan")

	syncExecuteWorkflowID := fmt.Sprintf("%s-sync-execute-%s", installID, deployID)
	_, err = w.execExecPlanWorkflow(ctx, sandboxMode, syncExecuteWorkflowID, &execv1.ExecutePlanRequest{
		Plan: planResp.Plan,
	})
	if err != nil {
		w.updateDeployStatus(ctx, deployID, StatusError, fmt.Sprintf("unable to execute sync plan: %s", err))
		return fmt.Errorf("unable to execute sync plan: %w", err)
	}

	w.updateDeployStatus(ctx, deployID, StatusPlanning, "creating deploy plan")

	// execute the plan phase here
	deployPlanWorkflowID := fmt.Sprintf("%s-deploy-plan-%s", installID, deployID)
	planResp, err = w.execCreatePlanWorkflow(ctx, sandboxMode, deployPlanWorkflowID, &planv1.CreatePlanRequest{
		Input: &planv1.CreatePlanRequest_Component{
			Component: &planv1.ComponentInput{
				OrgId:     install.App.OrgID,
				AppId:     install.App.ID,
				InstallId: install.ID,
				BuildId:   installDeploy.ComponentBuildID,
				DeployId:  deployID,
				Component: &deployCfg,
				Type:      planv1.ComponentInputType_COMPONENT_INPUT_TYPE_WAYPOINT_DEPLOY,
			},
		},
	})
	if err != nil {
		w.updateDeployStatus(ctx, deployID, StatusError, fmt.Sprintf("unable to create deploy plan: %s", err))
		return fmt.Errorf("unable to create deploy plan: %w", err)
	}

	w.updateDeployStatus(ctx, deployID, StatusExecuting, "executing deploy plan")

	deployExecuteWorkflowID := fmt.Sprintf("%s-deploy-execute-%s", installID, deployID)
	_, err = w.execExecPlanWorkflow(ctx, sandboxMode, deployExecuteWorkflowID, &execv1.ExecutePlanRequest{
		Plan: planResp.Plan,
	})
	if err != nil {
		w.updateDeployStatus(ctx, deployID, StatusError, fmt.Sprintf("unable to execute deploy plan: %s", err))
		return fmt.Errorf("unable to execute deploy plan: %w", err)
	}

	w.updateDeployStatus(ctx, deployID, StatusActive, "deploy is active")
	return nil
}
