package statestore

import (
	"encoding/json"
	"errors"
	"time"
)

const (
	RunStatusInProgress = "in-progress"
	RunStatusFinished   = "finished"
	RunStatusFailed     = "failed"
)

type Status struct {
	InstallID   string                     `json:"install_id"`
	RunID       string                     `json:"run_id"`
	Status      string                     `json:"status"`
	FailedStep  string                     `json:"failed_step,omitempty"`
	StartedAt   time.Time                  `json:"started_at"`
	FinishedAt  *time.Time                 `json:"finished_at,omitempty"`
	HeartbeatAt time.Time                  `json:"heartbeat_at"`
	Steps       []StepStatus               `json:"steps"`
	Outputs     map[string]json.RawMessage `json:"outputs,omitempty"`
}

type StepStatus struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Status      string     `json:"status"`
	ExecutionID string     `json:"execution_id,omitempty"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	FinishedAt  *time.Time `json:"finished_at,omitempty"`
	Error       string     `json:"error,omitempty"`
}

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
