package state

import (
	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/types/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
)

func (w *Workflows) getInputsStatePartial(ctx workflow.Context, installID string) (*state.InputsState, error) {
	inst, err := activities.AwaitGetByInstallID(ctx, installID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install")
	}
	inps, err := activities.AwaitGetInstallInputsStateByInstallID(ctx, installID)
	if err != nil {
		if generics.IsGormErrRecordNotFound(err) {
			return &state.InputsState{}, nil
		}
		return nil, errors.Wrap(err, "unable to get inputs state")
	}
	cfg, err := activities.AwaitGetAppConfigByID(ctx, inst.AppConfigID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get app config")
	}
	return helpers.ToInputState(inps, cfg, false), nil
}
