package v2

import (
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/awaitrunnerhealthy"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/componenthelmrecover"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/generatestate"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	statemanager "github.com/nuonco/nuon/services/ctl-api/internal/pkg/state"
)

// RecoverHelmRelease unsticks a Helm release helm left mid-operation. It runs no
// plan and no approval: it applies no chart and changes no desired state, and the
// recovery itself refuses to act unless the release really is pending.
//
// Install state is still generated first, because the release name and namespace
// are templates rendered against it — recovering with stale state would target
// the wrong release.
func RecoverHelmRelease(ctx workflow.Context, flw *app.Workflow) (*app.GenerateStepsResult, error) {
	installID := generics.FromPtrStr(flw.Metadata["install_id"])

	install, err := activities.AwaitGetByInstallID(ctx, installID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install")
	}

	componentID, ok := flw.Metadata["component_id"]
	if !ok {
		return nil, errors.New("component id is not set on the install workflow for a helm release recovery")
	}
	installDeployID, ok := flw.Metadata["install_deploy_id"]
	if !ok {
		return nil, errors.New("install deploy id is not set on the install workflow for a helm release recovery")
	}

	sg := newStepGroup(flw)
	steps := make([]*app.WorkflowStep, 0)

	// One eager group holds both the state generation and the runner check, as
	// the deploy and teardown generators do. Opening a second group here would
	// leave the first one empty whenever state-gen-v2 skips the state step, and
	// an empty group stops the conductor from ever fetching the later groups.
	sg.nextGroupEager()
	orgEnabled, err := activities.AwaitHasFeatureByFeature(ctx, string(app.OrgFeatureStateGenV2))
	if err != nil {
		return nil, errors.Wrap(err, "unable to check state-gen-v2 feature")
	}
	if !statemanager.UseStateGenV2(orgEnabled, install.Metadata) {
		step, err := sg.installSignalStep(ctx, installID, "generate install state", pgtype.Hstore{},
			&generatestate.Signal{InstallID: installID}, flw.PlanOnly, WithSkippable(false))
		if err != nil {
			return nil, err
		}
		steps = append(steps, step)
	}

	step, err := sg.installSignalStep(ctx, installID, "runner healthy", pgtype.Hstore{}, &awaitrunnerhealthy.Signal{
		InstallID: installID,
	}, flw.PlanOnly)
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	compIDStr := generics.FromPtrStr(componentID)
	comp, err := activities.AwaitGetComponentByComponentID(ctx, compIDStr)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get component")
	}

	installComp, err := activities.AwaitGetInstallComponent(ctx, activities.GetInstallComponentRequest{
		InstallID:   installID,
		ComponentID: compIDStr,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install component")
	}

	sg.nextGroup() // recover
	recoverStep, err := sg.installSignalStep(ctx, installID, "recover helm release "+comp.Name, pgtype.Hstore{},
		&componenthelmrecover.Signal{
			InstallID:          installID,
			InstallComponentID: installComp.ID,
			ComponentID:        compIDStr,
			InstallDeployID:    generics.FromPtrStr(installDeployID),
		}, flw.PlanOnly, WithSkippable(false))
	if err != nil {
		return nil, err
	}
	steps = append(steps, recoverStep)

	return sg.Result(steps), nil
}
