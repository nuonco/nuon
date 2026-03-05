package apisyncer

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/sdks/nuon-go"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/sync"
)

// syncer implements sync.Syncer using API calls to ctl-api.
// This is the original implementation that communicates over HTTP.
type syncer struct {
	cfg *config.AppConfig

	apiClient   nuon.Client
	appID       string
	appConfigID string

	state     *sync.State
	prevState *sync.State

	cmpBuildsScheduled []string
	cliVersion         string
}

type syncStep struct {
	Resource string
	Method   func(context.Context) error
}

// Sync implements sync.Syncer
func (s *syncer) Sync(ctx context.Context) error {
	s.cmpBuildsScheduled = make([]string, 0)
	if s.cfg == nil {
		return sync.SyncInternalErr{
			Description: "nil config",
			Err:         fmt.Errorf("config is nil"),
		}
	}
	if err := s.fetchState(ctx); err != nil {
		return sync.SyncInternalErr{
			Description: "unable to fetch state",
			Err:         err,
		}
	}

	if err := s.start(ctx); err != nil {
		return sync.SyncInternalErr{
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
		return sync.SyncInternalErr{
			Description: "unable to update config status after syncing",
			Err:         err,
		}
	}

	return nil
}

// GetAppConfigID implements sync.Syncer
func (s *syncer) GetAppConfigID() string {
	return s.appConfigID
}

// GetComponentStateIds implements sync.Syncer
func (s *syncer) GetComponentStateIds() []string {
	ids := make([]string, 0)
	if s.state.Components == nil {
		return ids
	}

	for _, comp := range s.state.Components {
		ids = append(ids, comp.ID)
	}

	return ids
}

// GetComponentsScheduled implements sync.Syncer
func (s *syncer) GetComponentsScheduled() []sync.ComponentState {
	states := make([]sync.ComponentState, 0)
	if s.state.Components == nil {
		return states
	}
	for _, comp := range s.state.Components {
		for _, cmpID := range s.cmpBuildsScheduled {
			if cmpID == comp.ID {
				states = append(states, comp)
			}
		}
	}
	return states
}

// New creates a new API-based syncer that communicates with ctl-api over HTTP.
// This is the original sync implementation used by the CLI.
//
// Parameters:
//   - apiClient: nuon SDK client for making API calls
//   - appID: ID of the app to sync
//   - cliVersion: version of the CLI performing the sync (for tracking)
//   - cfg: parsed app configuration to sync
//
// Returns a sync.Syncer interface that can be used to perform the sync operation.
func New(apiClient nuon.Client, appID, cliVersion string, cfg *config.AppConfig) sync.Syncer {
	return &syncer{
		cfg:       cfg,
		apiClient: apiClient,
		appID:     appID,
		state: &sync.State{
			Version: sync.DefaultStateVersion,
			AppID:   appID,
		},
		prevState:  &sync.State{},
		cliVersion: cliVersion,
	}
}
