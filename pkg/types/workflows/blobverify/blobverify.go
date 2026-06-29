package blobverify

const (
	WorkflowName    = "VerifyBlobs"
	DayWorkflowName = "VerifyBlobsDay"
	WorkflowID      = "general-blob-verify"

	ProgressQueryType = "progress"
)

// RangeRequest is the parent orchestrator input. Callers populate only Tables;
// the parent fills in the rest and carries it across continue-as-new.
type RangeRequest struct {
	Tables      []string                 `json:"tables"`
	Initialized bool                     `json:"initialized"`
	Pending     []DayBucket              `json:"pending"`
	Tallies     map[string]TableProgress `json:"tallies"`
	DaysTotal   int                      `json:"days_total"`
	DaysDone    int                      `json:"days_done"`
}

// DayBucket is one unit of work: one table on one UTC calendar day.
type DayBucket struct {
	Table string `json:"table"`
	Day   string `json:"day"`
}

type DayRequest struct {
	Table string `json:"table"`
	Day   string `json:"day"`
}

type DayResult struct {
	Checked       int64    `json:"checked"`
	Mismatched    int64    `json:"mismatched"`
	NotSet        int64    `json:"not_set"`
	MismatchedIDs []string `json:"mismatched_ids,omitempty"`
}

type TableProgress struct {
	Checked    int64 `json:"checked"`
	Mismatched int64 `json:"mismatched"`
	NotSet     int64 `json:"not_set"`
	// MismatchedIDs samples mismatched row ids, capped to keep history bounded.
	MismatchedIDs []string `json:"mismatched_ids,omitempty"`
}

// Progress is the live snapshot returned by the parent's progress query.
type Progress struct {
	CurrentTable string                   `json:"current_table"`
	CurrentDay   string                   `json:"current_day"`
	DaysTotal    int                      `json:"days_total"`
	DaysDone     int                      `json:"days_done"`
	Tables       map[string]TableProgress `json:"tables"`
	Done         bool                     `json:"done"`
}
