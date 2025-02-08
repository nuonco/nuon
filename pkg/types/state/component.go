package state

func NewComponentsState() *ComponentsState {
	return &ComponentsState{
		Components: make(map[string]*ComponentState, 0),
	}
}

type ComponentsState struct {
	Populated bool `json:"populated"`

	Components map[string]*ComponentState `json:"components"`
}

func NewComponentState() *ComponentState {
	return &ComponentState{
		Outputs: make(map[string]interface{}),
	}
}

type componentImageRepository struct {
	ID          string `json:"id,omitempty"`
	ARN         string `json:"arn,omitempty"`
	Name        string `json:"name,omitempty"`
	URI         string `json:"uri,omitempty"`
	Image       string `json:"image,omitempty"`
	LoginServer string `json:"login_server,omitempty"`
}

type componentImageRegistry struct {
	ID string `json:"id"`
}

type componentImage struct {
	Tag        string                   `json:"tag"`
	Repository componentImageRepository `json:"repository"`
	Registry   componentImageRegistry   `json:"registry"`
}

type ComponentState struct {
	Populated          bool                   `json:"populated"`
	Status             string                 `json:"status"`
	BuildID            string                 `json:"build_id"`
	ComponentID        string                 `json:"component_id"`
	InstallComponentID string                 `json:"install_component_id"`
	Image              componentImage         `json:"image"`
	Outputs            map[string]interface{} `json:"outputs" faker:"componentOutputs"`
}
