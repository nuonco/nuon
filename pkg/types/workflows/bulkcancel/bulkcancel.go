package bulkcancel

const (
	WorkflowName = "BulkCancelWorkflows"

	ProgressQueryType = "progress"
)

type Request struct {
	Pending []string `json:"pending"`
	Reason  string   `json:"reason"`

	Total     int      `json:"total"`
	Cancelled int      `json:"cancelled"`
	Skipped   int      `json:"skipped"`
	Failed    int      `json:"failed"`
	FailedIDs []string `json:"failed_ids"`
}

type Progress struct {
	Total             int      `json:"total"`
	Done              int      `json:"done"`
	Cancelled         int      `json:"cancelled"`
	Skipped           int      `json:"skipped"`
	Failed            int      `json:"failed"`
	FailedIDs         []string `json:"failed_ids,omitempty"`
	CurrentWorkflowID string   `json:"current_workflow_id,omitempty"`
	Finished          bool     `json:"finished"`
}
