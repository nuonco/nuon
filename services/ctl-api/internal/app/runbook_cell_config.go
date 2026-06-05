package app

type RunbookCellConfig struct {
	Type    string `json:"type"`
	Content string `json:"content,omitempty"`
	StepIdx *int   `json:"step_idx,omitempty"`
	Name    string `json:"name,omitempty"`
}
