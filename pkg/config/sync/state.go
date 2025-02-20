package sync

import (
	"context"
	"encoding/json"
	"slices"

	"github.com/nuonco/nuon-go"
	"github.com/nuonco/nuon-go/models"
)

const (
	defaultStateVersion string = "v1"
)

type componentState struct {
	Name     string                  `json:"name"`
	ID       string                  `json:"id"`
	ConfigID string                  `json:"config_id"`
	Type     models.AppComponentType `json:"type"`
	Checksum string                  `json:"checksum"`
}

type actionState struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

type state struct {
	Version string `json:"version"`

	CfgID           string           `json:"config_id"`
	AppID           string           `json:"app_id"`
	InstallerID     string           `json:"installer_id"`
	RunnerConfigID  string           `json:"runner_config_id"`
	SandboxConfigID string           `json:"sandbox_config_id"`
	InputConfigID   string           `json:"input_config_id"`
	Components      []componentState `json:"components"`
	Actions         []actionState    `json:"actions"`
}

func (s *sync) getComponentStateById(id string) *componentState {
	for _, comp := range s.prevState.Components {
		if comp.ID == id {
			return &comp
		}
	}

	return nil
}

func (s *sync) OrphanedActions() map[string]string {
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

func (s *sync) OrphanedComponents() map[string]string {
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

func (s *sync) fetchState(ctx context.Context) error {
	cfg, err := s.apiClient.GetAppLatestConfig(ctx, s.appID)
	if err != nil {
		if nuon.IsNotFound(err) {
			s.prevState = &state{}
			return nil
		}

		return err
	}

	var prevState state

	// NOTE(jm): this is required to handle in-flight configs, that do not have a previous state.
	if cfg.State != "" {
		if err := json.Unmarshal([]byte(cfg.State), &prevState); err != nil {
			return err
		}
	}

	s.prevState = &prevState
	return nil
}

// add previous component state to current state if it is present in the cfg and prevoius state
// but is not present in the current state.
// required to ensure state persists between configs if we have a partial sync.
func (s *sync) reconcileStates() {
	// reconcile components
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

func (s *sync) getPrevComponentStateByName(name string) *componentState {
	for _, comp := range s.prevState.Components {
		if comp.Name == name {
			return &comp
		}
	}

	return nil
}

func (s *sync) getComponentStateByName(name string) *componentState {
	for _, comp := range s.state.Components {
		if comp.Name == name {
			return &comp
		}
	}

	return nil
}
