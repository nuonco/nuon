package syncer

import (
	"github.com/nuonco/nuon/pkg/config/sync"
)

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

// GetActionStateIds returns the IDs of all actions in the current state.
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

// OrphanedComponents implements sync.Syncer
func (s *syncer) OrphanedComponents() map[string]string {
	orphaned := make(map[string]string)

	// Build map of current component names
	current := make(map[string]bool)
	for _, comp := range s.cfg.Components {
		current[comp.Name] = true
	}

	// Find components in previous state that are not in current config
	for _, prevComp := range s.prevState.Components {
		if !current[prevComp.Name] {
			orphaned[prevComp.Name] = prevComp.ID
		}
	}

	return orphaned
}

// OrphanedActions implements sync.Syncer
func (s *syncer) OrphanedActions() map[string]string {
	orphaned := make(map[string]string)

	// Build map of current action names
	current := make(map[string]bool)
	for _, action := range s.cfg.Actions {
		current[action.Name] = true
	}

	// Find actions in previous state that are not in current config
	for _, prevAction := range s.prevState.Actions {
		if !current[prevAction.Name] {
			orphaned[prevAction.Name] = prevAction.ID
		}
	}

	return orphaned
}
