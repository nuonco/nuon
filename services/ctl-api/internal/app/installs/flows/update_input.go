package flows

import (
	"context"
	"slices"
	"strings"

	"github.com/pkg/errors"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/helpers"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/pkg/config/refs"
	"github.com/powertoolsdev/mono/pkg/generics"
)

func InputUpdate(ctx workflow.Context, flw *app.Workflow) ([]*app.WorkflowStep, error) {
	installID := generics.FromPtrStr(flw.Metadata["install_id"])

	changedInputsRaw := generics.FromPtrStr(flw.Metadata["inputs"])
	changeInputs := strings.Split(changedInputsRaw, ",")

	install, err := activities.AwaitGetByInstallID(ctx, installID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install")
	}

	steps := make([]*app.WorkflowStep, 0)
	lifecycleSteps, err := getLifecycleActionsSteps(ctx, installID, flw, app.ActionWorkflowTriggerTypePreUpdateInputs)
	if err != nil {
		return nil, err
	}
	steps = append(steps, lifecycleSteps...)

	installComponents, err := activities.AwaitGetInstallComponentsByInstallID(ctx, installID)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get components for app %s", install.App.ID)
	}

	var compConfigs []app.ComponentConfigConnection
	for _, ic := range installComponents {
		component, err := activities.AwaitGetComponent(ctx, activities.GetComponentRequest{
			ComponentID: ic.ComponentID,
		})
		if err != nil {
			return nil, errors.Wrapf(err, "unable to get component %s for app %s", ic.ComponentID, install.App.ID)
		}

		compConfigs = append(compConfigs, component.ComponentConfigs...)
	}

	var componentIDs []string
	for _, compID := range getComponentsForChangedInputs(compConfigs, changeInputs) {
		componentIDs = append(componentIDs, compID)
		comps, err := activities.AwaitGetComponentDependents(ctx, activities.GetComponentDependents{
			AppID:           install.App.ID,
			ComponentRootID: compID,
			ConfigVersion:   install.AppConfig.Version,
		})
		if err != nil {
			return nil, errors.Wrapf(err, "unable to get component dependents for %s", compID)
		}
		var cmpIds []string
		for _, c := range comps {
			cmpIds = append(cmpIds, c.ID)
		}
		componentIDs = append(componentIDs, cmpIds...)
	}
	componentIDs = generics.UniqueSlice(componentIDs)

	var components []app.Component
	for _, c := range installComponents {
		if slices.Contains(componentIDs, c.ComponentID) {
			components = append(components, c.Component)
		}
	}
	orderedCompIDs, err := helpers.GetDeploymentOrderFromComponents(context.TODO(), &components)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get deployment order from components")
	}

	deploySteps, err := getComponentDeploySteps(ctx, installID, flw, *orderedCompIDs)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get component deploy steps")
	}
	steps = append(steps, deploySteps...)

	lifecycleSteps, err = getLifecycleActionsSteps(ctx, installID, flw, app.ActionWorkflowTriggerTypePostUpdateInputs)
	if err != nil {
		return nil, err
	}
	steps = append(steps, lifecycleSteps...)

	return steps, nil
}

func getComponentsForChangedInputs(ccc []app.ComponentConfigConnection, changeInputs []string) []string {
	componentIDs := make([]string, 0)
	for _, conConfigs := range ccc {
		for _, ref := range conConfigs.Refs {

			if ref.Type == refs.RefTypeInputs && slices.Contains(changeInputs, ref.Name) {
				componentIDs = append(componentIDs, conConfigs.ComponentID)
			}
		}
	}
	return componentIDs
}
