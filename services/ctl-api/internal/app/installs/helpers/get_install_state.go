package helpers

import (
	"context"
	"fmt"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/sourcegraph/conc/pool"
	"gorm.io/gorm"

	pkggenerics "github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/pkg/types/outputs"
	"github.com/nuonco/nuon/pkg/types/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/views"
	pkgstate "github.com/nuonco/nuon/services/ctl-api/internal/pkg/state"
)

// GetInstallState reads the current state of the install from the DB, and returns it in a structure that can be used for variable interpolation.
func (h *Helpers) GetInstallState(ctx context.Context, installID string, redacted bool, skipVersionCheck bool) (*state.State, error) {
	latestState, err := h.getLatestInstallStateRow(ctx, installID)
	if err == nil {
		es := latestState.State
		switch {
		case !latestState.StaleAt.Empty() && len(latestState.StalePartials) > 0:
			es, err = h.regenerateStalePartials(ctx, latestState, redacted, skipVersionCheck)
			if err != nil {
				return nil, errors.Wrap(err, "unable to regenerate stale partials")
			}
		case !latestState.StaleAt.Empty():
			es.StaleAt = &latestState.StaleAt.Time
		}

		// Labels are mutable and not persisted in the state snapshot,
		// so always hydrate them fresh from the database.
		if err := h.hydrateLabels(ctx, installID, es); err != nil {
			return nil, errors.Wrap(err, "unable to hydrate labels")
		}
		return es, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.Wrap(err, "unable to get install state from db")
	}

	is := state.New()

	var (
		install      *app.Install
		appCfg       *app.AppConfig
		installComps []app.InstallComponent
		stack        *app.InstallStack
		sandboxRuns  []app.InstallSandboxRun
		actions      []app.InstallActionWorkflow
		secrets      *outputs.SyncSecretsOutput
	)

	p := pool.New().WithErrors()

	// collect all data up front e
	p.Go(func() error {
		var err error
		install, err = h.getStateInstall(ctx, installID)
		if err != nil {
			return errors.Wrap(err, "unable to get install")
		}

		return nil
	})
	p.Go(func() error {
		var err error
		install, err := h.getStateInstall(ctx, installID)
		if err != nil {
			return errors.Wrap(err, "unable to get install")
		}

		appCfg, err = h.appsHelpers.GetFullAppConfig(ctx, install.AppConfigID, skipVersionCheck)
		if err != nil {
			return errors.Wrap(err, "unable to get app config")
		}

		return nil
	})
	p.Go(func() error {
		var err error
		installComps, err = h.getInstallComponentsState(ctx, installID)
		if err != nil {
			return errors.Wrap(err, "unable to get install components")
		}
		return nil
	})
	p.Go(func() error {
		var err error
		stack, err = h.getInstallStack(ctx, installID)
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.Wrap(err, "unable to get install stack")
			}
		}
		return nil
	})
	p.Go(func() error {
		var err error
		sandboxRuns, err = h.getInstallSandboxRuns(ctx, installID)
		if err != nil {
			return errors.Wrap(err, "unable to get install sandbox runs")
		}
		return nil
	})

	p.Go(func() error {
		var err error
		actions, err = h.getInstallActionWorkflows(ctx, installID)
		if err != nil {
			return errors.Wrap(err, "unable to get actions")
		}

		return nil
	})
	p.Go(func() error {
		var err error
		install, err := h.getStateInstall(ctx, installID)
		if err != nil {
			return errors.Wrap(err, "unable to get install")
		}

		secrets, err = h.getSecrets(ctx, installID, install.RunnerID)
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.Wrap(err, "unable to get secrets")
			}
		}

		return nil
	})
	if err := p.Wait(); err != nil {
		return nil, errors.Wrap(err, "unable to get data for state")
	}

	// build the state with all the resources here
	is.ID = install.ID
	is.Name = install.Name
	is.Inputs = h.ToInputState(install.CurrentInstallInputs, appCfg, redacted)
	is.Cloud = h.ToCloudAccount(install)

	comps := h.ToComponents(installComps)
	is.Components = make(map[string]any, 0)
	for name, c := range comps.Components {
		cMap, err := state.AsMap(c)
		if err != nil {
			return nil, errors.Wrap(err, "unable to create map")
		}

		is.Components[name] = cMap
	}

	is.App = h.ToAppState(install.App)
	is.Org = h.ToOrgState(install.Org)

	if len(install.RunnerGroup.Runners) > 0 {
		is.Runner = h.ToRunnerState(install.RunnerGroup.Runners[0])
	}

	is.Sandbox = h.ToSandboxesState(sandboxRuns)
	if len(sandboxRuns) > 0 {
		is.Domain = h.ToDomainState(&sandboxRuns[0])
	} else {
		is.Domain = h.ToDomainState(nil)
	}

	is.Actions = h.ToActions(actions)
	is.InstallStack = h.ToInstallStackState(stack)
	is.Secrets = secrets
	is.Labels = map[string]string(install.Labels)
	// NOTE(JM): this is purely for historical and legacy reasons, and will be removed once we migrate all users to
	// the flattened structure
	is.Install = &state.InstallState{
		Populated: true,
		ID:        install.ID,
		Name:      install.Name,
		Sandbox:   *is.Sandbox,
	}
	if is.Domain != nil {
		is.Install.PublicDomain = is.Domain.PublicDomain
		is.Install.InternalDomain = is.Domain.InternalDomain
	}

	if is.Inputs != nil {
		is.Install.Inputs = is.Inputs.Inputs
	}

	return is, nil
}

func (h *Helpers) getStateInstall(ctx context.Context, installID string) (*app.Install, error) {
	var install app.Install
	res := h.db.WithContext(ctx).
		Preload("App").
		Preload("App.AppSecrets").
		Preload("Org").
		Preload("CreatedBy").
		Preload("AppRunnerConfig").
		Preload("RunnerGroup").
		Preload("RunnerGroup.Runners").
		Preload("AzureAccount").
		Preload("GCPAccount").
		Preload("AWSAccount").
		Preload("InstallInputs", func(db *gorm.DB) *gorm.DB {
			return db.Order(views.TableOrViewName(db, &app.InstallInputs{}, ".created_at DESC")).Limit(1)
		}).
		First(&install, "id = ?", installID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to find install: %w", res.Error)
	}

	return &install, nil
}

func (h *Helpers) ToInstallStackState(stack *app.InstallStack) *state.InstallStackState {
	return ToInstallStackState(stack)
}

func (h *Helpers) ToInputState(inputs *app.InstallInputs, cfg *app.AppConfig, redacted bool) *state.InputsState {
	return ToInputState(inputs, cfg, redacted)
}

func (h *Helpers) ToCloudAccount(install *app.Install) *state.CloudAccount {
	return ToCloudAccount(install)
}

func (h *Helpers) ToSandboxesState(sandboxRuns []app.InstallSandboxRun) *state.SandboxState {
	if len(sandboxRuns) < 1 {
		return state.NewSandboxState()
	}
	return ToSandboxState(&sandboxRuns[0])
}

func (h *Helpers) ToAppState(currentApp app.App) *state.AppState {
	return ToAppState(currentApp)
}

func (h *Helpers) ToOrgState(org app.Org) *state.OrgState {
	return ToOrgState(org)
}

func (h *Helpers) ToRunnerState(runner app.Runner) *state.RunnerState {
	return ToRunnerState(runner)
}

func (h *Helpers) ToDomainState(run *app.InstallSandboxRun) *state.DomainState {
	return ToDomainState(run)
}

func (h *Helpers) ToComponentState(installComp app.InstallComponent) *state.ComponentState {
	return ToComponentState(installComp)
}

func (h *Helpers) ToComponents(installComps []app.InstallComponent) *state.ComponentsState {
	st := state.NewComponentsState()
	st.Populated = true

	for _, instCmp := range installComps {
		st.Components[instCmp.Component.Name] = h.ToComponentState(instCmp)
	}
	return st
}

func (h *Helpers) ToActions(installActions []app.InstallActionWorkflow) *state.ActionsState {
	st := state.NewActionsState()
	st.Populated = true

	for _, instAct := range installActions {
		st.Workflows[instAct.ActionWorkflow.Name] = h.ToActionWorkflowState(instAct)
	}
	return st
}

func (h *Helpers) ToActionWorkflowState(act app.InstallActionWorkflow) *state.ActionWorkflowState {
	return ToActionWorkflowState(act)
}

// hydrateLabels fetches install labels and populates the Labels field on the
// state. Called on both the cached and fresh paths so labels are always current.
func (h *Helpers) hydrateLabels(ctx context.Context, installID string, is *state.State) error {
	var install app.Install
	if err := h.db.WithContext(ctx).Select("id", "labels").First(&install, "id = ?", installID).Error; err != nil {
		return errors.Wrap(err, "unable to get install labels")
	}

	is.Labels = map[string]string(install.Labels)
	return nil
}

func (h *Helpers) getLatestInstallStateRow(ctx context.Context, installID string) (*app.InstallState, error) {
	var is app.InstallState
	res := h.db.WithContext(ctx).
		Where(app.InstallState{InstallID: installID}).
		Order("created_at DESC").
		First(&is)
	if res.Error != nil {
		return nil, res.Error
	}

	return &is, nil
}

// regenerateStalePartials regenerates only the partials listed in row.StalePartials, merges them into
// the cached state, persists the result in place (clearing the stale markers), and returns it.
func (h *Helpers) regenerateStalePartials(ctx context.Context, row *app.InstallState, redacted, skipVersionCheck bool) (*state.State, error) {
	is := row.State
	if is == nil {
		is = state.New()
	}

	install, err := h.getStateInstall(ctx, row.InstallID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install")
	}
	is.ID = install.ID
	is.Name = install.Name

	for _, partial := range row.StalePartials {
		if err := h.regenerateStalePartial(ctx, install, partial, is, redacted, skipVersionCheck); err != nil {
			return nil, errors.Wrapf(err, "unable to regenerate partial %s", partial)
		}
	}

	MapLegacyFields(is)

	if res := h.db.WithContext(ctx).Model(row).
		Select("state", "stale_at", "stale_partials").
		Updates(app.InstallState{
			State:         is,
			StaleAt:       pkggenerics.NullTime{},
			StalePartials: []pkgstate.PartialName{},
		}); res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to persist regenerated state")
	}

	return is, nil
}

// regenerateStalePartial refreshes a single partial into state, reusing the same data fetch path as state gen signal.
func (h *Helpers) regenerateStalePartial(ctx context.Context, install *app.Install, partial pkgstate.PartialName, is *state.State, redacted, skipVersionCheck bool) error {
	switch partial {
	case pkgstate.PartialOrg:
		is.Org = h.ToOrgState(install.Org)
	case pkgstate.PartialApp:
		is.App = h.ToAppState(install.App)
	case pkgstate.PartialRunner:
		if len(install.RunnerGroup.Runners) > 0 {
			is.Runner = h.ToRunnerState(install.RunnerGroup.Runners[0])
		}
	case pkgstate.PartialCloud:
		is.Cloud = h.ToCloudAccount(install)
	case pkgstate.PartialInputs:
		appCfg, err := h.appsHelpers.GetFullAppConfig(ctx, install.AppConfigID, skipVersionCheck)
		if err != nil {
			return errors.Wrap(err, "unable to get app config")
		}
		is.Inputs = h.ToInputState(install.CurrentInstallInputs, appCfg, redacted)
	case pkgstate.PartialDomain:
		sandboxRuns, err := h.getInstallSandboxRuns(ctx, install.ID)
		if err != nil {
			return errors.Wrap(err, "unable to get install sandbox runs")
		}
		if len(sandboxRuns) > 0 {
			is.Domain = h.ToDomainState(&sandboxRuns[0])
		} else {
			is.Domain = h.ToDomainState(nil)
		}
	case pkgstate.PartialSandbox:
		sandboxRuns, err := h.getInstallSandboxRuns(ctx, install.ID)
		if err != nil {
			return errors.Wrap(err, "unable to get install sandbox runs")
		}
		is.Sandbox = h.ToSandboxesState(sandboxRuns)
	case pkgstate.PartialStack:
		stack, err := h.getInstallStack(ctx, install.ID)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.Wrap(err, "unable to get install stack")
		}
		is.InstallStack = h.ToInstallStackState(stack)
	case pkgstate.PartialSecrets:
		secrets, err := h.getSecrets(ctx, install.ID, install.RunnerID)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.Wrap(err, "unable to get secrets")
		}
		is.Secrets = secrets
	case pkgstate.PartialComponents:
		installComps, err := h.getInstallComponentsState(ctx, install.ID)
		if err != nil {
			return errors.Wrap(err, "unable to get install components")
		}
		comps := h.ToComponents(installComps)
		is.Components = make(map[string]any, len(comps.Components))
		for name, c := range comps.Components {
			cMap, err := state.AsMap(c)
			if err != nil {
				return errors.Wrap(err, "unable to create component map")
			}
			is.Components[name] = cMap
		}
	case pkgstate.PartialActions:
		actions, err := h.getInstallActionWorkflows(ctx, install.ID)
		if err != nil {
			return errors.Wrap(err, "unable to get actions")
		}
		is.Actions = h.ToActions(actions)
	default:
		return errors.Errorf("unknown partial: %s", partial)
	}

	return nil
}

func (h *Helpers) MapLegacyFields(is *state.State) {
	MapLegacyFields(is)
}

func ToInstallStackState(stack *app.InstallStack) *state.InstallStackState {
	if stack == nil || len(stack.InstallStackVersions) < 1 {
		return nil
	}
	is := state.NewInstallStackState()
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

func ToInputState(inputs *app.InstallInputs, cfg *app.AppConfig, redacted bool) *state.InputsState {
	if inputs == nil {
		return nil
	}
	inputValues := inputs.Values
	if redacted {
		inputValues = inputs.ValuesRedacted
	}
	if len(inputValues) < 1 {
		return nil
	}
	is := state.NewInputsState()
	for _, inp := range cfg.InputConfig.AppInputs {
		val, ok := inputValues[inp.Name]
		if !ok {
			val = &inp.Default
		}
		is.Inputs[inp.Name] = pkggenerics.FromPtrStr(val)
	}
	return is
}

func ToAppState(currentApp app.App) *state.AppState {
	st := state.NewAppState()
	st.Populated = true
	st.ID = currentApp.ID
	st.Name = currentApp.Name
	st.Status = string(currentApp.Status)
	for _, secr := range currentApp.AppSecrets {
		st.Variables[secr.Name] = secr.Value
	}
	return st
}

func ToOrgState(org app.Org) *state.OrgState {
	st := state.NewOrgState()
	st.Populated = true
	st.ID = org.ID
	st.Name = org.Name
	st.Status = string(org.Status)
	return st
}

func ToRunnerState(runner app.Runner) *state.RunnerState {
	st := state.NewRunnerState()
	st.Populated = true
	st.ID = runner.ID
	st.RunnerGroupID = runner.RunnerGroupID
	st.Status = string(runner.Status)
	return st
}

func ToDomainState(run *app.InstallSandboxRun) *state.DomainState {
	st := state.NewDomainState()
	if run == nil {
		return st
	}
	if v, ok := run.Outputs["public_domain"].(string); ok {
		st.PublicDomain = v
	}
	if v, ok := run.Outputs["internal_domain"].(string); ok {
		st.InternalDomain = v
	}
	return st
}

func ToComponentState(installComp app.InstallComponent) *state.ComponentState {
	st := state.NewComponentState()
	st.Populated = true
	st.Name = installComp.Component.Name
	st.ComponentID = installComp.ComponentID
	st.InstallComponentID = installComp.ID
	if len(installComp.InstallDeploys) > 0 {
		st.Status = string(installComp.InstallDeploys[0].Status)
		st.BuildID = string(installComp.InstallDeploys[0].ComponentBuildID)
		st.Outputs = installComp.InstallDeploys[0].Outputs
	}
	return st
}

func ToActionWorkflowState(act app.InstallActionWorkflow) *state.ActionWorkflowState {
	st := state.NewActionWorkflowState()
	st.Populated = true
	st.Status = string(act.Status)
	st.ID = act.ActionWorkflow.ID
	for _, run := range act.Runs {
		if run.RunnerJob != nil {
			st.Outputs = run.RunnerJob.ParsedOutputs
			break
		}
	}
	return st
}

func ToSecretsState(parsedOutputs map[string]interface{}) *state.SecretsState {
	empty := state.NewSecretsState()
	if len(parsedOutputs) == 0 {
		return &empty
	}
	var secretsState state.SecretsState
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
		return &empty
	}
	if err := decoder.Decode(parsedOutputs); err != nil {
		return &empty
	}
	if secretsState == nil {
		return &empty
	}
	return &secretsState
}

func ToSandboxState(run *app.InstallSandboxRun) *state.SandboxState {
	st := state.NewSandboxState()
	st.Populated = true
	st.Status = string(run.Status)
	st.Outputs = run.Outputs
	if run.AppSandboxConfig.PublicGitVCSConfig != nil {
		st.Type = run.AppSandboxConfig.PublicGitVCSConfig.Directory
		st.Version = run.AppSandboxConfig.PublicGitVCSConfig.Branch
	}
	if run.AppSandboxConfig.ConnectedGithubVCSConfig != nil {
		st.Type = run.AppSandboxConfig.ConnectedGithubVCSConfig.Directory
		st.Version = run.AppSandboxConfig.ConnectedGithubVCSConfig.Branch
	}
	return st
}

func ToCloudAccount(install *app.Install) *state.CloudAccount {
	st := state.NewCloudAccount()
	if install.AWSAccount != nil {
		st.AWS = &state.AWSCloudAccount{Region: install.AWSAccount.Region}
	}
	if install.AzureAccount != nil {
		st.Azure = &state.AzureCloudAccount{Location: install.AzureAccount.Location}
	}
	if install.GCPAccount != nil {
		st.GCP = &state.GCPCloudAccount{
			ProjectID: install.GCPAccount.ProjectID,
			Region:    install.GCPAccount.Region,
		}
	}
	return st
}

func MapLegacyFields(is *state.State) {
	is.Install = &state.InstallState{
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
