package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/nuonco/nuon/pkg/shortid/domains"
)

type WorkflowRunType string

const (
	WorkflowRunTypeInitial WorkflowRunType = "initial"
	WorkflowRunTypeRetry   WorkflowRunType = "retry"
	WorkflowRunTypeSkip    WorkflowRunType = "skip"
	WorkflowRunTypeResume  WorkflowRunType = "resume"
)

type WorkflowRun struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null"`
	CreatedAt   time.Time             `json:"created_at,omitzero" gorm:"notnull"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" gorm:"notnull"`
	DeletedAt   soft_delete.DeletedAt `json:"-"`

	OrgID string `json:"org_id,omitzero" gorm:"notnull" swaggerignore:"true"`

	WorkflowID string          `json:"workflow_id,omitzero" gorm:"not null"`
	Type       WorkflowRunType `json:"type,omitzero" gorm:"not null"`
	Status     CompositeStatus `json:"status,omitzero"`

	// TriggerStepID is the step that triggered this run (empty for initial runs).
	TriggerStepID string `json:"trigger_step_id,omitempty"`

	// StartFromIdx is the step index to start execution from.
	StartFromIdx int `json:"start_from_idx"`

	StartedAt  time.Time `json:"started_at,omitzero" gorm:"default:null"`
	FinishedAt time.Time `json:"finished_at,omitzero" gorm:"default:null"`
}

func (r *WorkflowRun) TableName() string {
	return "workflow_runs"
}

func (r *WorkflowRun) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = domains.NewWorkflowRunID()
	}

	r.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	r.OrgID = orgIDFromContext(tx.Statement.Context)

	return nil
}
