package app

import (
	"time"

	"gorm.io/plugin/soft_delete"
)

type InstallWorkflowStepPolicyValidation struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `gorm:"index:idx_app_install_name,unique" json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	// used for RLS
	OrgID string `json:"org_id" gorm:"notnull" swaggerignore:"true" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `json:"-" faker:"-" temporaljson:"org,omitzero,omitempty"`

	// runnerJobID is the runner job that this was performed within
	RunnerJobID string `temporaljson:"runner_job_id,omitzero,omitempty"`

	// install workflow step is the install step that this was performed within
	InstallWorkflowStepID string `temporaljson:"install_workflow_step_id,omitzero,omitempty"`

	// status denotes whether this passed, or whether it failed the job.
	Status CompositeStatus `json:"status" temporaljson:"status,omitzero,omitempty"`
	// response is the kyverno response
	Response string `gorm:"jsonb" temporaljson:"response,omitzero,omitempty"`
}
