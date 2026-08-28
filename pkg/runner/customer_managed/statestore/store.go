package statestore

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

const (
	RunStatusInProgress         = "in-progress"
	RunStatusFailedPendingRetry = "failed-pending-retry"
	RunStatusFinished           = "finished"
	RunStatusFailed             = "failed"
	RunStatusCancelled          = "cancelled"
	RunTypeInstall              = "installation"
	RunTypeUpgrade              = "upgrade"
	InstallRunsPrefix           = "install-runs/"
	JobLogsPrefix               = "job-logs/"
)

func JobLogKey(jobID string) string { return JobLogsPrefix + jobID + ".ndjson" }

type ResultDirective string

const (
	DirectiveContinue      ResultDirective = "continue"
	DirectiveStop          ResultDirective = "stop"
	DirectiveRetryGroup    ResultDirective = "retry-group"
	DirectiveSkipGroup     ResultDirective = "skip-group"
	DirectiveAwaitApproval ResultDirective = "await-approval"
	DirectiveAwaitRetry    ResultDirective = "await-retry"
)

const (
	ControlActionRetry    = "retry"
	ControlActionUserSkip = "user-skip"
	ControlActionCancel   = "cancel"
)

func InstallRunStatusKey(runID string) string {
	return InstallRunsPrefix + runID + "/status.json"
}

func InstallRunEventsPrefix(runID string) string { return InstallRunsPrefix + runID + "/events/" }
func InstallRunEventKey(runID string, sequence uint64) string {
	return fmt.Sprintf("%s%020d.json", InstallRunEventsPrefix(runID), sequence)
}
func InstallControlKey(runID, action string) string {
	return "install-controls/" + runID + "/" + action + ".json"
}
func InstallControlHandledKey(runID, action string) string {
	return "install-controls/" + runID + "/" + action + ".handled.json"
}

type Status struct {
	InstallID        string                     `json:"install_id"`
	BundleDigest     string                     `json:"bundle_digest,omitempty"`
	RunID            string                     `json:"run_id"`
	RunType          string                     `json:"run_type,omitempty"`
	PreviousRunID    string                     `json:"previous_run_id,omitempty"`
	Status           string                     `json:"status"`
	FailedStep       string                     `json:"failed_step,omitempty"`
	StartedAt        time.Time                  `json:"started_at"`
	FinishedAt       *time.Time                 `json:"finished_at,omitempty"`
	HeartbeatAt      time.Time                  `json:"heartbeat_at"`
	Steps            []StepStatus               `json:"steps"`
	Outputs          map[string]json.RawMessage `json:"outputs,omitempty"`
	ApprovalRequired bool                       `json:"approval_required,omitempty"`
	ApprovalPhase    string                     `json:"approval_phase,omitempty"`
	ResultDirective  ResultDirective            `json:"result_directive,omitempty"`
}

type StepStatus struct {
	ID              string          `json:"id"`
	Name            string          `json:"name"`
	Status          string          `json:"status"`
	ExecutionID     string          `json:"execution_id,omitempty"`
	StartedAt       *time.Time      `json:"started_at,omitempty"`
	FinishedAt      *time.Time      `json:"finished_at,omitempty"`
	SourceRunID     string          `json:"source_run_id,omitempty"`
	Error           string          `json:"error,omitempty"`
	ResultDirective ResultDirective `json:"result_directive,omitempty"`
}

type StatusEvent struct {
	SchemaVersion int       `json:"schema_version"`
	Sequence      uint64    `json:"sequence"`
	CreatedAt     time.Time `json:"created_at"`
	Status        Status    `json:"status"`
}

type ControlRequest struct {
	RunID       string    `json:"run_id"`
	Action      string    `json:"action"`
	RequestedBy string    `json:"requested_by"`
	CreatedAt   time.Time `json:"created_at"`
}

type ControlHandled struct {
	ControlRequest
	HandledAt time.Time `json:"handled_at"`
}

// StepPlansPrefix holds the late-bound rendered plan for each bootstrap step,
// written when the runner fetches the plan for execution. The portal reads
// these to preview what a step will apply; operation raw terraform plans live
// under the separate "job-plans/" prefix.
const StepPlansPrefix = "step-plans/"

func StepPlanKey(stepID string) string { return StepPlansPrefix + stepID + ".json" }

// StepResultKey is the state-store key where WriteResult persists a bootstrap
// step's execution result (Disk.stepPath layout). For terraform steps the
// result carries the compressed `terraform plan` JSON; helm and kubernetes
// manifest steps carry their computed resource diffs. The portal decodes these
// to show the real plan/diff instead of the internal composite plan.
func StepResultKey(stepID string) string { return "steps/" + stepID + "/result.json" }

type Store interface {
	WriteFile(relPath string, data []byte) error
	ReadFile(relPath string) ([]byte, bool, error)
	ReadStatus() (*Status, error)
	WriteStatus(*Status) error
	WriteResult(stepID string, value any) error
	WriteOutputs(stepID string, value any) error
	AppendExecution(stepID string, value any) error
	ReadResult(stepID string) (json.RawMessage, bool, error)
	ReadOutputs(stepID string) (json.RawMessage, bool, error)
	ReadExecutions(stepID string) (json.RawMessage, bool, error)
	// WriteReport persists the collated post-trip report so it can be pulled
	// out of the state store after a run without reading per-step files.
	WriteReport(value any) error
	ReadReport() (json.RawMessage, bool, error)
	// Health documents replace the ctl-plane component-health API for offline
	// runs: the latest snapshot is overwritten, transitions are append-only,
	// and the context carries cluster access so a restarted runner can keep
	// observing without a new deploy.
	WriteHealth(value any) error
	ReadHealth() (json.RawMessage, bool, error)
	AppendHealthTransitions(values []any) error
	ReadHealthTransitions() (json.RawMessage, bool, error)
	WriteHealthContext(value any) error
	ReadHealthContext() (json.RawMessage, bool, error)
	GetTFState(workspaceID string) ([]byte, bool, error)
	PutTFState(workspaceID string, state []byte) error
	// PutTFStateShow stores the `terraform show -json` document the runner
	// uploads after an apply. It is introspection data (ctl-api keeps it in a
	// separate column) and must never overwrite the http-backend state.
	PutTFStateShow(workspaceID string, doc []byte) error
	LockTF(workspaceID string, lockInfoJSON []byte) error
	UnlockTF(workspaceID string) error
}

type LockConflictError struct{ Existing []byte }

func (e *LockConflictError) Error() string { return "terraform workspace is already locked" }

func IsLockConflict(err error) bool {
	var conflict *LockConflictError
	return errors.As(err, &conflict)
}
