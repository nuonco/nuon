package app

import (
	"time"

	"gorm.io/plugin/soft_delete"
)

type FlowStepPolicyValidation struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `gorm:"index" json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	// used for RLS
	OrgID string `json:"org_id,omitzero" gorm:"notnull" swaggerignore:"true" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `json:"-" faker:"-" temporaljson:"org,omitzero,omitempty"`

	// runnerJobID is the runner job that this was performed within
	RunnerJobID string `json:"runner_job_id,omitzero" temporaljson:"runner_job_id,omitzero,omitempty"`

	// flow step is the step that this was performed within
	FlowStepID string `json:"flow_step_id,omitzero" temporaljson:"flow_step_id,omitzero,omitempty"`

	// status denotes whether this passed, or whether it failed the job.
	Status CompositeStatus `json:"status,omitzero" temporaljson:"status,omitzero,omitempty"`
	// response is the kyverno response
	Response string `json:"response,omitzero" gorm:"jsonb" temporaljson:"response,omitzero,omitempty"`
}
