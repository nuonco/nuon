package v2

import (
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/actionworkflowrun"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/componentdeployapplyplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/componentdeploysyncandplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/componentsyncimage"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/componentteardownapplyplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/componentteardownsyncandplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/deprovisionsandboxapplyplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/deprovisionsandboxplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/executeactionworkflow"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/generatestate"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/reprovisionsandboxapplyplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/reprovisionsandboxplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	dbgenerics "github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
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
		case app.RunbookStepTypeComponentDeploy:
			deploySteps, err := runbookDeploySteps(ctx, installID, &stepCfg, sg, flw)
			if err != nil {
				return nil, errors.Wrapf(err, "unable to generate deploy step %s", stepCfg.Name)
			}
			steps = append(steps, deploySteps...)

		case app.RunbookStepTypeComponentTearDown:
			tdSteps, err := runbookTearDownSteps(ctx, installID, &stepCfg, sg, flw)
			if err != nil {
				return nil, errors.Wrapf(err, "unable to generate tear-down step %s", stepCfg.Name)
			}
			steps = append(steps, tdSteps...)

		case app.RunbookStepTypeAction:
			actionStep, err := runbookActionStep(ctx, installID, &stepCfg, flw, sg)
			if err != nil {
				return nil, errors.Wrapf(err, "unable to generate action step %s", stepCfg.Name)
			}
			steps = append(steps, actionStep)

		case app.RunbookStepTypeSandboxReprovision,
			app.RunbookStepTypeSandboxDeprovision:
			sbxSteps, err := runbookSandboxLifecycleSteps(ctx, installID, &stepCfg, flw, sg, install)
			if err != nil {
				return nil, errors.Wrapf(err, "unable to generate sandbox %s step %s", stepCfg.Type, stepCfg.Name)
			}
			steps = append(steps, sbxSteps...)
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

	if stepCfg.DeployDependents {
		// Forward graph walk from the target component returns the component itself
		// plus its transitive dependents (downstream subgraph), in dependency order.
		componentIDs, err := activities.AwaitGetAppComponentGraph(ctx, activities.GetAppComponentGraphRequest{
			InstallID:   installID,
			ComponentID: component.ID,
		})
		if err != nil {
			return nil, errors.Wrap(err, "unable to get component dependents graph")
		}

		for _, compID := range componentIDs {
			depSteps, err := runbookDeploySingleComponent(ctx, installID, compID, stepCfg.Name, sg, flw)
			if err != nil {
				return nil, errors.Wrapf(err, "unable to deploy dependent %s", compID)
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

func runbookTearDownSteps(ctx workflow.Context, installID string, stepCfg *app.RunbookStepConfig, sg *stepGroup, flw *app.Workflow) ([]*app.WorkflowStep, error) {
	component, err := activities.AwaitGetComponentByName(ctx, activities.GetComponentByNameRequest{
		InstallID:     installID,
		ComponentName: stepCfg.ComponentName,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "unable to find component %s", stepCfg.ComponentName)
	}

	result := make([]*app.WorkflowStep, 0)

	if stepCfg.TearDownDependents {
		// Reverse forward graph walk: target + transitive dependents, with dependents listed first
		// so children tear down before parents.
		componentIDs, err := activities.AwaitGetAppComponentGraph(ctx, activities.GetAppComponentGraphRequest{
			InstallID:   installID,
			ComponentID: component.ID,
			Reverse:     true,
		})
		if err != nil {
			return nil, errors.Wrap(err, "unable to get component dependents graph")
		}

		for _, compID := range componentIDs {
			depSteps, err := runbookTearDownSingleComponent(ctx, installID, compID, stepCfg.Name, sg, flw)
			if err != nil {
				return nil, errors.Wrapf(err, "unable to tear down dependent %s", compID)
			}
			result = append(result, depSteps...)
		}
	} else {
		steps, err := runbookTearDownSingleComponent(ctx, installID, component.ID, stepCfg.Name, sg, flw)
		if err != nil {
			return nil, err
		}
		result = append(result, steps...)
	}

	return result, nil
}

func runbookTearDownSingleComponent(ctx workflow.Context, installID, componentID, stepName string, sg *stepGroup, flw *app.Workflow) ([]*app.WorkflowStep, error) {
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

	// Image components have no infra to destroy; emit a placeholder skip step for visibility.
	if component.Type.IsImage() {
		sg.nextGroupNamed(fmt.Sprintf("tear down: %s (skipped)", name))
		skipStep, err := sg.installSignalStep(ctx, installID, fmt.Sprintf("skipped image tear down %s", name), pgtype.Hstore{
			"reason": generics.ToPtr("skipped image tear down"),
		}, nil, flw.PlanOnly)
		if err != nil {
			return nil, errors.Wrap(err, "unable to create skip step")
		}
		result = append(result, skipStep)
		return result, nil
	}

	sg.nextGroupNamed(fmt.Sprintf("tear down: %s", name))
	planStep, err := sg.installSignalStep(ctx, installID, fmt.Sprintf("sync and plan tear down %s", name), pgtype.Hstore{}, &componentteardownsyncandplan.Signal{
		InstallComponentID: installComponentID,
		InstallID:          installID,
		ComponentID:        componentID,
		Role:               flw.Role,
	}, flw.PlanOnly, WithSkippable(false))
	if err != nil {
		return nil, err
	}

	applyStep, err := sg.installSignalStep(ctx, installID, fmt.Sprintf("apply tear down %s", name), pgtype.Hstore{}, &componentteardownapplyplan.Signal{
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

func runbookSandboxLifecycleSteps(ctx workflow.Context, installID string, stepCfg *app.RunbookStepConfig, flw *app.Workflow, sg *stepGroup, install *app.Install) ([]*app.WorkflowStep, error) {
	sandbox, err := activities.AwaitGetInstallSandboxByInstallID(ctx, installID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install sandbox")
	}

	role := stepCfg.Role
	if role == "" {
		role = flw.Role
	}

	var (
		planLabel   string
		applyLabel  string
		planSignal  signal.Signal
		applySignal signal.Signal
	)
	switch stepCfg.Type {
	case app.RunbookStepTypeSandboxReprovision:
		planLabel = fmt.Sprintf("sandbox reprovision plan: %s", stepCfg.Name)
		applyLabel = fmt.Sprintf("sandbox reprovision apply: %s", stepCfg.Name)
		planSignal = &reprovisionsandboxplan.Signal{InstallSandboxID: sandbox.ID, InstallID: installID, Role: role}
		applySignal = &reprovisionsandboxapplyplan.Signal{InstallSandboxID: sandbox.ID, InstallID: installID}
	case app.RunbookStepTypeSandboxDeprovision:
		planLabel = fmt.Sprintf("sandbox deprovision plan: %s", stepCfg.Name)
		applyLabel = fmt.Sprintf("sandbox deprovision apply: %s", stepCfg.Name)
		planSignal = &deprovisionsandboxplan.Signal{InstallSandboxID: sandbox.ID, InstallID: installID, Role: role}
		applySignal = &deprovisionsandboxapplyplan.Signal{InstallSandboxID: sandbox.ID, InstallID: installID}
	default:
		return nil, errors.Errorf("unsupported sandbox lifecycle step type %q", stepCfg.Type)
	}

	sg.nextGroupNamed(fmt.Sprintf("sandbox %s: %s", strings.TrimPrefix(string(stepCfg.Type), "sandbox_"), stepCfg.Name))

	result := make([]*app.WorkflowStep, 0, 2)
	planStep, err := sg.installSignalStep(ctx, installID, planLabel, pgtype.Hstore{}, planSignal, flw.PlanOnly, WithSkippable(false))
	if err != nil {
		return nil, err
	}
	result = append(result, planStep)

	if flw.PlanOnly {
		return result, nil
	}

	applyStep, err := sg.installSignalStep(ctx, installID, applyLabel, pgtype.Hstore{}, applySignal, flw.PlanOnly, WithMaxAutoRetries(install.AppSandboxConfig.GetMaxAutoRetries()))
	if err != nil {
		return nil, err
	}
	result = append(result, applyStep)

	// Optionally redeploy all components after a sandbox (re)provision, mirroring the standalone workflows.
	// Deprovision never deploys; SkipComponentDeploys lets the runbook author opt out.
	if stepCfg.SkipComponentDeploys || stepCfg.Type == app.RunbookStepTypeSandboxDeprovision {
		return result, nil
	}

	appCfg, err := activities.AwaitGetAppConfigByID(ctx, install.AppConfigID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get app config")
	}
	awData, err := activities.AwaitGetActionWorkflows(ctx, &activities.GetActionWorkflows{InstallID: installID})
	if err != nil {
		return nil, errors.Wrap(err, "unable to get action workflows")
	}

	deploySteps, err := deployAllComponents(ctx, installID, flw, sg, appCfg, awData)
	if err != nil {
		return nil, errors.Wrap(err, "unable to generate component deploy steps")
	}
	result = append(result, deploySteps...)

	return result, nil
}
