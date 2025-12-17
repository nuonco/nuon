package state

import (
	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/types/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
)

func (w *Workflows) getRunnerStatePartial(ctx workflow.Context, installID string) (*state.RunnerState, error) {
	install, err := activities.AwaitGetByInstallID(ctx, installID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install")
	}

	runner, err := activities.AwaitGetRunnerByID(ctx, install.RunnerID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get runner")
	}

	st := state.NewRunnerState()
	st.Populated = true
	st.ID = runner.ID
	st.RunnerGroupID = runner.RunnerGroupID
	st.Status = string(runner.Status)

	return st, nil
}
