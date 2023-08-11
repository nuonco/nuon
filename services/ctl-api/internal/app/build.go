package app

type Build struct {
	Model

	ComponentID string
	Component   Component
	CreatedByID string
	GitRef      string `json:"git_ref"`
}
