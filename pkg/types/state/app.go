package state

func NewAppState() *AppState {
	return &AppState{
		Secrets: make(map[string]string, 0),
	}
}

type AppState struct {
	Populated bool `json:"populated"`

	Status  string            `json:"status"`
	ID      string            `json:"id"`
	Name    string            `json:"name"`
	Secrets map[string]string `json:"secrets"`
}
