package workflows

import (
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"

	"github.com/powertoolsdev/mono/pkg/config/refs"
	"github.com/powertoolsdev/mono/pkg/generics"
)

func InputUpdate(ctx workflow.Context, flw *app.Workflow) ([]*app.WorkflowStep, error) {
	installID := generics.FromPtrStr(flw.Metadata["install_id"])

	sg := newStepGroup()

	sg.nextGroup()
	steps := make([]*app.WorkflowStep, 0)
	step, err := sg.installSignalStep(ctx, installID, "generate install state", pgtype.Hstore{}, &signals.Signal{
		Type: signals.OperationGenerateState,
	}, flw.PlanOnly, WithSkippable(false))
	steps = append(steps, step)

	changedInputsRaw := generics.FromPtrStr(flw.Metadata["inputs"])
	changedInputs := strings.Split(changedInputsRaw, ",")

	install, err := activities.AwaitGetByInstallID(ctx, installID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install")
	}

	lifecycleSteps, err := getLifecycleActionsSteps(ctx, installID, flw, app.ActionWorkflowTriggerTypePreUpdateInputs, sg)
	if err != nil {
		return nil, err
	}
	steps = append(steps, lifecycleSteps...)

	appConfig, err := activities.AwaitGetAppConfig(ctx, activities.GetAppConfigRequest{
		ID: install.AppConfigID,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get app config for install %s", installID)
	}

	var changedRefs []refs.Ref
	for _, input := range changedInputs {
		changedRefs = append(changedRefs, refs.Ref{
			Name: input,
			Type: refs.RefTypeInputs,
		})
		changedRefs = append(changedRefs, refs.Ref{
			Name: input,
			Type: refs.RefTypeInstallInputs,
		})
	}

	var componentIDs []string
	for _, comp := range getComponentsForChangedInputs(appConfig, &changedRefs) {
		componentIDs = append(componentIDs, comp.ID)

		dependentCompIDs, err := activities.AwaitGetComponentDependents(ctx, &activities.GetComponentDependentsRequest{
			AppConfigID: appConfig.ID,
			ComponentID: comp.ID,
		})
		if err != nil {
			return nil, errors.Wrapf(err, "unable to get component dependents for %s", comp.ID)
		}

		componentIDs = append(componentIDs, dependentCompIDs...)
	}
	componentIDs = generics.UniqueSlice(componentIDs)

	deploySteps, err := getComponentDeploySteps(ctx, installID, flw, componentIDs, sg)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get component deploy steps")
	}
	steps = append(steps, deploySteps...)

	lifecycleSteps, err = getLifecycleActionsSteps(ctx, installID, flw, app.ActionWorkflowTriggerTypePostUpdateInputs, sg)
	if err != nil {
		return nil, err
	}
	steps = append(steps, lifecycleSteps...)

	return steps, nil
}

func getComponentsForChangedInputs(appConfig *app.AppConfig, changedRefs *[]refs.Ref) []app.Component {
	components := make([]app.Component, 0)

	for _, conConfigs := range appConfig.ComponentConfigConnections {
		for _, ref := range conConfigs.Refs {
			for _, changedRef := range *changedRefs {
				if ref.Name == changedRef.Name && ref.Type == changedRef.Type {
					components = append(components, conConfigs.Component)
				}
			}
		}
	}
	return components
}
