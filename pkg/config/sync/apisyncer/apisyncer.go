package apisyncer

import (
	"context"
	"fmt"
	stdsync "sync"

	"golang.org/x/sync/errgroup"

	"github.com/nuonco/nuon/sdks/nuon-go"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/sync"
)

// syncConcurrency caps the number of concurrent syncStep goroutines per
// phase. ctl-api comfortably handles much more, but bounding here keeps a
// runaway client retry from blasting the API and keeps interleaved progress
// output legible.
const syncConcurrency = 10

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

	// mu guards concurrent writes to state.Components / state.Actions /
	// state.Runbooks and cmpBuildsScheduled, which are appended to by
	// per-component / per-action / per-runbook syncStep goroutines.
	mu stdsync.Mutex
}

type syncStep struct {
	Resource string
	Method   func(context.Context) error
}

// trackBuildScheduled records that a component build was scheduled during
// this sync. Safe to call from concurrent syncStep goroutines.
func (s *syncer) trackBuildScheduled(compID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cmpBuildsScheduled = append(s.cmpBuildsScheduled, compID)
}

// appendComponentState records a component's resolved state. Safe to call
// from concurrent syncStep goroutines.
func (s *syncer) appendComponentState(state sync.ComponentState) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Components = append(s.state.Components, state)
}

// appendActionState records an action's resolved state. Safe to call from
// concurrent syncStep goroutines.
func (s *syncer) appendActionState(state sync.ActionState) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Actions = append(s.state.Actions, state)
}

// appendRunbookState records a runbook's resolved state. Safe to call from
// concurrent syncStep goroutines.
func (s *syncer) appendRunbookState(state sync.RunbookState) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Runbooks = append(s.state.Runbooks, state)
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

	phases, err := s.syncPhases()
	if err != nil {
		return err
	}

	// Phases run sequentially relative to each other (e.g. component-ensure
	// must complete before component-sync, and both must complete before
	// actions / runbooks which reference component IDs). Within a phase the
	// steps are independent and run concurrently, capped by syncConcurrency.
	for _, phase := range phases {
		g, gctx := errgroup.WithContext(ctx)
		g.SetLimit(syncConcurrency)
		for _, step := range phase {
			step := step
			g.Go(func() error {
				return s.syncStep(gctx, step)
			})
		}
		if err := g.Wait(); err != nil {
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

// GetActionStateIds implements sync.Syncer
func (s *syncer) GetActionStateIds() []string {
	ids := make([]string, 0)
	if s.state == nil || s.state.Actions == nil {
		return ids
	}
	for _, action := range s.state.Actions {
		ids = append(ids, action.ID)
	}
	return ids
}

// GetRunbookStateIds implements sync.Syncer
func (s *syncer) GetRunbookStateIds() []string {
	ids := make([]string, 0)
	if s.state == nil || s.state.Runbooks == nil {
		return ids
	}
	for _, runbook := range s.state.Runbooks {
		ids = append(ids, runbook.ID)
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
