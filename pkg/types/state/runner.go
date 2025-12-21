package state

func NewRunnerState() *RunnerState {
	return &RunnerState{}
}

type RunnerState struct {
	Populated bool `json:"populated"`

	ID            string `json:"id"`
	RunnerGroupID string `json:"runner_group_id"`
	Status        string `json:"status"`
}
