package helpers

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	pkggenerics "github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/pkg/types/state"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/generics"
)

// GetInstallState reads the current state of the install from the DB, and returns it in a structure that can be used for variable interpolation.
func (h *Helpers) GetInstallState(ctx context.Context, installID string) (*state.State, error) {
	is := state.New()

	// collect all data up front
	install, err := h.getStateInstall(ctx, installID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install")
	}
	is.ID = install.ID
	is.Name = install.Name
	is.Inputs = h.toInputState(install.CurrentInstallInputs)
	is.Cloud = h.toCloudAccount(install)

	installComps, err := h.GetInstallComponents(ctx, installID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install components")
	}
	is.Components = h.toComponents(installComps)
	is.App = h.toAppState(install.App)
	is.Org = h.toOrgState(install.Org)

	if len(install.RunnerGroup.Runners) > 0 {
		is.Runner = h.toRunnerState(install.RunnerGroup.Runners[0])
	}

	sandboxRuns, err := h.getInstallSandboxRuns(ctx, installID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install sandbox runs")
	}
	is.Sandbox = h.toSandboxesState(sandboxRuns)
	if len(sandboxRuns) > 0 {
		is.Domain = h.toDomainState(&sandboxRuns[0])
	}

	actions, err := h.getInstallActionWorkflows(ctx, installID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get actions")
	}
	is.Actions = h.toActions(actions)

	stack, err := h.getInstallStack(ctx, installID)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.Wrap(err, "unable to get install stack")
		}
	}
	is.InstallStack = h.toInstallStackState(stack)

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
		Preload("Org").
		Preload("CreatedBy").
		Preload("AppRunnerConfig").
		Preload("RunnerGroup").
		Preload("RunnerGroup.Runners").
		Preload("AzureAccount").
		Preload("AWSAccount").
		Preload("InstallInputs", func(db *gorm.DB) *gorm.DB {
			return db.Order("install_inputs_view_v1.created_at DESC").Limit(1)
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

func (h *Helpers) toInputState(inputs *app.InstallInputs) *state.InputsState {
	if inputs == nil || len(inputs.Values) < 1 {
		return nil
	}

	is := state.NewInputsState()
	for key, val := range inputs.ValuesRedacted {
		is.Inputs[key] = pkggenerics.FromPtrStr(val)
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
	st.Outputs = run.RunnerJob.ParsedOutputs

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
		st.Secrets[secr.Name] = secr.Value
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

	publicDomain, ok := run.RunnerJob.ParsedOutputs["public_domain"].(string)
	if ok {
		st.PublicDomain = publicDomain
	}

	internalDomain, ok := run.RunnerJob.ParsedOutputs["internal_domain"].(string)
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
