package app

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
)

// generic statuses
type Status string

// define standard statuses
const (
	StatusError      Status = "error"
	StatusPending    Status = "pending"
	StatusInProgress Status = "in-progress"
	StatusSuccess    Status = "success"
)

// type specific statuses
const (
	InstallStackVersionStatusGenerating   Status = "generating"
	InstallStackVersionStatusPendingUser  Status = "pending-user"
	InstallStackVersionStatusProvisioning Status = "provisioning"
	InstallStackVersionStatusActive       Status = "active"
	InstallStackVersionStatusOutdated     Status = "outdated"
)

const (
	WorkflowStepApprovalStatusAwaitingResponse Status = "awaiting-response"
)

func (s Status) DefaultHumanDescription() string {
	switch s {
	case StatusError:
		return "error"
	case StatusPending:
		return "pending"
	case StatusInProgress:
		return "pending"
	}

	return string(s)
}

func NewCompositeStatus(ctx context.Context, status Status) CompositeStatus {
	return CompositeStatus{
		CreatedByID: createdByIDFromContext(ctx),
		CreatedAtTS: time.Now().Unix(),
		Status:      status,
		Metadata:    make(map[string]any, 0),
	}
}

func NewCompositeTemporalStatus(ctx workflow.Context, status Status, vals ...map[string]any) CompositeStatus {
	metadata := make(map[string]any, 0)
	for _, val := range vals {
		for k, v := range val {
			metadata[k] = v
		}
	}

	return CompositeStatus{
		CreatedByID: createdByIDFromTemporalContext(ctx),
		CreatedAtTS: time.Now().Unix(),
		Status:      status,
		Metadata:    metadata,
	}
}

type CompositeStatus struct {
	CreatedByID string `json:"created_by_id,omitempty"`
	CreatedAtTS int64  `json:"created_at_ts,omitempty"`

	Status                 Status         `json:"status,omitempty"`
	StatusHumanDescription string         `json:"status_human_description,omitempty"`
	Metadata               map[string]any `json:"metadata,omitempty"`

	History []CompositeStatus `json:"history,omitempty"`
}

// Scan implements the database/sql.Scanner interface.
func (c *CompositeStatus) Scan(v interface{}) (err error) {
	switch v := v.(type) {
	case nil:
		return nil
	case []byte:
		if err := json.Unmarshal(v, c); err != nil {
			return errors.Wrap(err, "unable to scan composite status")
		}
	}
	return
}

// Value implements the driver.Valuer interface.
func (c *CompositeStatus) Value() (driver.Value, error) {
	return json.Marshal(c)
}

func (CompositeStatus) GormDataType() string {
	return "jsonb"
}
