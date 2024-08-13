package sync

import (
	"context"
	"fmt"

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

func (s *sync) Sync(ctx context.Context) (string, error) {
	if s.cfg == nil {
		return "", SyncInternalErr{
			Description: "nil config",
			Err:		 fmt.Errorf("config is nil"),
		}
	}
	if err := s.fetchState(ctx); err != nil {
		return "", SyncInternalErr{
			Description: "unable to fetch state",
			Err:         err,
		}
	}

	if err := s.start(ctx); err != nil {
		return "", SyncInternalErr{
			Description: "unable to start sync",
			Err:         err,
		}
	}

	steps, err := s.syncSteps()
	if err != nil {
		return "", err
	}

	// sync steps
	for _, step := range steps {
		if err := s.syncStep(ctx, step); err != nil {
			return "", err
		}
	}

	if err := s.finish(ctx); err != nil {
		return "", SyncInternalErr{
			Description: "unable to update config status after syncing",
			Err:         err,
		}
	}

	msg := s.notifyOrphanedComponents()

	return msg, nil
}

func (s *sync) GetComponentStateIds() []string {
	ids := make([]string, 0)
	if s.state.ComponentIDs == nil {
		return ids
	}

	for _, comp := range s.state.ComponentIDs {
		ids = append(ids, comp.ID)
	}

	return ids
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
