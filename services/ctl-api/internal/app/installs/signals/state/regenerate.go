package state

import (
	"time"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/pkg/metrics"
	pkgstate "github.com/nuonco/nuon/pkg/types/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	installactivities "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
	state "github.com/nuonco/nuon/services/ctl-api/internal/pkg/state"
)

func Regenerate(ctx workflow.Context, req *state.ExecuteRegenerationRequest) (*state.ExecuteRegenerationResponse, error) {
	lastModifiedAt := make(map[state.PartialName]time.Time, len(req.LastModifiedAt))
	for k, v := range req.LastModifiedAt {
		lastModifiedAt[k] = v
	}

	type group struct {
		targets []state.PartialTarget
	}
	groups := make(map[state.PartialName]*group)

	for _, t := range req.Targets {
		g := groups[t.Name]
		if g == nil {
			g = &group{}
			groups[t.Name] = g
		}
		g.targets = append(g.targets, t)
	}

	is := req.CachedState
	var appID, appName, appConfigID string
	if is == nil {
		existing, err := installactivities.AwaitGetLatestInstallStateByInstallID(ctx, req.InstallID)
		if err != nil {
			return nil, errors.Wrap(err, "unable to get latest install state")
		}
		if existing != nil {
			is = existing
		} else {
			is = pkgstate.New()
		}
		install, err := installactivities.AwaitGetByInstallID(ctx, req.InstallID)
		if err != nil {
			return nil, errors.Wrap(err, "unable to get install")
		}
		is.ID = install.ID
		is.Name = install.Name
		appID = install.AppID
		appConfigID = install.AppConfigID
		appName = install.App.Name
	}

	var updatedPartials []state.PartialName
	for _, partial := range state.AllPartials {
		g, ok := groups[partial]
		if !ok {
			continue
		}
		partialStart := time.Now()
		if err := fetchPartialWithTargets(ctx, req.InstallID, partial, g.targets, is); err != nil {
			return nil, errors.Wrapf(err, "error fetching partial %s", partial)
		}
		if req.MetricsWriter != nil {
			dur := time.Since(partialStart)
			req.MetricsWriter.Timing("nuon.state.regenerate.partial.fetch.duration",
				dur,
				metrics.ToTags(map[string]string{
					"install_id":    req.InstallID,
					"partial":       string(partial),
					"app_id":        appID,
					"app_name":      appName,
					"app_config_id": appConfigID,
				}),
			)
		}
		lastModifiedAt[partial] = workflow.Now(ctx)
		updatedPartials = append(updatedPartials, partial)
	}

	if len(updatedPartials) > 0 {
		helpers.MapLegacyFields(is)

		if _, err := installactivities.AwaitSaveState(ctx, &installactivities.SaveStateRequest{
			State:           is,
			InstallID:       req.InstallID,
			TriggeredByID:   req.TriggeredByID,
			TriggeredByType: req.TriggeredByType,
			GeneratedBy:     app.InstallStateGenerateSourceStateManager,
		}); err != nil {
			return nil, errors.Wrap(err, "error while saving state")
		}

		if err := installactivities.AwaitArchiveState(ctx, &installactivities.ArchiveStateRequest{
			InstallID: req.InstallID,
		}); err != nil {
			return nil, errors.Wrap(err, "error while archiving state")
		}
	}

	return &state.ExecuteRegenerationResponse{
		State:           is,
		UpdatedPartials: updatedPartials,
		LastModifiedAt:  lastModifiedAt,
		GeneratedAt:     workflow.Now(ctx),
		AppID:           appID,
		AppName:         appName,
		AppConfigID:     appConfigID,
	}, nil
}

func fetchPartialWithTargets(ctx workflow.Context, installID string, partial state.PartialName, targets []state.PartialTarget, is *pkgstate.State) error {
	entityIDs := collectEntityIDs(targets)
	switch partial {
	case state.PartialOrg:
		return fetchOrgPartial(ctx, installID, is)
	case state.PartialApp:
		return fetchAppPartial(ctx, installID, is)
	case state.PartialDomain:
		return fetchDomainPartial(ctx, installID, is)
	case state.PartialRunner:
		return fetchRunnerPartial(ctx, installID, is)
	case state.PartialCloud:
		return fetchCloudPartial(ctx, installID, is)
	case state.PartialActions:
		return fetchActionsPartial(ctx, installID, entityIDs, is)
	case state.PartialInputs:
		return fetchInputsPartial(ctx, installID, is)
	case state.PartialComponents:
		return fetchComponentsPartial(ctx, installID, entityIDs, is)
	case state.PartialSandbox:
		return fetchSandboxPartial(ctx, installID, is)
	case state.PartialStack:
		return fetchStackPartial(ctx, installID, is)
	case state.PartialSecrets:
		return fetchSecretsPartial(ctx, installID, is)
	default:
		return errors.Errorf("unknown partial: %s", partial)
	}
}

func collectEntityIDs(targets []state.PartialTarget) []string {
	var ids []string
	for _, t := range targets {
		if t.EntityID != "" {
			ids = append(ids, t.EntityID)
		}
	}
	return ids
}

func fetchOrgPartial(ctx workflow.Context, installID string, is *pkgstate.State) error {
	org, err := installactivities.AwaitGetOrgByInstallID(ctx, installID)
	if err != nil {
		return errors.Wrap(err, "unable to get org")
	}
	is.Org = helpers.ToOrgState(*org)
	return nil
}

func fetchAppPartial(ctx workflow.Context, installID string, is *pkgstate.State) error {
	install, err := installactivities.AwaitGetByInstallID(ctx, installID)
	if err != nil {
		return errors.Wrap(err, "unable to get install")
	}
	is.App = helpers.ToAppState(install.App)
	return nil
}

func fetchDomainPartial(ctx workflow.Context, installID string, is *pkgstate.State) error {
	sandboxRun, err := installactivities.AwaitGetInstallSandboxRunStateByInstallID(ctx, installID)
	if err != nil {
		if generics.IsGormErrRecordNotFound(err) {
			is.Domain = helpers.ToDomainState(nil)
			return nil
		}
		return errors.Wrap(err, "unable to get domain state")
	}
	is.Domain = helpers.ToDomainState(sandboxRun)
	return nil
}

func fetchRunnerPartial(ctx workflow.Context, installID string, is *pkgstate.State) error {
	install, err := installactivities.AwaitGetByInstallID(ctx, installID)
	if err != nil {
		return errors.Wrap(err, "unable to get install")
	}
	runner, err := installactivities.AwaitGetRunnerByID(ctx, install.RunnerID)
	if err != nil {
		return errors.Wrap(err, "unable to get runner")
	}
	is.Runner = helpers.ToRunnerState(*runner)
	return nil
}

func fetchCloudPartial(ctx workflow.Context, installID string, is *pkgstate.State) error {
	install, err := installactivities.AwaitGetByInstallID(ctx, installID)
	if err != nil {
		return errors.Wrap(err, "unable to get install")
	}
	is.Cloud = helpers.ToCloudAccount(install)
	return nil
}

func fetchActionsPartial(ctx workflow.Context, installID string, entityIDs []string, is *pkgstate.State) error {
	if len(entityIDs) > 0 {
		return fetchTargetedActionsPartial(ctx, entityIDs, is)
	}
	return fetchAllActionsPartial(ctx, installID, is)
}

func fetchAllActionsPartial(ctx workflow.Context, installID string, is *pkgstate.State) error {
	actions, err := installactivities.AwaitGetActionWorkflowsByInstallID(ctx, installID)
	if err != nil {
		return errors.Wrap(err, "unable to get actions")
	}
	st := pkgstate.NewActionsState()
	st.Populated = true
	for _, action := range actions {
		act, err := installactivities.AwaitGetInstallActionWorkflowStateByInstallActionWorkflowID(ctx, action.ID)
		if err != nil {
			return errors.Wrap(err, "unable to get action workflow state")
		}
		st.Workflows[action.ActionWorkflow.Name] = helpers.ToActionWorkflowState(*act)
	}
	is.Actions = st
	return nil
}

func fetchTargetedActionsPartial(ctx workflow.Context, entityIDs []string, is *pkgstate.State) error {
	if is.Actions == nil {
		is.Actions = pkgstate.NewActionsState()
		is.Actions.Populated = true
	}
	for _, id := range entityIDs {
		act, err := installactivities.AwaitGetInstallActionWorkflowStateByInstallActionWorkflowID(ctx, id)
		if err != nil {
			return errors.Wrapf(err, "unable to get action workflow state %s", id)
		}
		is.Actions.Workflows[act.ActionWorkflow.Name] = helpers.ToActionWorkflowState(*act)
	}
	return nil
}

func fetchInputsPartial(ctx workflow.Context, installID string, is *pkgstate.State) error {
	inst, err := installactivities.AwaitGetByInstallID(ctx, installID)
	if err != nil {
		return errors.Wrap(err, "unable to get install")
	}
	inps, err := installactivities.AwaitGetInstallInputsStateByInstallID(ctx, installID)
	if err != nil {
		if generics.IsGormErrRecordNotFound(err) {
			is.Inputs = &pkgstate.InputsState{}
			return nil
		}
		return errors.Wrap(err, "unable to get inputs state")
	}
	cfg, err := installactivities.AwaitGetAppConfigByID(ctx, inst.AppConfigID)
	if err != nil {
		return errors.Wrap(err, "unable to get app config")
	}
	is.Inputs = helpers.ToInputState(inps, cfg, false)
	return nil
}

func fetchComponentsPartial(ctx workflow.Context, installID string, entityIDs []string, is *pkgstate.State) error {
	if len(entityIDs) > 0 {
		return fetchTargetedComponentsPartial(ctx, entityIDs, is)
	}
	return fetchAllComponentsPartial(ctx, installID, is)
}

func fetchAllComponentsPartial(ctx workflow.Context, installID string, is *pkgstate.State) error {
	installComps, err := installactivities.AwaitGetInstallComponentIDsByInstallID(ctx, installID)
	if err != nil {
		return errors.Wrap(err, "unable to get install components")
	}
	newMap := make(map[string]any, len(installComps))
	for _, instCmpID := range installComps {
		installComp, err := installactivities.AwaitGetInstallComponentStateByInstallComponentID(ctx, instCmpID)
		if err != nil {
			return errors.Wrap(err, "unable to get install component state")
		}
		cMap, err := pkgstate.AsMap(buildComponentState(installComp))
		if err != nil {
			return errors.Wrap(err, "unable to create component map")
		}
		newMap[installComp.Component.Name] = cMap
	}
	is.Components = newMap
	return nil
}

func fetchTargetedComponentsPartial(ctx workflow.Context, entityIDs []string, is *pkgstate.State) error {
	if is.Components == nil {
		is.Components = make(map[string]any)
	}
	for _, id := range entityIDs {
		installComp, err := installactivities.AwaitGetInstallComponentStateByInstallComponentID(ctx, id)
		if err != nil {
			return errors.Wrapf(err, "unable to get install component state %s", id)
		}
		cMap, err := pkgstate.AsMap(buildComponentState(installComp))
		if err != nil {
			return errors.Wrapf(err, "unable to create component map %s", id)
		}
		is.Components[installComp.Component.Name] = cMap
	}
	return nil
}

func buildComponentState(installComp *app.InstallComponent) *pkgstate.ComponentState {
	return helpers.ToComponentState(*installComp)
}

func fetchSandboxPartial(ctx workflow.Context, installID string, is *pkgstate.State) error {
	sandboxRun, err := installactivities.AwaitGetInstallSandboxRunStateByInstallID(ctx, installID)
	if err != nil {
		if generics.IsGormErrRecordNotFound(err) {
			is.Sandbox = &pkgstate.SandboxState{}
			return nil
		}
		return errors.Wrap(err, "unable to get sandbox run")
	}
	is.Sandbox = helpers.ToSandboxState(sandboxRun)
	return nil
}

func fetchStackPartial(ctx workflow.Context, installID string, is *pkgstate.State) error {
	stack, err := installactivities.AwaitGetInstallStackStateByInstallID(ctx, installID)
	if err != nil {
		return errors.Wrap(err, "unable to get stack")
	}
	is.InstallStack = helpers.ToInstallStackState(stack)
	return nil
}

func fetchSecretsPartial(ctx workflow.Context, installID string, is *pkgstate.State) error {
	runnerJob, err := installactivities.AwaitGetSecretsSyncJobByInstallID(ctx, installID)
	if err != nil {
		if generics.IsGormErrRecordNotFound(err) {
			is.Secrets = helpers.ToSecretsState(nil)
			return nil
		}
		return errors.Wrap(err, "unable to get secrets state")
	}
	is.Secrets = helpers.ToSecretsState(runnerJob.ParsedOutputs)
	return nil
}
