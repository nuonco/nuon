package state

func NewAppState() *AppState {
	return &AppState{
		Variables: make(map[string]string, 0),
	}
}

type AppState struct {
	Populated bool `json:"populated"`

	Status    string            `json:"status"`
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Variables map[string]string `json:"variables"`
}
