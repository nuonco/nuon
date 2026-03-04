package apisyncer

import (
	"context"
	"encoding/json"
	"slices"

	"github.com/nuonco/nuon/sdks/nuon-go"

	"github.com/nuonco/nuon/pkg/config/sync"
)

func (s *syncer) getComponentStateById(id string) *sync.ComponentState {
	for _, comp := range s.prevState.Components {
		if comp.ID == id {
			return &comp
		}
	}

	return nil
}

// OrphanedActions implements sync.Syncer
func (s *syncer) OrphanedActions() map[string]string {
	actions := map[string]string{}
	currentStateNames := make([]string, 0)
	for _, actionState := range s.state.Actions {
		currentStateNames = append(currentStateNames, actionState.Name)
	}
	for _, prevActionState := range s.prevState.Actions {
		if !slices.Contains(currentStateNames, prevActionState.Name) {
			actions[prevActionState.Name] = prevActionState.ID
		}
	}
	return actions
}

// OrphanedComponents implements sync.Syncer
func (s *syncer) OrphanedComponents() map[string]string {
	components := map[string]string{}
	currentStateIDs := make([]string, 0)
	for _, compState := range s.state.Components {
		currentStateIDs = append(currentStateIDs, compState.ID)
	}
	for _, prevCompState := range s.prevState.Components {
		if !slices.Contains(currentStateIDs, prevCompState.ID) {
			components[prevCompState.Name] = prevCompState.ID
		}
	}
	return components
}

func (s *syncer) fetchState(ctx context.Context) error {
	cfg, err := s.apiClient.GetAppLatestConfig(ctx, s.appID)
	if err != nil {
		if nuon.IsNotFound(err) {
			s.prevState = &sync.State{}
			return nil
		}

		return err
	}

	var prevState sync.State

	// NOTE(jm): this is required to handle in-flight configs, that do not have a previous state.
	if cfg.State != "" {
		if err := json.Unmarshal([]byte(cfg.State), &prevState); err != nil {
			return err
		}
	}

	s.prevState = &prevState
	return nil
}

// add previous component state to current state if it is present in the cfg and previous state
// but is not present in the current state.
// required to ensure state persists between configs if we have a partial sync.
func (s *syncer) reconcileStates() {
	// reconcile components
	if s.state.Components == nil {
		s.state.Components = make([]sync.ComponentState, 0)
	}
	for _, comp := range s.cfg.Components {
		obj := comp

		resourceName := obj.Name
		prevCompState := s.getPrevComponentStateByName(resourceName)
		if prevCompState != nil && s.getComponentStateByName(resourceName) == nil {
			// if component is not in the current state, add it
			s.state.Components = append(s.state.Components, *prevCompState)
		}
	}
}

func (s *syncer) getPrevComponentStateByName(name string) *sync.ComponentState {
	for _, comp := range s.prevState.Components {
		if comp.Name == name {
			return &comp
		}
	}

	return nil
}

func (s *syncer) getComponentStateByName(name string) *sync.ComponentState {
	for _, comp := range s.state.Components {
		if comp.Name == name {
			return &comp
		}
	}

	return nil
}
