package helpers

import (
	"context"
	"encoding/json"

	"github.com/powertoolsdev/mono/pkg/types/state"
)

// GetInstallState reads the current state of the install from the DB, and returns it in a structure that can be used for variable interpolation.
func (h *Helpers) GetInstallState(ctx context.Context, installID string) (*state.InstallState, error) {
	installState := state.NewInstallState()

	// {{ .nuon.install.inputs }}
	err := h.populateInputs(ctx, installID, installState)
	if err != nil {
		return nil, err
	}

	// {{ .nuon.install.sandbox }}
	err = h.populateSandbox(ctx, installID, installState)
	if err != nil {
		return nil, err
	}

	// {{ .nuon.install }}
	// {{ .nuon.org.id }}
	// {{ .nuon.app.id }}
	err = h.populateInstall(ctx, installID, installState)
	if err != nil {
		return nil, err
	}

	// {{ .nuon.components }}
	err = h.populateComponents(ctx, installID, installState)
	if err != nil {
		return nil, err
	}

	// TODO(ja): add runner state?

	return installState, nil
}

func (h *Helpers) populateInputs(ctx context.Context, installID string, installState *state.InstallState) error {
	inputs, err := h.getInstallInputs(ctx, installID)
	if err != nil {
		return err
	}
	activeInputs := inputs[0]
	for key, val := range activeInputs.ValuesRedacted {
		value := ""
		if val != nil {
			value = *val
		}
		installState.Install.Inputs[key] = value
	}
	return nil
}

func (h *Helpers) populateSandbox(ctx context.Context, installID string, installState *state.InstallState) error {
	sandboxRuns, err := h.getInstallSandboxRuns(ctx, installID)
	if err != nil {
		return err
	}
	activeSandboxRun := sandboxRuns[0]
	if len(activeSandboxRun.RunnerJob.Outputs) > 0 {
		err = json.Unmarshal(activeSandboxRun.RunnerJob.Outputs, &installState.Install.Sandbox.Outputs)
		if err != nil {
			return err
		}
	}
	publicVCSConfig := activeSandboxRun.AppSandboxConfig.PublicGitVCSConfig
	connectedVCSConfig := activeSandboxRun.AppSandboxConfig.ConnectedGithubVCSConfig
	if publicVCSConfig != nil {
		installState.Install.Sandbox.Type = publicVCSConfig.Directory
		installState.Install.Sandbox.Version = publicVCSConfig.Branch
	} else {
		installState.Install.Sandbox.Type = connectedVCSConfig.Directory
		installState.Install.Sandbox.Version = connectedVCSConfig.Branch
	}
	return nil
}

func (h *Helpers) populateInstall(ctx context.Context, installID string, installState *state.InstallState) error {
	install, err := h.GetInstall(ctx, installID)
	if err != nil {
		return err
	}

	installState.Org.ID = install.OrgID

	installState.App.ID = install.AppID
	for _, secret := range install.App.AppSecrets {
		installState.App.Secrets[secret.Name] = secret.Value
	}

	installState.Install.ID = install.ID
	activeSandboxRun := install.InstallSandboxRuns[0]
	installState.Install.PublicDomain = ""
	publicDomain, ok := activeSandboxRun.RunnerJob.ParsedOutputs["public_domain"].(string)
	if ok {
		installState.Install.PublicDomain = publicDomain
	}
	installState.Install.InternalDomain = ""
	internalDomain, ok := activeSandboxRun.RunnerJob.ParsedOutputs["internal_domain"].(string)
	if ok {
		installState.Install.InternalDomain = internalDomain
	}
	return nil
}

func (h *Helpers) populateComponents(ctx context.Context, installID string, installState *state.InstallState) error {
	installComponents, err := h.getInstallComponents(ctx, installID)
	if err != nil {
		return err
	}
	for _, val := range installComponents {
		installState.Components[val.Component.Name] = state.NewComponentState()
		installDeploys := val.InstallDeploys
		if len(installDeploys) > 0 {
			runnerJobs := installDeploys[0].RunnerJobs
			if len(runnerJobs) > 0 && len(runnerJobs[0].Outputs) > 0 {
				err := json.Unmarshal(runnerJobs[0].Outputs, &installState.Components[val.Component.Name].Outputs)
				if err != nil {
					return err
				}
			}
		}

	}
	return nil
}
