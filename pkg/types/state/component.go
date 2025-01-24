package state

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
	Image   componentImage         `json:"image"`
	Outputs map[string]interface{} `json:"outputs" faker:"componentOutputs"`
}
