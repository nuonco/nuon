package helpers

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/sourcegraph/conc/pool"
	"gorm.io/gorm"

	pkggenerics "github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/pkg/types/outputs"
	"github.com/powertoolsdev/mono/pkg/types/state"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/views"
)

// GetInstallState reads the current state of the install from the DB, and returns it in a structure that can be used for variable interpolation.
func (h *Helpers) GetInstallState(ctx context.Context, installID string, redacted bool) (*state.State, error) {
	es, err := h.getInstallStateFromDB(ctx, installID)
	if err == nil {
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

		appCfg, err = h.appsHelpers.GetFullAppConfig(ctx, install.AppConfigID)
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
	is.Inputs = h.toInputState(install.CurrentInstallInputs, appCfg, redacted)
	is.Cloud = h.toCloudAccount(install)

	comps := h.toComponents(installComps)
	is.Components = make(map[string]any, 0)
	for name, c := range comps.Components {
		cMap, err := state.AsMap(c)
		if err != nil {
			return nil, errors.Wrap(err, "unable to create map")
		}

		is.Components[name] = cMap
	}
	is.Components["comps"] = comps

	is.App = h.toAppState(install.App)
	is.Org = h.toOrgState(install.Org)

	if len(install.RunnerGroup.Runners) > 0 {
		is.Runner = h.toRunnerState(install.RunnerGroup.Runners[0])
	}

	is.Sandbox = h.toSandboxesState(sandboxRuns)
	if len(sandboxRuns) > 0 {
		is.Domain = h.toDomainState(&sandboxRuns[0])
	}

	is.Actions = h.toActions(actions)
	is.InstallStack = h.toInstallStackState(stack)
	is.Secrets = secrets
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

func (h *Helpers) toInstallStackState(stack *app.InstallStack) *state.InstallStackState {
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

	is.Outputs = generics.ToStringMap(stack.InstallStackOutputs.Data)

	return is
}

func (h *Helpers) toInputState(inputs *app.InstallInputs, cfg *app.AppConfig, redacted bool) *state.InputsState {
	inputValues := inputs.Values
	if redacted {
		inputValues = inputs.ValuesRedacted
	}
	if inputs == nil || len(inputValues) < 1 {
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

func (h *Helpers) toCloudAccount(install *app.Install) *state.CloudAccount {
	st := state.NewCloudAccount()

	if install.AWSAccount != nil {
		st.AWS = &state.AWSCloudAccount{
			Region: install.AWSAccount.Region,
		}
	}

	if install.AzureAccount != nil {
		st.Azure = &state.AzureCloudAccount{
			Location: install.AzureAccount.Location,
		}
	}

	return st
}

func (h *Helpers) toSandboxesState(sandboxRuns []app.InstallSandboxRun) *state.SandboxState {
	if len(sandboxRuns) < 1 {
		return state.NewSandboxState()
	}

	st := h.toSandboxRunState(sandboxRuns[0])
	for _, run := range sandboxRuns[1:] {
		runSt := h.toSandboxRunState(run)
		st.RecentRuns = append(st.RecentRuns, runSt)
	}

	return st
}

func (h *Helpers) toSandboxRunState(run app.InstallSandboxRun) *state.SandboxState {
	st := state.NewSandboxState()

	st.Populated = true
	st.Status = string(run.Status)
	st.Outputs = run.Outputs

	publicVCSConfig := run.AppSandboxConfig.PublicGitVCSConfig
	connectedVCSConfig := run.AppSandboxConfig.ConnectedGithubVCSConfig
	if publicVCSConfig != nil {
		st.Type = publicVCSConfig.Directory
		st.Version = publicVCSConfig.Branch
	}
	if connectedVCSConfig != nil {
		st.Type = connectedVCSConfig.Directory
		st.Version = connectedVCSConfig.Branch
	}

	return st
}

func (h *Helpers) toAppState(currentApp app.App) *state.AppState {
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

func (h *Helpers) toOrgState(org app.Org) *state.OrgState {
	st := state.NewOrgState()
	st.Populated = true
	st.ID = org.ID
	st.Name = org.Name
	st.Status = string(org.Status)

	return st
}

func (h *Helpers) toRunnerState(runner app.Runner) *state.RunnerState {
	st := state.NewRunnerState()
	st.Populated = true
	st.ID = runner.ID
	st.RunnerGroupID = runner.RunnerGroupID
	st.Status = string(runner.Status)

	return st
}

func (h *Helpers) toDomainState(run *app.InstallSandboxRun) *state.DomainState {
	st := state.NewDomainState()
	if run == nil {
		return st
	}

	publicDomain, ok := run.Outputs["public_domain"].(string)
	if ok {
		st.PublicDomain = publicDomain
	}

	internalDomain, ok := run.Outputs["internal_domain"].(string)
	if ok {
		st.InternalDomain = internalDomain
	}

	return st
}

func (h *Helpers) toComponent(installComp app.InstallComponent) *state.ComponentState {
	st := state.NewComponentState()

	st.Populated = true
	st.ComponentID = installComp.ComponentID
	st.InstallComponentID = installComp.ID

	installDeploys := installComp.InstallDeploys
	if len(installDeploys) < 1 {
		return st
	}
	st.Status = string(installDeploys[0].Status)
	st.BuildID = string(installDeploys[0].ComponentBuildID)
	st.Outputs = installDeploys[0].Outputs

	return st
}

func (h *Helpers) toComponents(installComps []app.InstallComponent) *state.ComponentsState {
	st := state.NewComponentsState()
	st.Populated = true

	for _, instCmp := range installComps {
		st.Components[instCmp.Component.Name] = h.toComponent(instCmp)
	}
	return st
}

func (h *Helpers) toActions(installActions []app.InstallActionWorkflow) *state.ActionsState {
	st := state.NewActionsState()
	st.Populated = true

	for _, instAct := range installActions {
		st.Workflows[instAct.ActionWorkflow.Name] = h.toActionWorkflow(instAct)
	}
	return st
}

func (h *Helpers) toActionWorkflow(act app.InstallActionWorkflow) *state.ActionWorkflowState {
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

func (h *Helpers) getInstallStateFromDB(ctx context.Context, installID string) (*state.State, error) {
	var is app.InstallState
	res := h.db.WithContext(ctx).
		Where("install_id = ?", installID).
		Order("created_at DESC").
		First(&is)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to find install state")
	}

	if !is.StaleAt.Empty() {
		is.State.StaleAt = &is.StaleAt.Time
	}

	return is.State, nil
}
