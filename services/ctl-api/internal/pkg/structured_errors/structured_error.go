package structured_errors

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/pkg/errors"
)

type OwnerType string

const (
	OwnerTypePlan           OwnerType = "plan"
	OwnerTypeApply          OwnerType = "apply"
	OwnerTypeActionRun      OwnerType = "action-run"
	OwnerTypeVariableRender OwnerType = "variable-renderer"
	OwnerTypeRunner         OwnerType = "runner"
	OwnerTypeK8sDiagnostics OwnerType = "k8s-diagnostics"
)

type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityWarning  Severity = "warning"
	SeverityInfo     Severity = "info"
)

type CompositeError struct {
	CreatedByID  string           `json:"created_by_id,omitzero,omitempty"`
	CreatedAtTS  int64            `json:"created_at_ts,omitzero,omitempty"`
	OwnerID   string    `json:"owner_id,omitempty"`
	OwnerType OwnerType `json:"owner_type"`
	Severity     Severity         `json:"severity"`
	Summary      string           `json:"summary"`
	Detail       string           `json:"detail,omitempty"`
	Metadata     map[string]any   `json:"metadata,omitempty"`
	History      []CompositeError `json:"history,omitzero,omitempty"`
}

type CompositeErrors []CompositeError

// Scan implements the database/sql.Scanner interface.
func (c *CompositeErrors) Scan(v interface{}) (err error) {
	switch v := v.(type) {
	case nil:
		return nil
	case []byte:
		if err := json.Unmarshal(v, c); err != nil {
			return errors.Wrap(err, "unable to scan composite errors")
		}
	}
	return
}

// Value implements the driver.Valuer interface.
func (c CompositeErrors) Value() (driver.Value, error) {
	if c == nil {
		return nil, nil
	}
	return json.Marshal(c)
}

// GormDataType returns the GORM data type for this field.
func (CompositeErrors) GormDataType() string {
	return "jsonb"
}
