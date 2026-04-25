package state

import (
	"strings"
	"time"

	"go.temporal.io/sdk/workflow"
	"gorm.io/gorm"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	pkggenerics "github.com/nuonco/nuon/pkg/generics"
	pkgstate "github.com/nuonco/nuon/pkg/types/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
)

// executeRegeneration regenerates the specified partials, saves to DB, and updates cached state.
func (sm *stateManager) executeRegeneration(ctx workflow.Context, partials map[PartialName]bool) ([]PartialName, error) {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return nil, err
	}

	is := sm.state.CachedState
	if is == nil {
		is = pkgstate.New()
		// Fetch install for ID/Name.
		install, err := activities.AwaitGetByInstallID(ctx, sm.installID)
		if err != nil {
			return nil, errors.Wrap(err, "unable to get install")
		}
		is.ID = install.ID
		is.Name = install.Name
	}

	var updatedPartials []PartialName

	for _, partial := range AllPartials {
		if !partials[partial] {
			continue
		}

		l.Debug("regenerating partial", zap.String("partial", string(partial)))

		if err := sm.fetchPartial(ctx, partial, is); err != nil {
			return nil, errors.Wrapf(err, "unable to fetch partial %s", partial)
		}

		sm.state.LastModifiedAt[partial] = workflow.Now(ctx)
		updatedPartials = append(updatedPartials, partial)
	}

	// Always map legacy fields if anything changed.
	if len(updatedPartials) > 0 {
		mapLegacyFields(is)
	}

	sm.state.CachedState = is

	// Persist to DB.
	if len(updatedPartials) > 0 {
		if _, err := activities.AwaitSaveState(ctx, &activities.SaveStateRequest{
			State:           is,
			InstallID:       sm.installID,
			TriggeredByID:   sm.installID,
			TriggeredByType: "state-manager",
		}); err != nil {
			return nil, errors.Wrap(err, "unable to save state")
		}

		if err := activities.AwaitArchiveState(ctx, &activities.ArchiveStateRequest{
			InstallID: sm.installID,
		}); err != nil {
			return nil, errors.Wrap(err, "unable to archive stale state")
		}
	}

	sm.state.LastGeneratedAt = workflow.Now(ctx)
	sm.state.GenerationCount++

	return updatedPartials, nil
}

// fetchPartial regenerates a single partial and sets it on the state.
func (sm *stateManager) fetchPartial(ctx workflow.Context, partial PartialName, is *pkgstate.State) error {
	switch partial {
	case PartialOrg:
		return sm.fetchOrgPartial(ctx, is)
	case PartialApp:
		return sm.fetchAppPartial(ctx, is)
	case PartialDomain:
		return sm.fetchDomainPartial(ctx, is)
	case PartialRunner:
		return sm.fetchRunnerPartial(ctx, is)
	case PartialCloud:
		return sm.fetchCloudPartial(ctx, is)
	case PartialActions:
		return sm.fetchActionsPartial(ctx, is)
	case PartialInputs:
		return sm.fetchInputsPartial(ctx, is)
	case PartialComponents:
		return sm.fetchComponentsPartial(ctx, is)
	case PartialSandbox:
		return sm.fetchSandboxPartial(ctx, is)
	case PartialStack:
		return sm.fetchStackPartial(ctx, is)
	case PartialSecrets:
		return sm.fetchSecretsPartial(ctx, is)
	default:
		return errors.Errorf("unknown partial: %s", partial)
	}
}

func (sm *stateManager) fetchOrgPartial(ctx workflow.Context, is *pkgstate.State) error {
	org, err := activities.AwaitGetOrgByInstallID(ctx, sm.installID)
	if err != nil {
		return errors.Wrap(err, "unable to get org")
	}

	st := pkgstate.NewOrgState()
	st.Populated = true
	st.ID = org.ID
	st.Name = org.Name
	st.Status = string(org.Status)
	is.Org = st
	return nil
}

func (sm *stateManager) fetchAppPartial(ctx workflow.Context, is *pkgstate.State) error {
	install, err := activities.AwaitGetByInstallID(ctx, sm.installID)
	if err != nil {
		return errors.Wrap(err, "unable to get install")
	}

	currentApp := install.App
	st := pkgstate.NewAppState()
	st.Populated = true
	st.ID = currentApp.ID
	st.Name = currentApp.Name
	st.Status = string(currentApp.Status)
	for _, secr := range currentApp.AppSecrets {
		st.Variables[secr.Name] = secr.Value
	}
	is.App = st
	return nil
}

func (sm *stateManager) fetchDomainPartial(ctx workflow.Context, is *pkgstate.State) error {
	sandboxRun, err := activities.AwaitGetInstallSandboxRunStateByInstallID(ctx, sm.installID)
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

func (sm *stateManager) fetchRunnerPartial(ctx workflow.Context, is *pkgstate.State) error {
	install, err := activities.AwaitGetByInstallID(ctx, sm.installID)
	if err != nil {
		return errors.Wrap(err, "unable to get install")
	}

	runner, err := activities.AwaitGetRunnerByID(ctx, install.RunnerID)
	if err != nil {
		return errors.Wrap(err, "unable to get runner")
	}

	st := pkgstate.NewRunnerState()
	st.Populated = true
	st.ID = runner.ID
	st.RunnerGroupID = runner.RunnerGroupID
	st.Status = string(runner.Status)
	is.Runner = st
	return nil
}

func (sm *stateManager) fetchCloudPartial(ctx workflow.Context, is *pkgstate.State) error {
	install, err := activities.AwaitGetByInstallID(ctx, sm.installID)
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

func (sm *stateManager) fetchActionsPartial(ctx workflow.Context, is *pkgstate.State) error {
	actions, err := activities.AwaitGetActionWorkflowsByInstallID(ctx, sm.installID)
	if err != nil {
		return errors.Wrap(err, "unable to get actions")
	}

	st := pkgstate.NewActionsState()
	st.Populated = true
	for _, action := range actions {
		act, err := activities.AwaitGetInstallActionWorkflowStateByInstallActionWorkflowID(ctx, action.ID)
		if err != nil {
			return errors.Wrap(err, "unable to get action workflow state")
		}

		aws := pkgstate.NewActionWorkflowState()
		aws.Populated = true
		aws.Status = string(act.Status)
		aws.ID = act.ActionWorkflow.ID
		for _, run := range act.Runs {
			if run.RunnerJob != nil {
				aws.Outputs = run.RunnerJob.ParsedOutputs
				break
			}
		}
		st.Workflows[action.ActionWorkflow.Name] = aws
	}
	is.Actions = st
	return nil
}

func (sm *stateManager) fetchInputsPartial(ctx workflow.Context, is *pkgstate.State) error {
	inst, err := activities.AwaitGetByInstallID(ctx, sm.installID)
	if err != nil {
		return errors.Wrap(err, "unable to get install")
	}

	inps, err := activities.AwaitGetInstallInputsStateByInstallID(ctx, sm.installID)
	if err != nil {
		if strings.Contains(err.Error(), gorm.ErrRecordNotFound.Error()) {
			is.Inputs = &pkgstate.InputsState{}
			return nil
		}
		return errors.Wrap(err, "unable to get inputs state")
	}

	cfg, err := activities.AwaitGetAppConfigByID(ctx, inst.AppConfigID)
	if err != nil {
		return errors.Wrap(err, "unable to get app config")
	}

	is.Inputs = toInputState(inps, cfg, false)
	return nil
}

func (sm *stateManager) fetchComponentsPartial(ctx workflow.Context, is *pkgstate.State) error {
	installComps, err := activities.AwaitGetInstallComponentIDsByInstallID(ctx, sm.installID)
	if err != nil {
		return errors.Wrap(err, "unable to get install components")
	}

	comps := pkgstate.NewComponentsState()
	comps.Populated = true

	for _, instCmpID := range installComps {
		installComp, err := activities.AwaitGetInstallComponentStateByInstallComponentID(ctx, instCmpID)
		if err != nil {
			return errors.Wrap(err, "unable to get install component state")
		}

		st := pkgstate.NewComponentState()
		st.Name = installComp.Component.Name
		st.Populated = true
		st.ComponentID = installComp.ComponentID
		st.InstallComponentID = installComp.ID
		if len(installComp.InstallDeploys) > 0 {
			st.Status = string(installComp.InstallDeploys[0].Status)
			st.BuildID = string(installComp.InstallDeploys[0].ComponentBuildID)
			st.Outputs = installComp.InstallDeploys[0].Outputs
		}
		comps.Components[installComp.Component.Name] = st
	}

	is.Components = make(map[string]any)
	for name, c := range comps.Components {
		cMap, err := pkgstate.AsMap(c)
		if err != nil {
			return errors.Wrap(err, "unable to create map")
		}
		is.Components[name] = cMap
	}
	return nil
}

func (sm *stateManager) fetchSandboxPartial(ctx workflow.Context, is *pkgstate.State) error {
	sandboxRun, err := activities.AwaitGetInstallSandboxRunStateByInstallID(ctx, sm.installID)
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

func (sm *stateManager) fetchStackPartial(ctx workflow.Context, is *pkgstate.State) error {
	stack, err := activities.AwaitGetInstallStackStateByInstallID(ctx, sm.installID)
	if err != nil {
		return errors.Wrap(err, "unable to get stack")
	}

	is.InstallStack = toInstallStackState(stack)
	return nil
}

func (sm *stateManager) fetchSecretsPartial(ctx workflow.Context, is *pkgstate.State) error {
	runnerJob, err := activities.AwaitGetSecretsSyncJobByInstallID(ctx, sm.installID)
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

// Helper functions ported from the old state package.

func toInputState(inputs *app.InstallInputs, cfg *app.AppConfig, redacted bool) *pkgstate.InputsState {
	inputValues := inputs.Values
	if redacted {
		inputValues = inputs.ValuesRedacted
	}
	if inputs == nil || len(inputValues) < 1 {
		return nil
	}

	is := pkgstate.NewInputsState()
	for _, inp := range cfg.InputConfig.AppInputs {
		val, ok := inputValues[inp.Name]
		if !ok {
			val = &inp.Default
		}
		is.Inputs[inp.Name] = pkggenerics.FromPtrStr(val)
	}
	return is
}

func toInstallStackState(stack *app.InstallStack) *pkgstate.InstallStackState {
	if stack == nil || len(stack.InstallStackVersions) < 1 {
		return nil
	}

	is := pkgstate.NewInstallStackState()
	is.Populated = true

	version := stack.InstallStackVersions[0]
	is.QuickLinkURL = version.QuickLinkURL
	is.TemplateURL = version.TemplateURL
	is.TemplateJSON = string(version.Contents)
	is.Checksum = version.Checksum
	is.Status = string(version.Status.Status)
	is.Outputs = stack.InstallStackOutputs.DataContents
	return is
}

func mapLegacyFields(is *pkgstate.State) {
	is.Install = &pkgstate.InstallState{
		Populated: true,
		ID:        is.ID,
		Name:      is.Name,
	}
	if is.Sandbox != nil {
		is.Install.Sandbox = *is.Sandbox
	}
	if is.Domain != nil {
		is.Install.PublicDomain = is.Domain.PublicDomain
		is.Install.InternalDomain = is.Domain.InternalDomain
	}
	if is.Inputs != nil {
		is.Install.Inputs = is.Inputs.Inputs
	}
}
