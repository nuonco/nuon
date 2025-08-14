package state

import (
	"strings"

	"go.temporal.io/sdk/workflow"
	"gorm.io/gorm"

	"github.com/pkg/errors"

	pkggenerics "github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/pkg/types/state"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
)

func (w *Workflows) getInputsStatePartial(ctx workflow.Context, installID string) (*state.InputsState, error) {
	inst, err := activities.AwaitGetByInstallID(ctx, installID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install")
	}

	inps, err := activities.AwaitGetInstallInputsStateByInstallID(ctx, installID)
	if err != nil {
		if strings.Contains(err.Error(), gorm.ErrRecordNotFound.Error()) {
			return &state.InputsState{}, nil
		}

		return nil, errors.Wrap(err, "unable to get domain state")
	}

	cfg, err := activities.AwaitGetAppConfigByID(ctx, inst.AppConfigID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get app config")
	}

	return w.toInputState(inps, cfg, false), nil
}

func (h *Workflows) toInputState(inputs *app.InstallInputs, cfg *app.AppConfig, redacted bool) *state.InputsState {
	inputValues := inputs.Values
	if redacted {
		inputValues = inputs.ValuesRedacted
	}
	if inputs == nil || len(inputValues) < 1 {
		return nil
	}

	is := state.NewInputsState()

	for _, inp := range cfg.InputConfig.AppInputs {
		val, ok := inputValues[inp.Name]
		if !ok {
			val = &inp.Default
		}

		is.Inputs[inp.Name] = pkggenerics.FromPtrStr(val)
	}

	return is
}
