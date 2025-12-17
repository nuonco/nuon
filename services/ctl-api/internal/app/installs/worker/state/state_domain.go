package state

import (
	"strings"

	"go.temporal.io/sdk/workflow"
	"gorm.io/gorm"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/types/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
)

func (w *Workflows) getDomainPartial(ctx workflow.Context, installID string) (*state.DomainState, error) {
	sandboxRun, err := activities.AwaitGetInstallSandboxRunStateByInstallID(ctx, installID)
	if err != nil {
		if strings.Contains(err.Error(), gorm.ErrRecordNotFound.Error()) {
			return &state.DomainState{}, nil
		}

		return nil, errors.Wrap(err, "unable to get domain state")
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
