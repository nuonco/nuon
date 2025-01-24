package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/pkg/types/state"
)

// @ID GetInstallState
// @Summary	Get the current state of an install.
// @Description.markdown	get_install_state.md
// @Param			install_id	path	string	true	"install ID"
// @Tags			installs
// @Accept			json
// @Produce		json
// @Security APIKey
// @Security OrgID
// @Failure		400				{object}	stderr.ErrResponse
// @Failure		401				{object}	stderr.ErrResponse
// @Failure		403				{object}	stderr.ErrResponse
// @Failure		404				{object}	stderr.ErrResponse
// @Failure		500				{object}	stderr.ErrResponse
// @Success		200				{object} state.InstallState
// @Router			/v1/installs/{install_id}/state [get]
func (s *service) GetInstallState(ctx *gin.Context) {
	installID := ctx.Param("install_id")
	state, err := s.getInstallState(ctx, installID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, state)
}

// getInstallState reads the current state of the install, and returns it in a structure that supports variables interpolation,
// as defined here: https://docs.nuon.co/guides/using-variables#variable-data-sources
func (s *service) getInstallState(ctx context.Context, installID string) (*state.InstallState, error) {
	installState := state.NewInstallState()

	// {{ .nuon.install.inputs }}
	err := s.populateInputs(ctx, installID, installState)
	if err != nil {
		return nil, err
	}

	// {{ .nuon.install.sandbox }}
	err = s.populateSandbox(ctx, installID, installState)
	if err != nil {
		return nil, err
	}

	// {{ .nuon.install }}
	// {{ .nuon.org.id }}
	// {{ .nuon.app.id }}
	err = s.populateInstall(ctx, installID, installState)
	if err != nil {
		return nil, err
	}

	// {{ .nuon.components }}
	err = s.populateComponents(ctx, installID, installState)
	if err != nil {
		return nil, err
	}

	// TODO(ja): add runner state?

	return installState, nil
}

func (s *service) populateInputs(ctx context.Context, installID string, installState *state.InstallState) error {
	inputs, err := s.getInstallInputs(ctx, installID)
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

func (s *service) populateSandbox(ctx context.Context, installID string, installState *state.InstallState) error {
	sandboxRuns, err := s.getInstallSandboxRuns(ctx, installID)
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

func (s *service) populateInstall(ctx context.Context, installID string, installState *state.InstallState) error {
	install, err := s.getInstall(ctx, installID)
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

func (s *service) populateComponents(ctx context.Context, installID string, installState *state.InstallState) error {
	installComponents, err := s.getInstallComponents(ctx, installID)
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
