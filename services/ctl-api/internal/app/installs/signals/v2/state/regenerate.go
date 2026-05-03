package state

import (
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"gorm.io/gorm"

	pkggenerics "github.com/nuonco/nuon/pkg/generics"
	pkgstate "github.com/nuonco/nuon/pkg/types/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	installactivities "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
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

	if req.ForceAll {
		for _, t := range req.Targets {
			g := groups[t.Name]
			if g == nil {
				g = &group{}
				groups[t.Name] = g
			}
			g.targets = append(g.targets, t)
		}
	} else {
		checked := make(map[state.PartialName]bool)
		for _, t := range req.Targets {
			if !checked[t.Name] {
				checked[t.Name] = true
				resp, err := installactivities.AwaitCheckModified(ctx, &installactivities.CheckModifiedRequest{
					InstallID:   req.InstallID,
					PartialName: string(t.Name),
					LastKnownAt: lastModifiedAt[t.Name],
				})
				if err != nil {
					return nil, errors.Wrapf(err, "check modified %s", t.Name)
				}
				if !resp.Changed {
					continue
				}
				groups[t.Name] = &group{}
			}
			if g := groups[t.Name]; g != nil {
				g.targets = append(g.targets, t)
			}
		}
	}

	if len(groups) == 0 {
		return &state.ExecuteRegenerationResponse{
			State:          req.CachedState,
			LastModifiedAt: lastModifiedAt,
			GeneratedAt:    workflow.Now(ctx),
		}, nil
	}

	is := req.CachedState
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
	}

	var updatedPartials []state.PartialName
	for _, partial := range state.AllPartials {
		g, ok := groups[partial]
		if !ok {
			continue
		}
		if err := fetchPartialWithTargets(ctx, req.InstallID, partial, g.targets, is); err != nil {
			return nil, errors.Wrapf(err, "fetch partial %s", partial)
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
			TriggeredByType: string(req.TriggeredByType),
			GeneratedBy:     app.InstallStateGenerateSourceStateManager,
		}); err != nil {
			return nil, errors.Wrap(err, "save state")
		}

		if err := installactivities.AwaitArchiveState(ctx, &installactivities.ArchiveStateRequest{
			InstallID: req.InstallID,
		}); err != nil {
			return nil, errors.Wrap(err, "archive state")
		}
	}

	return &state.ExecuteRegenerationResponse{
		State:           is,
		UpdatedPartials: updatedPartials,
		LastModifiedAt:  lastModifiedAt,
		GeneratedAt:     workflow.Now(ctx),
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
		if strings.Contains(err.Error(), gorm.ErrRecordNotFound.Error()) {
			is.Domain = &pkgstate.DomainState{}
			return nil
		}
		return errors.Wrap(err, "unable to get domain state")
	}
	st := pkgstate.NewDomainState()
	if sandboxRun != nil {
		if v, ok := sandboxRun.Outputs["public_domain"].(string); ok {
			st.PublicDomain = v
		}
		if v, ok := sandboxRun.Outputs["internal_domain"].(string); ok {
			st.InternalDomain = v
		}
	}
	is.Domain = st
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
	st := pkgstate.NewCloudAccount()
	if install.AWSAccount != nil {
		st.AWS = &pkgstate.AWSCloudAccount{Region: install.AWSAccount.Region}
	}
	if install.AzureAccount != nil {
		st.Azure = &pkgstate.AzureCloudAccount{Location: install.AzureAccount.Location}
	}
	if install.GCPAccount != nil {
		st.GCP = &pkgstate.GCPCloudAccount{
			ProjectID: install.GCPAccount.ProjectID,
			Region:    install.GCPAccount.Region,
		}
	}
	is.Cloud = st
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
		st.Workflows[action.ActionWorkflow.Name] = buildActionWorkflowState(act)
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
		is.Actions.Workflows[act.ActionWorkflow.Name] = buildActionWorkflowState(act)
	}
	return nil
}

func buildActionWorkflowState(act *app.InstallActionWorkflow) *pkgstate.ActionWorkflowState {
	return helpers.ToActionWorkflowState(*act)
}

func fetchInputsPartial(ctx workflow.Context, installID string, is *pkgstate.State) error {
	inst, err := installactivities.AwaitGetByInstallID(ctx, installID)
	if err != nil {
		return errors.Wrap(err, "unable to get install")
	}
	inps, err := installactivities.AwaitGetInstallInputsStateByInstallID(ctx, installID)
	if err != nil {
		if strings.Contains(err.Error(), gorm.ErrRecordNotFound.Error()) {
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
		if strings.Contains(err.Error(), gorm.ErrRecordNotFound.Error()) {
			is.Sandbox = &pkgstate.SandboxState{}
			return nil
		}
		return errors.Wrap(err, "unable to get sandbox run")
	}
	st := pkgstate.NewSandboxState()
	st.Populated = true
	st.Status = string(sandboxRun.Status)
	st.Outputs = sandboxRun.Outputs
	publicVCSConfig := sandboxRun.AppSandboxConfig.PublicGitVCSConfig
	connectedVCSConfig := sandboxRun.AppSandboxConfig.ConnectedGithubVCSConfig
	if publicVCSConfig != nil {
		st.Type = publicVCSConfig.Directory
		st.Version = publicVCSConfig.Branch
	}
	if connectedVCSConfig != nil {
		st.Type = connectedVCSConfig.Directory
		st.Version = connectedVCSConfig.Branch
	}
	is.Sandbox = st
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
		if strings.Contains(err.Error(), gorm.ErrRecordNotFound.Error()) {
			is.Secrets = &pkgstate.SecretsState{}
			return nil
		}
		return errors.Wrap(err, "unable to get secrets state")
	}
	var secretsState pkgstate.SecretsState
	decoderConfig := &mapstructure.DecoderConfig{
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToSliceHookFunc(","),
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToTimeHookFunc(time.RFC3339Nano),
			pkggenerics.StringToMapDecodeHook(),
		),
		WeaklyTypedInput: true,
		Result:           &secretsState,
	}
	decoder, err := mapstructure.NewDecoder(decoderConfig)
	if err != nil {
		return errors.Wrap(err, "unable to create decoder")
	}
	if err := decoder.Decode(runnerJob.ParsedOutputs); err != nil {
		return errors.Wrap(err, "unable to parse secrets outputs")
	}
	is.Secrets = &secretsState
	return nil
}
