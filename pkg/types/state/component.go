package state

func NewComponentsState() *ComponentsState {
	return &ComponentsState{
		Components: make(map[string]*ComponentState, 0),
	}
}

type ComponentsStateLegacy map[string]any

type ComponentsState struct {
	Populated bool `json:"populated"`

	Components map[string]*ComponentState `json:"components"`
}

func NewComponentState() *ComponentState {
	return &ComponentState{
		Outputs: make(map[string]interface{}),
	}
}

type ComponentState struct {
	Populated          bool                   `json:"populated"`
	Status             string                 `json:"status"`
	BuildID            string                 `json:"build_id"`
	ComponentID        string                 `json:"component_id"`
	InstallComponentID string                 `json:"install_component_id"`
	Outputs            map[string]interface{} `json:"outputs" faker:"componentOutputs"`
}
