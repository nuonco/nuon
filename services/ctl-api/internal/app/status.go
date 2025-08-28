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
	StatusError        Status = "error"
	StatusPending      Status = "pending"
	StatusInProgress   Status = "in-progress"
	StatusSuccess      Status = "success"
	StatusNotAttempted Status = "not-attempted"
	StatusCancelled    Status = "cancelled"
	StatusRetrying     Status = "retrying"
	StatusDiscarded    Status = "discarded"
	StatusUserSkipped  Status = "user-skipped"
	StatusAutoSkipped  Status = "auto-skipped"
	StatusPlanning     Status = "planning"
	StatusApplying     Status = "applying"
)

// type specific statuses
const (
	InstallStackVersionStatusGenerating   Status = "generating"
	InstallStackVersionStatusPendingUser  Status = "awaiting-user-run"
	InstallStackVersionStatusProvisioning Status = "provisioning"
	InstallStackVersionStatusActive       Status = "active"
	InstallStackVersionStatusOutdated     Status = "outdated"
	InstallStackVersionStatusExpired      Status = "expired"
)

const (
	WorkflowStepApprovalStatusApproved Status = "approved"
	// WorkflowStepApprovalStatusAwaitingResponse  Status = "approval-awaiting" // NOTE(fd): superceded by shared const below
	WorkflowStepApprovalStatusApprovalExpired   Status = "approval-expired"
	WorkflowStepApprovalStatusApprovalDenied    Status = "approval-denied"
	WorkflowStepApprovalStatusApprovalRetryPlan Status = "approval-retry"
)

// component build specific statuses
const (
	StatusBuilding Status = "building"
	StatusDeleting Status = "deleting"
)

// release specific statuses
const (
	StatusProvisioning   ReleaseStatus = "provisioning"
	StatusDeprovisioning ReleaseStatus = "deprovisioning"
	StatusSyncing        ReleaseStatus = "syncing"
	StatusExecuting      ReleaseStatus = "executing"
)

const (
	InstallDeployStatusV2Noop Status = "noop"
)

// const (
// 	WorkflowAwaitingApproval Status = "approval-awaiting" // NOTE(fd): superceded by shared const below
// )

// shared by WorkflowStep and Workflow
const (
	AwaitingApproval Status = "approval-awaiting"
)

func (s Status) DefaultHumanDescription() string {
	switch s {
	case StatusError:
		return "error"
	case StatusPending:
		return "pending"
	case StatusInProgress:
		return "in-progress"
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
	CreatedByID string `json:"created_by_id,omitzero,omitempty" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedAtTS int64  `json:"created_at_ts,omitzero,omitempty" temporaljson:"created_at_ts,omitzero,omitempty"`

	Status                 Status         `json:"status,omitzero,omitempty" temporaljson:"status,omitzero,omitempty"`
	StatusHumanDescription string         `json:"status_human_description,omitzero,omitempty" temporaljson:"status_human_description,omitzero,omitempty"`
	Metadata               map[string]any `json:"metadata,omitzero,omitempty" temporaljson:"metadata,omitzero,omitempty"`

	History []CompositeStatus `json:"history,omitzero,omitempty" temporaljson:"history,omitzero,omitempty"`
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
