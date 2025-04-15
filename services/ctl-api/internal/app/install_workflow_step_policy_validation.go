package app

import (
	"time"

	"gorm.io/plugin/soft_delete"
)

type InstallWorkflowStepPolicyValidation struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string                `json:"created_by_id" gorm:"not null;default:null"`
	CreatedBy   Account               `json:"-"`
	CreatedAt   time.Time             `json:"created_at" gorm:"notnull"`
	UpdatedAt   time.Time             `json:"updated_at" gorm:"notnull"`
	DeletedAt   soft_delete.DeletedAt `gorm:"index:idx_app_install_name,unique" json:"-"`

	// used for RLS
	OrgID string `json:"org_id" gorm:"notnull" swaggerignore:"true"`
	Org   Org    `json:"-" faker:"-"`

	// runnerJobID is the runner job that this was performed within
	RunnerJobID string

	// install workflow step is the install step that this was performed within
	InstallWorkflowStepID string

	// status denotes whether this passed, or whether it failed the job.
	Status CompositeStatus `json:"status"`
	// response is the kyverno response
	Response string `gorm:"jsonb"`
}
