package app

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
)

type AdHocStepConfig ActionWorkflowStepConfig

func (a *AdHocStepConfig) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, a)
}

// Value implements driver.Valuer for database serialization to JSONB.
// See queue/signal/db/signal.go for a more complex example with Type() info.
func (a AdHocStepConfig) Value() (driver.Value, error) {
	return json.Marshal(a)
}

type InstallActionWorkflowRunStepStatus string

const (
	InstallActionWorkflowRunStepStatusFinished   InstallActionWorkflowRunStepStatus = "finished"
	InstallActionWorkflowRunStepStatusPending    InstallActionWorkflowRunStepStatus = "pending"
	InstallActionWorkflowRunStepStatusInProgress InstallActionWorkflowRunStepStatus = "in-progress"
	InstallActionWorkflowRunStepStatusTimedOut   InstallActionWorkflowRunStepStatus = "timed-out"
	InstallActionWorkflowRunStepStatusError      InstallActionWorkflowRunStepStatus = "error"
)

type InstallActionWorkflowRunStep struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	// used for RLS
	OrgID string `json:"org_id,omitzero" gorm:"notnull" swaggerignore:"true" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `json:"-" faker:"-" temporaljson:"org,omitzero,omitempty"`

	Status InstallActionWorkflowRunStepStatus `json:"status,omitzero" temporaljson:"status,omitzero,omitempty"`

	InstallActionWorkflowRunID string                   `json:"install_action_workflow_run_id,omitzero" temporaljson:"install_action_workflow_run_id,omitzero,omitempty"`
	InstallActionWorkflowRun   InstallActionWorkflowRun `json:"-" temporaljson:"install_action_workflow_run,omitzero,omitempty"`

	StepID generics.NullString      `json:"step_id,omitzero" temporaljson:"step_id,omitzero,omitempty"`
	Step   ActionWorkflowStepConfig `json:"-" temporaljson:"step,omitzero,omitempty"`

	AdHocConfig *AdHocStepConfig `json:"adhoc_config,omitzero" gorm:"type:jsonb" temporaljson:"adhoc_config,omitzero,omitempty"`

	ExecutionDuration time.Duration `json:"execution_duration,omitzero" gorm:"default null;not null" swaggertype:"primitive,integer" temporaljson:"execution_duration,omitzero,omitempty"`
}

func (i *InstallActionWorkflowRunStep) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &InstallActionWorkflowRunStep{}, "org_id"),
			Columns: []string{
				"org_id",
			},
		},
	}
}

func (i *InstallActionWorkflowRunStep) BeforeCreate(tx *gorm.DB) error {
	if i.ID == "" {
		i.ID = domains.NewInstallActionWorkflowRunID()
	}

	if i.CreatedByID == "" {
		i.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	if i.OrgID == "" {
		i.OrgID = orgIDFromContext(tx.Statement.Context)
	}
	return nil
}
