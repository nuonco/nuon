package state

func NewActionsState() *ActionsState {
	return &ActionsState{
		Workflows: make(map[string]*ActionWorkflowState, 0),
	}
}

type ActionsState struct {
	Populated bool `json:"populated"`

	Workflows map[string]*ActionWorkflowState `json:"workflows"`
}

func NewActionWorkflowState() *ActionWorkflowState {
	return &ActionWorkflowState{}
}

type ActionWorkflowState struct {
	Populated bool `json:"populated"`

	Status  string         `json:"status"`
	ID      string         `json:"id"`
	Outputs map[string]any `json:"outputs"`
}
