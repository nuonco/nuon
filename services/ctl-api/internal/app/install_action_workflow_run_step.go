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
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string                `json:"created_by_id" gorm:"not null;default:null"`
	CreatedBy   Account               `json:"-"`
	CreatedAt   time.Time             `json:"created_at" gorm:"notnull"`
	UpdatedAt   time.Time             `json:"updated_at" gorm:"notnull"`
	DeletedAt   soft_delete.DeletedAt `json:"-"`

	// used for RLS
	OrgID string `json:"org_id" gorm:"notnull" swaggerignore:"true"`
	Org   Org    `json:"-" faker:"-"`

	Status InstallActionWorkflowRunStepStatus `json:"status"`

	InstallActionWorkflowRunID string                   `json:"install_action_workflow_run_id"`
	InstallActionWorkflowRun   InstallActionWorkflowRun `json:"-"`

	StepID string                   `json:"step_id"`
	Step   ActionWorkflowStepConfig `json:"-"`

	ExecutionDuration time.Duration `json:"execution_duration" gorm:"default null;not null" swaggertype:"primitive,integer"`
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
