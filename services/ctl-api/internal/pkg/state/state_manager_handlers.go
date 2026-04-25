package state

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	stateactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/state/activities"
)

type handlerType string

const (
	handlerTypeQuery  handlerType = "query"
	handlerTypeUpdate handlerType = "update"
)

type handler struct {
	typ              handlerType
	handler          any
	handlerValidator any
}

func (sm *stateManager) registerHandlers(ctx workflow.Context) error {
	handlers := map[string]handler{
		ForceRegenerateUpdateName: {handlerTypeUpdate, sm.forceRegenerateHandler, nil},
		RegenerateUpdateName:      {handlerTypeUpdate, sm.regenerateHandler, nil},
		HintUpdateName:            {handlerTypeUpdate, sm.hintHandler, nil},
		FetchStateUpdateName:      {handlerTypeUpdate, sm.fetchStateHandler, nil},
		StatusQueryName:           {handlerTypeQuery, sm.statusHandler, nil},
		StopUpdateName:            {handlerTypeUpdate, sm.stopHandler, nil},
		RestartUpdateName:         {handlerTypeUpdate, sm.restartHandler, nil},
	}

	for name, h := range handlers {
		switch h.typ {
		case handlerTypeQuery:
			if err := workflow.SetQueryHandler(ctx, name, h.handler); err != nil {
				return errors.Wrapf(err, "unable to create query handler %s", name)
			}
		case handlerTypeUpdate:
			opts := workflow.UpdateHandlerOptions{
				Validator: h.handlerValidator,
			}
			if err := workflow.SetUpdateHandlerWithOptions(ctx, name, h.handler, opts); err != nil {
				return errors.Wrapf(err, "unable to create update handler %s", name)
			}
		}
	}

	return nil
}

// forceRegenerateHandler rebuilds all partials from scratch.
func (sm *stateManager) forceRegenerateHandler(ctx workflow.Context, req *ForceRegenerateRequest) (*ForceRegenerateResponse, error) {
	if err := workflow.Await(ctx, func() bool { return sm.ready }); err != nil {
		return nil, err
	}

	updated, err := sm.executeRegeneration(ctx, allPartialsSet())
	if err != nil {
		return nil, errors.Wrap(err, "force regeneration failed")
	}
	_ = updated

	return &ForceRegenerateResponse{
		State:       sm.state.CachedState,
		GeneratedAt: sm.state.LastGeneratedAt,
	}, nil
}

// regenerateHandler checks all partials for changes and updates stale ones.
func (sm *stateManager) regenerateHandler(ctx workflow.Context, req *RegenerateRequest) (*RegenerateResponse, error) {
	if err := workflow.Await(ctx, func() bool { return sm.ready }); err != nil {
		return nil, err
	}

	// Check each partial for modifications.
	stalePartials := make(map[PartialName]bool)
	for _, partial := range AllPartials {
		lastKnown := sm.state.LastModifiedAt[partial]
		resp, err := stateactivities.AwaitCheckModified(ctx, &stateactivities.CheckModifiedRequest{
			InstallID:   sm.installID,
			PartialName: string(partial),
			LastKnownAt: lastKnown,
		})
		if err != nil {
			return nil, errors.Wrapf(err, "unable to check modified for %s", partial)
		}
		if resp.Changed {
			stalePartials[partial] = true
		}
	}

	var updatedPartials []PartialName
	if len(stalePartials) > 0 {
		var err error
		updatedPartials, err = sm.executeRegeneration(ctx, stalePartials)
		if err != nil {
			return nil, errors.Wrap(err, "regeneration failed")
		}
	}

	return &RegenerateResponse{
		State:           sm.state.CachedState,
		UpdatedPartials: updatedPartials,
		GeneratedAt:     sm.state.LastGeneratedAt,
	}, nil
}

// hintHandler regenerates the partials affected by a specific change.
func (sm *stateManager) hintHandler(ctx workflow.Context, req *HintRequest) (*HintResponse, error) {
	if err := workflow.Await(ctx, func() bool { return sm.ready }); err != nil {
		return nil, err
	}

	partials, ok := HintToPartials[req.HintType]
	if !ok {
		return nil, fmt.Errorf("unknown hint type: %s", req.HintType)
	}

	partialsToRegen := make(map[PartialName]bool, len(partials))
	for _, p := range partials {
		partialsToRegen[p] = true
	}

	updatedPartials, err := sm.executeRegeneration(ctx, partialsToRegen)
	if err != nil {
		return nil, errors.Wrap(err, "hint regeneration failed")
	}

	return &HintResponse{
		State:           sm.state.CachedState,
		UpdatedPartials: updatedPartials,
		GeneratedAt:     sm.state.LastGeneratedAt,
	}, nil
}

// fetchStateHandler returns the current cached state without regenerating.
// If no cached state exists, triggers a full regeneration first.
func (sm *stateManager) fetchStateHandler(ctx workflow.Context, req *FetchStateRequest) (*FetchStateResponse, error) {
	if err := workflow.Await(ctx, func() bool { return sm.ready }); err != nil {
		return nil, err
	}

	if sm.state.CachedState == nil {
		if _, err := sm.executeRegeneration(ctx, allPartialsSet()); err != nil {
			return nil, errors.Wrap(err, "unable to generate state for fetch")
		}
	}

	return &FetchStateResponse{
		State:       sm.state.CachedState,
		GeneratedAt: sm.state.LastGeneratedAt,
	}, nil
}

// statusHandler returns workflow metadata.
func (sm *stateManager) statusHandler() (*StatusResponse, error) {
	return &StatusResponse{
		Ready:           sm.ready,
		LastGeneratedAt: sm.state.LastGeneratedAt,
		GenerationCount: sm.state.GenerationCount,
		LastModifiedAt:  sm.state.LastModifiedAt,
	}, nil
}

// stopHandler terminates the workflow.
func (sm *stateManager) stopHandler(ctx workflow.Context, req *StopRequest) (*StopResponse, error) {
	sm.stopped = true
	return &StopResponse{}, nil
}

// restartHandler triggers continue-as-new.
func (sm *stateManager) restartHandler(ctx workflow.Context, req *RestartRequest) (*RestartResponse, error) {
	sm.restarted = true
	return &RestartResponse{}, nil
}
