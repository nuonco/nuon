package v2

import (
	"fmt"
	"sort"

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

// filterActionWorkflowsByTrigger filters pre-fetched install action workflows by trigger type,
// optionally scoped to a specific component. It uses the version-pinned configs from
// appCfg.ActionWorkflowConfigs (with Triggers preloaded) instead of fetching latest configs.
// This replaces individual AwaitGetInstallActionWorkflowsByTriggerType activity calls
// with in-memory filtering.
func filterActionWorkflowsByTrigger(installActionWorkflows []*app.InstallActionWorkflow, triggerTyp app.ActionWorkflowTriggerType, componentID string, appCfg *app.AppConfig) []*app.InstallActionWorkflow {
	awcMap := make(map[string]app.ActionWorkflowConfig, len(appCfg.ActionWorkflowConfigs))
	for _, awc := range appCfg.ActionWorkflowConfigs {
		awcMap[awc.ActionWorkflowID] = awc
	}

	indices := map[string]int{}
	wkflows := make(map[string]*app.InstallActionWorkflow, len(installActionWorkflows))

	for _, wf := range installActionWorkflows {
		cfg, ok := awcMap[wf.ActionWorkflowID]
		if !ok {
			continue
		}

		if componentID == "" {
			if cfg.HasTrigger(triggerTyp) {
				wkflows[wf.ID] = wf
				indices[wf.ID] = cfg.GetTriggerIndex(triggerTyp)
			}
		} else {
			if cfg.HasComponentTrigger(triggerTyp, componentID) {
				wkflows[wf.ID] = wf
				indices[wf.ID] = cfg.GetComponentTriggerIndex(triggerTyp, componentID)
			}
		}
	}

	workflowIDs := make([]string, 0, len(indices))
	for wkflowID := range indices {
		workflowIDs = append(workflowIDs, wkflowID)
	}

	sort.SliceStable(workflowIDs, func(i, j int) bool {
		return indices[workflowIDs[i]] < indices[workflowIDs[j]]
	})

	result := make([]*app.InstallActionWorkflow, 0, len(workflowIDs))
	for _, wkflowID := range workflowIDs {
		if wf, ok := wkflows[wkflowID]; ok {
			result = append(result, wf)
		}
	}

	return result
}

func getComponentLifecycleActionsSteps(ctx workflow.Context, flw *app.Workflow, comp *app.Component, installID string, triggerTyp app.ActionWorkflowTriggerType, sg *stepGroup, appCfg *app.AppConfig, awData []*app.InstallActionWorkflow) ([]*app.WorkflowStep, error) {
	steps := make([]*app.WorkflowStep, 0)
	installActions := filterActionWorkflowsByTrigger(awData, triggerTyp, comp.ID, appCfg)

	for _, installAction := range installActions {
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
				Role: flw.Role,
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

func getLifecycleActionsSteps(ctx workflow.Context, installID string, flw *app.Workflow, triggerTyp app.ActionWorkflowTriggerType, sg *stepGroup, appCfg *app.AppConfig, awData []*app.InstallActionWorkflow) ([]*app.WorkflowStep, error) {
	steps := make([]*app.WorkflowStep, 0)
	installActions := filterActionWorkflowsByTrigger(awData, triggerTyp, "", appCfg)

	if len(installActions) == 0 {
		return steps, nil
	}

	sg.nextGroup() // lifecycleSteps

	for _, installAction := range installActions {
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
				Role: flw.Role,
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

func getComponentDeploySteps(ctx workflow.Context, installID string, flw *app.Workflow, componentIDs []string, sg *stepGroup, appCfg *app.AppConfig, awData []*app.InstallActionWorkflow) ([]*app.WorkflowStep, error) {
	steps := make([]*app.WorkflowStep, 0)

	components := make(map[string]app.Component)
	for _, ccc := range appCfg.ComponentConfigConnections {
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
			preDeploySteps, err := getComponentLifecycleActionsSteps(ctx, flw, &comp, installID, app.ActionWorkflowTriggerTypePreDeployComponent, sg, appCfg, awData)
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
				Role:               flw.Role,
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
				Role:               flw.Role,
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
			postDeploySteps, err := getComponentLifecycleActionsSteps(ctx, flw, &comp, installID, app.ActionWorkflowTriggerTypePostDeployComponent, sg, appCfg, awData)
			if err != nil {
				return nil, err
			}
			steps = append(steps, postDeploySteps...)
		}
	}

	return steps, nil
}

func deployAllComponents(ctx workflow.Context, installID string, flw *app.Workflow, sg *stepGroup, appCfg *app.AppConfig, awData []*app.InstallActionWorkflow) ([]*app.WorkflowStep, error) {
	componentIDs, err := activities.AwaitGetAppGraph(ctx, activities.GetAppGraphRequest{
		InstallID: installID,
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
		lifecycleSteps, err = getLifecycleActionsSteps(ctx, installID, flw, app.ActionWorkflowTriggerTypePreDeployAllComponents, sg, appCfg, awData)
		if err != nil {
			return nil, err
		}
		steps = append(steps, lifecycleSteps...)
	}
	deploySteps, err := getComponentDeploySteps(ctx, installID, flw, componentIDs, sg, appCfg, awData)
	if err != nil {
		return nil, err
	}
	steps = append(steps, deploySteps...)
	if !flw.PlanOnly {
		lifecycleSteps, err = getLifecycleActionsSteps(ctx, installID, flw, app.ActionWorkflowTriggerTypePostDeployAllComponents, sg, appCfg, awData)
		if err != nil {
			return nil, err
		}
		steps = append(steps, lifecycleSteps...)
	}

	return steps, nil
}
