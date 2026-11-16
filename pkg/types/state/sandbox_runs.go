package state

func NewSandboxState() *SandboxState {
	return &SandboxState{
		Outputs:    make(map[string]any, 0),
		RecentRuns: make([]*SandboxState, 0),
	}
}

type SandboxState struct {
	Populated bool `json:"populated"`

	BuildID string                 `json:"build_id,omitempty"`
	Status  string                 `json:"status"`
	Type    string                 `json:"type"`
	Version string                 `json:"version"`
	Outputs map[string]interface{} `json:"outputs" faker:"-"`

	RecentRuns []*SandboxState `json:"recent_runs"`
}
