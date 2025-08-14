package state

import (
	"go.temporal.io/sdk/workflow"
	"gorm.io/gorm"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/pkg/types/state"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
)

func (w *Workflows) getSandboxStatePartial(ctx workflow.Context, installID string) (*state.SandboxState, error) {
	sandboxRun, err := activities.AwaitGetInstallSandboxRunStateByInstallID(ctx, installID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &state.SandboxState{}, nil
		}

		return nil, errors.Wrap(err, "unable to get sandbox run")
	}

	st := w.toSandboxRunState(sandboxRun)
	return st, nil
}

func (h *Workflows) toSandboxRunState(run *app.InstallSandboxRun) *state.SandboxState {
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
