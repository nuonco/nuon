package state

import (
	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/types/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
)

func (w *Workflows) getSandboxStatePartial(ctx workflow.Context, installID string) (*state.SandboxState, error) {
	sandboxRun, err := activities.AwaitGetInstallSandboxRunStateByInstallID(ctx, installID)
	if err != nil {
		if generics.IsGormErrRecordNotFound(err) {
			return &state.SandboxState{}, nil
		}
		return nil, errors.Wrap(err, "unable to get sandbox run")
	}
	return helpers.ToSandboxState(sandboxRun), nil
}
