package state

type appState struct {
	ID      string            `json:"id"`
	Secrets map[string]string `json:"secrets"`
}
