// Package phonehomesecretbackfill holds the shared constants and payload types for
// the phone-home secret backfill workflow, so the ctl-api service (which starts and
// queries the workflow) and the worker (which implements it) can share them without
// an import cycle.
//
// The backfill provisions the Secrets Manager entry, the token map and the
// cross-account grant for installs that already exist. Note it is deliberately
// pre-provisioning, not enforcement: an already-deployed phone-home Lambda has no
// NUON_PHONE_HOME_* environment variables and will not send a token until its stack
// version is regenerated and re-applied by the customer.
package phonehomesecretbackfill

import "time"

const (
	// WorkflowName must match the registered workflow function name.
	WorkflowName = "BackfillPhoneHomeSecrets"
	// WorkflowID is fixed so a re-triggered backfill reuses the running one rather
	// than starting a duplicate.
	WorkflowID = "installs-phone-home-secret-backfill"

	ProgressQueryType = "progress"

	// DefaultBatchSize is how many installs each activity call reconciles. Kept
	// small because each install can make several AWS calls.
	DefaultBatchSize = 10
)

// Request is the orchestrator input. Callers leave it zero; the workflow fills in
// the cursor and running totals and carries them across continue-as-new.
type Request struct {
	// CursorCreatedAt / CursorID form a keyset cursor ordered by (created_at, id)
	// ascending, so the fleet is processed oldest-first. The id is a tiebreaker for
	// installs sharing a created_at. Zero values start from the beginning.
	CursorCreatedAt time.Time `json:"cursor_created_at"`
	CursorID        string    `json:"cursor_id"`
	// BatchSize overrides DefaultBatchSize when non-zero.
	BatchSize int `json:"batch_size"`

	InstallsProcessed int `json:"installs_processed"`
	SecretsEnsured    int `json:"secrets_ensured"`
	TokensMinted      int `json:"tokens_minted"`
	Skipped           int `json:"skipped"`
	Errors            int `json:"errors"`
}

// Progress is the live snapshot returned by the workflow's progress query.
type Progress struct {
	InstallsProcessed int       `json:"installs_processed"`
	SecretsEnsured    int       `json:"secrets_ensured"`
	TokensMinted      int       `json:"tokens_minted"`
	Skipped           int       `json:"skipped"`
	Errors            int       `json:"errors"`
	CursorCreatedAt   time.Time `json:"cursor_created_at"`
	CursorID          string    `json:"cursor_id"`
	Done              bool      `json:"done"`
}
