// Package runnerhealthcheckbackfill holds the shared constants and payload
// types for the runner-healthcheck-emitter backfill workflow, so the ctl-api
// service (which starts and queries the workflow) and the runners worker (which
// implements it) can share them without an import cycle.
package runnerhealthcheckbackfill

const (
	// WorkflowName must match the registered workflow function name.
	WorkflowName = "BackfillRunnerHealthcheckEmitters"
	// WorkflowID is fixed so a re-triggered backfill reuses the running one
	// rather than starting a duplicate.
	WorkflowID = "runners-healthcheck-emitter-backfill"

	ProgressQueryType = "progress"

	// DefaultBatchSize is how many runners each activity call processes. Kept
	// small so each batch holds little in memory and starts few emitter
	// workflows at once.
	DefaultBatchSize = 10
)

// Request is the orchestrator input. Callers leave it zero; the workflow fills
// in the cursor and running totals and carries them across continue-as-new.
type Request struct {
	// Cursor is the last runner ID processed (keyset pagination). Empty starts
	// from the beginning.
	Cursor string `json:"cursor"`
	// BatchSize overrides DefaultBatchSize when non-zero.
	BatchSize int `json:"batch_size"`

	RunnersProcessed int `json:"runners_processed"`
	EmittersCreated  int `json:"emitters_created"`
	AlreadyPresent   int `json:"already_present"`
	Errors           int `json:"errors"`
}

// Progress is the live snapshot returned by the workflow's progress query.
type Progress struct {
	RunnersProcessed int    `json:"runners_processed"`
	EmittersCreated  int    `json:"emitters_created"`
	AlreadyPresent   int    `json:"already_present"`
	Errors           int    `json:"errors"`
	Cursor           string `json:"cursor"`
	Done             bool   `json:"done"`
}
