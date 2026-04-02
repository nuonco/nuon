package v2

import (
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/actionworkflowrun"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/awaitrunnerhealthy"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/componentdeployapplyplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/componentdeploysyncandplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/componentsyncimage"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/executeactionworkflow"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
)

func getComponentLifecycleActionsSteps(ctx workflow.Context, flw *app.Workflow, comp *app.Component, installID string, triggerTyp app.ActionWorkflowTriggerType, sg *stepGroup) ([]*app.WorkflowStep, error) {
	steps := make([]*app.WorkflowStep, 0)
	installActions, err := activities.AwaitGetInstallActionWorkflowsByTriggerType(ctx, activities.GetInstallActionWorkflowsByTriggerTypeRequest{
		ComponentID: comp.ID,
		InstallID:   installID,
		TriggerType: triggerTyp,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to get action workflows")
	}

	install, err := activities.AwaitGetByInstallID(ctx, installID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install")
	}

	appCfg, err := activities.AwaitGetAppConfigByID(ctx, install.AppConfigID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get current app config")
	}

	awcMap := make(map[string]app.ActionWorkflowConfig, len(appCfg.ActionWorkflowConfigs))
	for _, awc := range appCfg.ActionWorkflowConfigs {
		awcMap[awc.ActionWorkflowID] = awc
	}

	for _, installAction := range installActions {
		if _, ok := awcMap[installAction.ActionWorkflowID]; !ok {
			// skip actions that are not part of current app config
			continue
		}

		sig := &executeactionworkflow.Signal{
			Signal: &actionworkflowrun.Signal{
				InstallID:               installID,
				InstallActionWorkflowID: installAction.ID,
				TriggerType:             triggerTyp,
				TriggeredByID:           flw.ID,
				TriggeredByType:         string(triggerTyp),
				RunEnvVars: map[string]string{
					"TRIGGER_TYPE":   string(triggerTyp),
					"COMPONENT_ID":   comp.ID,
					"COMPONENT_NAME": comp.Name,
				},
			},
		}
		name := fmt.Sprintf("%s Action Run (%s)", installAction.ActionWorkflow.Name, triggerTyp)
		step, err := sg.installSignalStep(ctx, installID, name, pgtype.Hstore{}, sig, flw.PlanOnly)
		if err != nil {
			return nil, err
		}

		steps = append(steps, step)
	}

	return steps, nil
}

func getLifecycleActionsSteps(ctx workflow.Context, installID string, flw *app.Workflow, triggerTyp app.ActionWorkflowTriggerType, sg *stepGroup) ([]*app.WorkflowStep, error) {
	steps := make([]*app.WorkflowStep, 0)

	installActions, err := activities.AwaitGetInstallActionWorkflowsByTriggerType(ctx, activities.GetInstallActionWorkflowsByTriggerTypeRequest{
		InstallID:   installID,
		TriggerType: triggerTyp,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to get action workflows")
	}

	install, err := activities.AwaitGetByInstallID(ctx, installID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install")
	}

	appCfg, err := activities.AwaitGetAppConfigByID(ctx, install.AppConfigID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get current app config")
	}

	awcMap := make(map[string]app.ActionWorkflowConfig, len(appCfg.ActionWorkflowConfigs))
	for _, awc := range appCfg.ActionWorkflowConfigs {
		awcMap[awc.ActionWorkflowID] = awc
	}

	sg.nextGroup() // lifecycleSteps

	for _, installAction := range installActions {
		if _, ok := awcMap[installAction.ActionWorkflowID]; !ok {
			// skip actions that are not part of current app config
			continue
		}

		sig := &executeactionworkflow.Signal{
			Signal: &actionworkflowrun.Signal{
				InstallID:               installID,
				InstallActionWorkflowID: installAction.ID,
				TriggerType:             triggerTyp,
				TriggeredByID:           flw.ID,
				TriggeredByType:         string(triggerTyp),
				RunEnvVars: map[string]string{
					"TRIGGER_TYPE": string(triggerTyp),
					"FLOW_TYPE":    string(flw.Type),
					"FLOW_ID":      flw.ID,
					// TODO(sdboyer) remove these once they're updated on the other end
					"INSTALL_WORKFLOW_TYPE": string(flw.Type),
					"INSTALL_WORKFLOW_ID":   flw.ID,
				},
			},
		}
		name := fmt.Sprintf("%s Action Run (%s)", installAction.ActionWorkflow.Name, triggerTyp)
		step, err := sg.installSignalStep(ctx, installID, name, pgtype.Hstore{}, sig, flw.PlanOnly)
		if err != nil {
			return nil, err
		}

		steps = append(steps, step)
	}

	return steps, nil
}

func getComponentDeploySteps(ctx workflow.Context, installID string, flw *app.Workflow, componentIDs []string, sg *stepGroup) ([]*app.WorkflowStep, error) {
	steps := make([]*app.WorkflowStep, 0)

	install, err := activities.AwaitGetByInstallID(ctx, installID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install")
	}

	appcfg, err := activities.AwaitGetAppConfigByID(ctx, install.AppConfigID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get app config")
	}
	components := make(map[string]app.Component)
	for _, ccc := range appcfg.ComponentConfigConnections {
		components[ccc.ComponentID] = ccc.Component
	}

	for _, compID := range componentIDs {
		sg.nextGroup()
		comp, has := components[compID]
		if !has {
			return nil, errors.Errorf("component %s not found in app config", compID)
		}

		// Resolve install component ID
		installComp, err := activities.AwaitGetInstallComponent(ctx, activities.GetInstallComponentRequest{
			InstallID:   installID,
			ComponentID: compID,
		})
		if err != nil {
			return nil, errors.Wrap(err, "unable to get install component")
		}

		var installComponentID string
		if installComp != nil {
			installComponentID = installComp.ID
		}

		if !flw.PlanOnly {
			preDeploySteps, err := getComponentLifecycleActionsSteps(ctx, flw, &comp, installID, app.ActionWorkflowTriggerTypePreDeployComponent, sg)
			if err != nil {
				return nil, err
			}
			steps = append(steps, preDeploySteps...)
		}

		// sync image
		if comp.Type.IsImage() && !flw.PlanOnly {
			deployStep, err := sg.installSignalStep(ctx, installID, "sync "+comp.Name, pgtype.Hstore{}, &componentsyncimage.Signal{
				InstallComponentID: installComponentID,
				ComponentID:        comp.ID,
			}, flw.PlanOnly)
			if err != nil {
				return nil, errors.Wrap(err, "unable to create image sync")
			}

			steps = append(steps, deployStep)
		} else {
			if flw.PlanOnly && comp.Type == app.ComponentTypeExternalImage || comp.Type == app.ComponentTypeDockerBuild {
				continue
			}

			planStep, err := sg.installSignalStep(ctx, installID, "sync and plan "+comp.Name, pgtype.Hstore{}, &componentdeploysyncandplan.Signal{
				InstallComponentID: installComponentID,
				ComponentID:        comp.ID,
			}, flw.PlanOnly, WithSkippable(false))
			if err != nil {
				return nil, errors.Wrap(err, "unable to create image sync")
			}

			applyPlanStep, err := sg.installSignalStep(ctx, installID, "apply "+comp.Name, pgtype.Hstore{}, &componentdeployapplyplan.Signal{
				InstallComponentID: installComponentID,
				ComponentID:        comp.ID,
			}, flw.PlanOnly)
			if err != nil {
				return nil, errors.Wrap(err, "unable to create image sync")
			}
			if flw.PlanOnly {
				steps = append(steps, planStep)
			} else {
				steps = append(steps, planStep, applyPlanStep)
			}
		}
		if !flw.PlanOnly {
			postDeploySteps, err := getComponentLifecycleActionsSteps(ctx, flw, &comp, installID, app.ActionWorkflowTriggerTypePostDeployComponent, sg)
			if err != nil {
				return nil, err
			}
			steps = append(steps, postDeploySteps...)
		}
	}

	return steps, nil
}

func deployAllComponents(ctx workflow.Context, installID string, flw *app.Workflow, sg *stepGroup) ([]*app.WorkflowStep, error) {
	install, err := activities.AwaitGetByInstallID(ctx, installID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install")
	}

	componentIDs, err := activities.AwaitGetAppGraph(ctx, activities.GetAppGraphRequest{
		InstallID: install.ID,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install graph")
	}

	steps := make([]*app.WorkflowStep, 0)

	sg.nextGroup() // runner health

	step, err := sg.installSignalStep(ctx, installID, "await runner healthy", pgtype.Hstore{}, &awaitrunnerhealthy.Signal{
		InstallID: installID,
	}, flw.PlanOnly)
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	var lifecycleSteps []*app.WorkflowStep
	if !flw.PlanOnly {
		lifecycleSteps, err = getLifecycleActionsSteps(ctx, installID, flw, app.ActionWorkflowTriggerTypePreDeployAllComponents, sg)
		if err != nil {
			return nil, err
		}
		steps = append(steps, lifecycleSteps...)
	}
	deploySteps, err := getComponentDeploySteps(ctx, installID, flw, componentIDs, sg)
	if err != nil {
		return nil, err
	}
	steps = append(steps, deploySteps...)
	if !flw.PlanOnly {
		lifecycleSteps, err = getLifecycleActionsSteps(ctx, installID, flw, app.ActionWorkflowTriggerTypePostDeployAllComponents, sg)
		if err != nil {
			return nil, err
		}
		steps = append(steps, lifecycleSteps...)
	}

	return steps, nil
}
