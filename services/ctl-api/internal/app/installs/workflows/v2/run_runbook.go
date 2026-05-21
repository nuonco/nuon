package v2

import (
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/actionworkflowrun"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/componentdeployapplyplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/componentdeploysyncandplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/componentsyncimage"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/executeactionworkflow"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/generatestate"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	dbgenerics "github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
	statemanager "github.com/nuonco/nuon/services/ctl-api/internal/pkg/state"
)

func RunRunbook(ctx workflow.Context, flw *app.Workflow) (*app.GenerateStepsResult, error) {
	sg := newStepGroup(flw)

	installID := flw.OwnerID
	if flw.OwnerType != "installs" {
		return nil, errors.New("invalid owner set on workflow")
	}

	steps := make([]*app.WorkflowStep, 0)

	runbookConfigID, ok := flw.Metadata["runbook_config_id"]
	if !ok {
		return nil, errors.New("runbook_config_id is not set on the workflow metadata")
	}

	rbConfig, err := activities.AwaitGetRunbookConfigByID(ctx, activities.GetRunbookConfigByIDRequest{
		RunbookConfigID: generics.FromPtrStr(runbookConfigID),
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to get runbook config")
	}

	install, err := activities.AwaitGetByInstallID(ctx, installID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install")
	}

	// Generate state
	orgEnabled, err := activities.AwaitHasFeatureByFeature(ctx, string(app.OrgFeatureStateGenV2))
	if err != nil {
		return nil, errors.Wrap(err, "unable to check state-gen-v2 feature")
	}
	stateGenV2 := statemanager.UseStateGenV2(orgEnabled, install.Metadata)

	sg.nextGroupEager()
	if !stateGenV2 {
		stateStep, err := sg.installSignalStep(ctx, installID, "generate-state", pgtype.Hstore{}, &generatestate.Signal{
			InstallID: installID,
		}, false)
		if err != nil {
			return nil, errors.Wrap(err, "unable to generate state step")
		}
		steps = append(steps, stateStep)
	}

	// Generate steps for each runbook step
	for _, stepCfg := range rbConfig.Steps {
		switch stepCfg.Type {
		case app.RunbookStepTypeDeploy:
			deploySteps, err := runbookDeploySteps(ctx, installID, &stepCfg, sg, flw)
			if err != nil {
				return nil, errors.Wrapf(err, "unable to generate deploy step %s", stepCfg.Name)
			}
			steps = append(steps, deploySteps...)

		case app.RunbookStepTypeAction:
			actionStep, err := runbookActionStep(ctx, installID, &stepCfg, flw, sg)
			if err != nil {
				return nil, errors.Wrapf(err, "unable to generate action step %s", stepCfg.Name)
			}
			steps = append(steps, actionStep)
		}
	}

	return sg.Result(steps), nil
}

func runbookDeploySteps(ctx workflow.Context, installID string, stepCfg *app.RunbookStepConfig, sg *stepGroup, flw *app.Workflow) ([]*app.WorkflowStep, error) {
	// Find the primary component by name
	component, err := activities.AwaitGetComponentByName(ctx, activities.GetComponentByNameRequest{
		InstallID:     installID,
		ComponentName: stepCfg.ComponentName,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "unable to find component %s", stepCfg.ComponentName)
	}

	result := make([]*app.WorkflowStep, 0)

	if stepCfg.DeployDependencies {
		// Get the ordered dependency graph for this component
		componentIDs, err := activities.AwaitGetAppComponentGraph(ctx, activities.GetAppComponentGraphRequest{
			InstallID:   installID,
			ComponentID: component.ID,
		})
		if err != nil {
			return nil, errors.Wrap(err, "unable to get component dependency graph")
		}

		// Deploy each dependency in order (the graph includes the component itself)
		for _, compID := range componentIDs {
			depSteps, err := runbookDeploySingleComponent(ctx, installID, compID, stepCfg.Name, sg, flw)
			if err != nil {
				return nil, errors.Wrapf(err, "unable to deploy dependency %s", compID)
			}
			result = append(result, depSteps...)
		}
	} else {
		steps, err := runbookDeploySingleComponent(ctx, installID, component.ID, stepCfg.Name, sg, flw)
		if err != nil {
			return nil, err
		}
		result = append(result, steps...)
	}

	return result, nil
}

func runbookDeploySingleComponent(ctx workflow.Context, installID, componentID, stepName string, sg *stepGroup, flw *app.Workflow) ([]*app.WorkflowStep, error) {
	installComp, err := activities.AwaitGetInstallComponent(ctx, activities.GetInstallComponentRequest{
		InstallID:   installID,
		ComponentID: componentID,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install component")
	}

	var installComponentID string
	if installComp != nil {
		installComponentID = installComp.ID
	}

	component, err := activities.AwaitGetComponentByComponentID(ctx, componentID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get component")
	}

	name := fmt.Sprintf("%s/%s", stepName, component.Name)
	result := make([]*app.WorkflowStep, 0)

	if component.Type.IsImage() {
		sg.nextGroupNamed(fmt.Sprintf("deploy: %s (sync)", name))
		syncStep, err := sg.installSignalStep(ctx, installID, fmt.Sprintf("sync %s", name), pgtype.Hstore{}, &componentsyncimage.Signal{
			InstallComponentID: installComponentID,
			ComponentID:        componentID,
			Role:               flw.Role,
		}, flw.PlanOnly)
		if err != nil {
			return nil, err
		}
		result = append(result, syncStep)
	} else {
		sg.nextGroupNamed(fmt.Sprintf("deploy: %s", name))
		planStep, err := sg.installSignalStep(ctx, installID, fmt.Sprintf("sync and plan %s", name), pgtype.Hstore{}, &componentdeploysyncandplan.Signal{
			InstallComponentID: installComponentID,
			InstallID:          installID,
			ComponentID:        componentID,
			Role:               flw.Role,
		}, flw.PlanOnly, WithSkippable(false))
		if err != nil {
			return nil, err
		}

		applyStep, err := sg.installSignalStep(ctx, installID, fmt.Sprintf("apply %s", name), pgtype.Hstore{}, &componentdeployapplyplan.Signal{
			InstallComponentID: installComponentID,
			InstallID:          installID,
			ComponentID:        componentID,
		}, flw.PlanOnly)
		if err != nil {
			return nil, err
		}

		if flw.PlanOnly {
			result = append(result, planStep)
		} else {
			result = append(result, planStep, applyStep)
		}
	}

	return result, nil
}

func runbookActionStep(ctx workflow.Context, installID string, stepCfg *app.RunbookStepConfig, flw *app.Workflow, sg *stepGroup) (*app.WorkflowStep, error) {
	triggeredByID := ""
	if v, ok := flw.Metadata["triggerred_by_id"]; ok {
		triggeredByID = generics.FromPtrStr(v)
	}

	sg.nextGroupNamed(fmt.Sprintf("action: %s", stepCfg.Name))

	if stepCfg.ActionWorkflowID.ValueString() != "" {
		iaw, err := activities.AwaitGetInstallActionWorkflow(ctx, activities.GetInstallActionWorkflowRequest{
			InstallID:        installID,
			ActionWorkflowID: stepCfg.ActionWorkflowID.ValueString(),
		})
		if err != nil {
			return nil, errors.Wrap(err, "unable to get install action workflow")
		}

		sig := &executeactionworkflow.Signal{
			Signal: &actionworkflowrun.Signal{
				InstallID:               installID,
				InstallActionWorkflowID: iaw.ID,
				TriggerType:             app.ActionWorkflowTriggerTypeManual,
				TriggeredByID:           triggeredByID,
				TriggeredByType:         "runbook",
				RunEnvVars:              dbgenerics.ToStringMap(stepCfg.EnvVars),
			},
		}

		return sg.installSignalStep(ctx, installID, fmt.Sprintf("action: %s", stepCfg.Name), pgtype.Hstore{}, sig, false)
	}

	// Inline ad-hoc action: create the run record first, then reference by ID
	adHocRun, err := activities.AwaitCreateAdHocActionRunForRunbook(ctx, activities.CreateAdHocActionRunForRunbookRequest{
		InstallID:       installID,
		Command:         stepCfg.Command,
		InlineContents:  stepCfg.InlineContents,
		EnvVars:         stepCfg.EnvVars,
		Timeout:         stepCfg.Timeout,
		Role:            stepCfg.Role,
		TriggeredByID:   triggeredByID,
		TriggeredByType: "runbook",
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to create ad-hoc action run for runbook")
	}

	sig := &executeactionworkflow.Signal{
		Signal: &actionworkflowrun.Signal{
			InstallID:        installID,
			AdhocActionRunID: adHocRun.ID,
			TriggerType:      app.ActionWorkflowTriggerTypeAdHoc,
			TriggeredByID:    triggeredByID,
			TriggeredByType:  "runbook",
			Role:             stepCfg.Role,
		},
	}

	return sg.installSignalStep(ctx, installID, fmt.Sprintf("action: %s", stepCfg.Name), pgtype.Hstore{}, sig, false)
}
