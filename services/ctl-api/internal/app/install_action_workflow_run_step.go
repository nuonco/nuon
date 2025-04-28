package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
)

type InstallActionWorkflowRunStepStatus string

const (
	InstallActionWorkflowRunStepStatusFinished   InstallActionWorkflowRunStepStatus = "finished"
	InstallActionWorkflowRunStepStatusPending    InstallActionWorkflowRunStepStatus = "pending"
	InstallActionWorkflowRunStepStatusInProgress InstallActionWorkflowRunStepStatus = "in-progress"
	InstallActionWorkflowRunStepStatusTimedOut   InstallActionWorkflowRunStepStatus = "timed-out"
	InstallActionWorkflowRunStepStatusError      InstallActionWorkflowRunStepStatus = "error"
)

type InstallActionWorkflowRunStep struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	// used for RLS
	OrgID string `json:"org_id" gorm:"notnull" swaggerignore:"true" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `json:"-" faker:"-" temporaljson:"org,omitzero,omitempty"`

	Status InstallActionWorkflowRunStepStatus `json:"status" temporaljson:"status,omitzero,omitempty"`

	InstallActionWorkflowRunID string                   `json:"install_action_workflow_run_id" temporaljson:"install_action_workflow_run_id,omitzero,omitempty"`
	InstallActionWorkflowRun   InstallActionWorkflowRun `json:"-" temporaljson:"install_action_workflow_run,omitzero,omitempty"`

	StepID string                   `json:"step_id" temporaljson:"step_id,omitzero,omitempty"`
	Step   ActionWorkflowStepConfig `json:"-" temporaljson:"step,omitzero,omitempty"`

	ExecutionDuration time.Duration `json:"execution_duration" gorm:"default null;not null" swaggertype:"primitive,integer" temporaljson:"execution_duration,omitzero,omitempty"`
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
