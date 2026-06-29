package blobbackfill

const (
	WorkflowName    = "BackfillBlobs"
	DayWorkflowName = "BackfillBlobsDay"
	WorkflowID      = "general-blob-backfill"

	ProgressQueryType = "progress"
)

// RangeRequest is the parent orchestrator input. Callers populate only Tables;
// the parent fills in the rest and carries it across continue-as-new.
type RangeRequest struct {
	Tables      []string         `json:"tables"`
	Initialized bool             `json:"initialized"`
	Pending     []DayBucket      `json:"pending"`
	Processed   map[string]int64 `json:"processed"`
	DaysTotal   int              `json:"days_total"`
	DaysDone    int              `json:"days_done"`
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
	Processed int64 `json:"processed"`
}

// Progress is the live snapshot returned by the parent's progress query.
type Progress struct {
	CurrentTable string           `json:"current_table"`
	CurrentDay   string           `json:"current_day"`
	DaysTotal    int              `json:"days_total"`
	DaysDone     int              `json:"days_done"`
	Processed    map[string]int64 `json:"processed"`
	Done         bool             `json:"done"`
}
