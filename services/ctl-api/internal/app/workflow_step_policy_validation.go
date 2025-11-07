package app

import (
	"time"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/migrations"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

type WorkflowStepPolicyValidation struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `gorm:"index:idx_app_install_name,unique" json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	// used for RLS
	OrgID string `json:"org_id,omitzero" gorm:"notnull" swaggerignore:"true" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `json:"-" faker:"-" temporaljson:"org,omitzero,omitempty"`

	// runnerJobID is the runner job that this was performed within
	RunnerJobID string `json:"runner_job_id,omitzero" temporaljson:"runner_job_id,omitzero,omitempty"`

	// install workflow step is the install step that this was performed within
	InstallWorkflowStepID string `json:"install_workflow_step_id,omitzero" temporaljson:"install_workflow_step_id,omitzero,omitempty"`

	// status denotes whether this passed, or whether it failed the job.
	Status CompositeStatus `json:"status,omitzero" temporaljson:"status,omitzero,omitempty"`
	// response is the kyverno response
	Response string `json:"response,omitzero" gorm:"jsonb" temporaljson:"response,omitzero,omitempty"`
}

func (v *WorkflowStepPolicyValidation) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &WorkflowStepPolicyValidation{}, "org_id"),
			Columns: []string{
				"org_id",
			},
		},
	}
}

func (v *WorkflowStepPolicyValidation) TableName() string {
	return "install_workflow_step_policy_validations"
}
