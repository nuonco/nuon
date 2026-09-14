package releases

type RunbookTemplate struct {
	ID    string        `json:"id"`
	Name  string        `json:"name"`
	Steps []RunbookStep `json:"steps"`
}

type RunbookStep struct {
	Kind      string `json:"kind"`
	RefID     string `json:"ref_id,omitempty"`
	Component string `json:"component,omitempty"`
}
