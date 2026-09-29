package callback

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"go.temporal.io/sdk/workflow"
)

// Ref describes where to send a Temporal signal when an operation completes.
// It is stored as a JSONB column on QueueSignal and passed through the
// enqueue → handler → completion chain.
type Ref struct {
	WorkflowID string `json:"workflow_id,omitempty"`
	SignalName string `json:"signal_name,omitempty"`
	Namespace  string `json:"namespace,omitempty"`
}

func (c Ref) IsSet() bool {
	return c.WorkflowID != ""
}

// why: New creates a Ref pointing back to the current workflow with a deterministic
// signal name derived from id. The caller should register the signal channel
// before the handler starts to avoid races.
func New(ctx workflow.Context, id string) Ref {
	info := workflow.GetInfo(ctx)
	return Ref{
		WorkflowID: info.WorkflowExecution.ID,
		SignalName: SignalName(id),
		Namespace:  info.Namespace,
	}
}

func SignalName(id string) string {
	return fmt.Sprintf("signal-complete-%s", id)
}

// why: NewAttempt scopes the signal name per dispatch to avoid stale completions.
func NewAttempt(ctx workflow.Context, id string, attempt int) Ref {
	return New(ctx, fmt.Sprintf("%s#%d", id, attempt))
}

func (c *Ref) Scan(v interface{}) error {
	switch v := v.(type) {
	case nil:
		return nil
	case []byte:
		return json.Unmarshal(v, c)
	}
	return nil
}

func (c Ref) Value() (driver.Value, error) {
	return json.Marshal(c)
}

func (Ref) GormDataType() string {
	return "jsonb"
}

type Refs []Ref

func (r Refs) IsSet() bool {
	return len(r) > 0
}

func (r *Refs) Add(ref Ref) {
	if ref.IsSet() {
		*r = append(*r, ref)
	}
}

func (r *Refs) Scan(v interface{}) error {
	switch v := v.(type) {
	case nil:
		return nil
	case []byte:
		return json.Unmarshal(v, r)
	}
	return nil
}

func (r Refs) Value() (driver.Value, error) {
	if r == nil {
		return nil, nil
	}
	return json.Marshal(r)
}

func (Refs) GormDataType() string {
	return "jsonb"
}
