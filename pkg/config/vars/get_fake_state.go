package vars

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/pkg/config"
	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"github.com/powertoolsdev/mono/pkg/types/outputs"
	"github.com/powertoolsdev/mono/pkg/types/state"
)

func (v *varsValidator) GetFakeState(ctx context.Context) (state.State, error) {
	fmt.Println("GetFakeState")
	fakeState := state.State{}
	fakeState.ID = domains.NewInstallID()
	fakeState.Name = "fake-state"
	fakeState.App = v.getFakeAppState()
	fakeState.Org = v.getFakeOrgState()

	if v.cfg.Sandbox != nil {
		fakeState.Sandbox = v.getSandboxState()
	}

	if v.cfg.Inputs != nil {
		fakeState.Inputs = v.getInputState()
	}

	fakeState.Actions = v.getFakeActionsState()
	fakeState.Runner = v.getFakeRunnerState()

	fakeState.Components = make(map[string]any, 0)
	for _, comp := range v.cfg.Components {
		fmt.Println("comp", comp.Name)
		fakeComp, err := v.createFakeComponent(comp)
		if err != nil {
			return state.State{}, errors.Wrap(err, "unable to create fake component")
		}
		fakeState.Components[comp.Name] = fakeComp
	}

	fakeState.Domain = v.getFakeDomainState()
	fakeState.Cloud = v.getFakeCloudState()
	fakeState.InstallStack = v.getFakeInstallStackState()
	fakeState.Secrets = v.getFakeSecretsState()

	return fakeState, nil
}

func (v *varsValidator) getFakeAppState() *state.AppState {
	fakeApp := state.AppState{}
	fakeApp.Populated = true
	fakeApp.ID = domains.NewAppID()
	fakeApp.Name = "fake-app"
	fakeApp.Status = "active"
	// TODO: fake defaults
	fakeApp.Variables = make(map[string]string)
	return &fakeApp
}

func (v *varsValidator) getFakeOrgState() *state.OrgState {
	fakeOrg := state.OrgState{}
	fakeOrg.Populated = true
	fakeOrg.ID = domains.NewOrgID()
	fakeOrg.Name = "fake-org"
	return &fakeOrg
}

func (v *varsValidator) getSandboxState() *state.SandboxState {
	fakeSandbox := state.SandboxState{}
	fakeSandbox.Populated = true
	fakeSandbox.Status = "active"
	fakeSandbox.Type = "fake-type"
	// TODO: fake outputs
	fakeSandbox.Outputs = make(map[string]any)
	return &fakeSandbox
}

func (v *varsValidator) getInputState() *state.InputsState {
	fakeInputs := state.InputsState{}
	fakeInputs.Populated = true
	return &fakeInputs
}

func (v *varsValidator) getFakeActionState(actionCfg *config.ActionConfig) *state.ActionWorkflowState {
	fakeActions := state.ActionWorkflowState{}
	fakeActions.Populated = true
	fakeActions.Status = "active"
	fakeActions.ID = domains.NewActionID()
	// TODO: fake outputs
	fakeActions.Outputs = make(map[string]any)
	return &fakeActions
}

func (v *varsValidator) getFakeActionsState() *state.ActionsState {
	fakeActions := state.ActionsState{}
	fakeActions.Populated = true
	fakeActions.Workflows = make(map[string]*state.ActionWorkflowState)
	for _, action := range v.cfg.Actions {
		fakeAction := v.getFakeActionState(action)
		fakeActions.Workflows[action.Name] = fakeAction
	}
	return &fakeActions
}

func (v *varsValidator) getFakeRunnerState() *state.RunnerState {
	fakeRunner := state.RunnerState{}
	fakeRunner.Populated = true
	fakeRunner.Status = "active"
	fakeRunner.ID = domains.NewRunnerID()
	fakeRunner.RunnerGroupID = domains.NewRunnerGroupID()
	return &fakeRunner
}

func (v *varsValidator) createFakeComponent(compCfg *config.Component) (state.ComponentState, error) {
	fakeComp := state.ComponentState{}
	fakeComp.Populated = true
	fakeComp.Status = "active"
	fakeComp.ComponentID = domains.NewComponentID()
	fakeComp.BuildID = domains.NewBuildID()
	fakeComp.InstallComponentID = domains.NewInstallComponentID()
	// TODO: fake outputs
	fakeComp.Outputs = make(map[string]any)

	return fakeComp, nil
}

func (v *varsValidator) getFakeDomainState() *state.DomainState {
	fakeDomain := state.DomainState{}
	fakeDomain.Populated = true
	fakeDomain.PublicDomain = "fake-public-domain.com"
	fakeDomain.InternalDomain = "fake-internal-domain.com"
	return &fakeDomain
}

func (v *varsValidator) getFakeCloudState() *state.CloudAccount {
	isAWS := true // TODO: determine if AWS or Azure
	fakeCloud := state.CloudAccount{}
	if isAWS {
		fakeCloud.AWS = &state.AWSCloudAccount{
			Region: "us-east-1",
		}
	} else {
		fakeCloud.Azure = &state.AzureCloudAccount{
			Location: "eastus",
		}
	}

	return &fakeCloud
}

func (v *varsValidator) getFakeInstallStackState() *state.InstallStackState {
	fakeInstallStack := state.InstallStackState{}
	fakeInstallStack.Populated = true
	fakeInstallStack.Status = "active"
	fakeInstallStack.QuickLinkURL = "https://fake-quick-link-url.com"
	fakeInstallStack.TemplateURL = "https://fake-template-url.com"
	fakeInstallStack.Checksum = "fake-checksum"

	// TODO: fake the following fields
	fakeInstallStack.TemplateJSON = "{}"
	fakeInstallStack.Outputs = make(map[string]string)
	return &fakeInstallStack
}

func (v *varsValidator) getFakeSecretsState() *state.SecretsState {
	fakeSecrets := make(outputs.SyncSecretsOutput, 0)
	// TODO: fake secrets
	return &fakeSecrets
}
