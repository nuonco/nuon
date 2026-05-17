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

	rbConfig, err := activities.AwaitGetRunbookConfigByID(ctx, generics.FromPtrStr(runbookConfigID))
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

	if !stateGenV2 {
		sg.nextGroupEager()
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

	return sg.Result(steps)
}

func runbookDeploySteps(ctx workflow.Context, installID string, stepCfg *app.RunbookStepConfig, sg *stepGroup, flw *app.Workflow) ([]*app.WorkflowStep, error) {
	// Find the component by name
	component, err := activities.AwaitGetComponentByName(ctx, installID, stepCfg.ComponentName)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to find component %s", stepCfg.ComponentName)
	}

	// Find the install component
	installComp, err := activities.AwaitGetInstallComponent(ctx, activities.GetInstallComponentRequest{
		InstallID:   installID,
		ComponentID: component.ID,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install component")
	}

	// Create deploy record
	deploy, err := activities.AwaitCreateDeployForRunbook(ctx, installID, component.ID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create deploy")
	}

	result := make([]*app.WorkflowStep, 0)

	if component.Type.IsImage() {
		sg.nextGroupNamed(fmt.Sprintf("deploy: %s (sync)", stepCfg.Name))
		syncStep, err := sg.installSignalStep(ctx, installID, fmt.Sprintf("sync %s", stepCfg.Name), pgtype.Hstore{}, &componentsyncimage.Signal{
			InstallComponentID: installComp.ID,
			DeployID:           deploy.ID,
			ComponentID:        component.ID,
			Role:               flw.Role,
		}, flw.PlanOnly)
		if err != nil {
			return nil, err
		}
		result = append(result, syncStep)
	} else {
		sg.nextGroupNamed(fmt.Sprintf("deploy: %s (plan)", stepCfg.Name))
		planStep, err := sg.installSignalStep(ctx, installID, fmt.Sprintf("plan %s", stepCfg.Name), pgtype.Hstore{}, &componentdeploysyncandplan.Signal{
			InstallComponentID: installComp.ID,
			InstallID:          installID,
			DeployID:           deploy.ID,
			ComponentID:        component.ID,
			Role:               flw.Role,
		}, flw.PlanOnly, WithSkippable(false))
		if err != nil {
			return nil, err
		}
		result = append(result, planStep)

		if !flw.PlanOnly {
			sg.nextGroupNamed(fmt.Sprintf("deploy: %s (apply)", stepCfg.Name))
			applyStep, err := sg.installSignalStep(ctx, installID, fmt.Sprintf("apply %s", stepCfg.Name), pgtype.Hstore{}, &componentdeployapplyplan.Signal{
				InstallComponentID: installComp.ID,
				InstallID:          installID,
				ComponentID:        component.ID,
			}, flw.PlanOnly)
			if err != nil {
				return nil, err
			}
			result = append(result, applyStep)
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
		iaw, err := activities.AwaitGetInstallActionWorkflowByID(ctx, stepCfg.ActionWorkflowID.ValueString())
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
			},
		}

		return sg.installSignalStep(ctx, installID, fmt.Sprintf("action: %s", stepCfg.Name), pgtype.Hstore{}, sig, false)
	}

	// Inline ad-hoc action
	sig := &executeactionworkflow.Signal{
		Signal: &actionworkflowrun.Signal{
			InstallID:       installID,
			TriggerType:     app.ActionWorkflowTriggerTypeAdHoc,
			TriggeredByID:   triggeredByID,
			TriggeredByType: "runbook",
			Command:         stepCfg.Command,
			InlineContents:  stepCfg.InlineContents,
			Role:            stepCfg.Role,
		},
	}

	return sg.installSignalStep(ctx, installID, fmt.Sprintf("action: %s", stepCfg.Name), pgtype.Hstore{}, sig, false)
}
