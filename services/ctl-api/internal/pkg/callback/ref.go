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

// IsSet returns true if the callback has a target workflow.
func (c Ref) IsSet() bool {
	return c.WorkflowID != ""
}

// New creates a Ref pointing back to the current workflow with a deterministic
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

// SignalName returns the deterministic signal name for a given id.
func SignalName(id string) string {
	return fmt.Sprintf("signal-complete-%s", id)
}

// Scan implements database/sql.Scanner for reading JSONB from PostgreSQL.
func (c *Ref) Scan(v interface{}) error {
	switch v := v.(type) {
	case nil:
		return nil
	case []byte:
		return json.Unmarshal(v, c)
	}
	return nil
}

// Value implements driver.Valuer for writing JSONB to PostgreSQL.
func (c Ref) Value() (driver.Value, error) {
	return json.Marshal(c)
}

// GormDataType tells GORM to use the jsonb PostgreSQL type.
func (Ref) GormDataType() string {
	return "jsonb"
}
