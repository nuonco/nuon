package state

import (
	"strings"
	"time"

	"go.temporal.io/sdk/workflow"
	"gorm.io/gorm"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"

	pkggenerics "github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/pkg/types/state"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
)

func (w *Workflows) getSecretsStatePartial(ctx workflow.Context, installID string) (*state.SecretsState, error) {
	runnerJob, err := activities.AwaitGetSecretsSyncJobByInstallID(ctx, installID)
	if err != nil {
		if strings.Contains(err.Error(), gorm.ErrRecordNotFound.Error()) {
			return &state.SecretsState{}, nil
		}

		return nil, errors.Wrap(err, "unable to get secrets state")
	}

	var state state.SecretsState
	decoderConfig := &mapstructure.DecoderConfig{
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToSliceHookFunc(","),
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToTimeHookFunc(time.RFC3339Nano),
			pkggenerics.StringToMapDecodeHook(),
		),
		WeaklyTypedInput: true,
		Result:           &state,
	}
	decoder, err := mapstructure.NewDecoder(decoderConfig)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create decoder")
	}
	if err := decoder.Decode(runnerJob.ParsedOutputs); err != nil {
		return nil, errors.Wrap(err, "unable to parse aws outputs")
	}
	return nil, nil
}
