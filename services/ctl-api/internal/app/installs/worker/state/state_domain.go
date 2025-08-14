package state

import (
	"go.temporal.io/sdk/workflow"
	"gorm.io/gorm"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/pkg/types/state"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
)

func (w *Workflows) getDomainPartial(ctx workflow.Context, installID string) (*state.DomainState, error) {
	sandboxRun, err := activities.AwaitGetInstallSandboxRunStateByInstallID(ctx, installID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, errors.Wrap(err, "unable to get sandbox run")
	}

	st := w.toDomainState(sandboxRun)
	return st, nil
}

func (h *Workflows) toDomainState(run *app.InstallSandboxRun) *state.DomainState {
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
