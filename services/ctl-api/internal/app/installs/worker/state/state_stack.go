package state

import (
	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/types/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
)

func (w *Workflows) getStackStatePartial(ctx workflow.Context, installID string) (*state.InstallStackState, error) {
	stack, err := activities.AwaitGetInstallStackStateByInstallID(ctx, installID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get stack")
	}
	return helpers.ToInstallStackState(stack), nil
}
