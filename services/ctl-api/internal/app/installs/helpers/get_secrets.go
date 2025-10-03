package helpers

import (
	"context"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	pkggenerics "github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/pkg/types/state"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func (h *Helpers) getSecrets(ctx context.Context, installID, runnerID string) (*state.SecretsState, error) {
	runnerJob, err := h.getSecretsSyncRunnerJob(ctx, installID, runnerID)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.Wrap(err, "unable to get secrets")
		}

		return &state.SecretsState{}, nil
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

	return &state, nil
}

func (h *Helpers) getSecretsSyncRunnerJob(ctx context.Context, installID, runnerID string) (*app.RunnerJob, error) {
	job := app.RunnerJob{}
	res := h.db.WithContext(ctx).
		Where(app.RunnerJob{
			Type:     app.RunnerJobTypeSandboxSyncSecrets,
			RunnerID: runnerID,
		}).
		Order("created_at desc").
		Limit(1).
		First(&job)

	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get runner job")
	}

	return &job, nil
}
