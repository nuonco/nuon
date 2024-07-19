package sync

import (
	"context"

	"github.com/nuonco/nuon-go"

	"github.com/powertoolsdev/mono/pkg/config"
)

type sync struct {
	cfg *config.AppConfig

	apiClient nuon.Client
	appID     string

	state     *state
	prevState *state
}

type syncStep struct {
	Resource string
	Method   func(context.Context) error
}

func (s *sync) Sync(ctx context.Context) error {
	if err := s.fetchState(ctx); err != nil {
		return SyncInternalErr{
			Description: "unable to fetch state",
			Err:         err,
		}
	}

	if err := s.start(ctx); err != nil {
		return SyncInternalErr{
			Description: "unable to start sync",
			Err:         err,
		}
	}

	steps, err := s.syncSteps()
	if err != nil {
		return err
	}

	// sync steps
	for _, step := range steps {
		if err := s.syncStep(ctx, step); err != nil {
			return err
		}
	}

	if err := s.finish(ctx); err != nil {
		return SyncInternalErr{
			Description: "unable to update config status after syncing",
			Err:         err,
		}
	}

	return nil
}

func New(apiClient nuon.Client, appID string, cfg *config.AppConfig) *sync {
	return &sync{
		cfg:       cfg,
		apiClient: apiClient,
		appID:     appID,
		state: &state{
			Version: defaultStateVersion,
			AppID:   appID,
		},
	}
}
