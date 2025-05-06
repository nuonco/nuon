package app

import (
	"time"

	"gorm.io/plugin/soft_delete"
)

type InstallWorkflowStepApprovalType string

const (
	NoopApprovalType InstallWorkflowStepApprovalType = "noop"

	TerraformPlanApprovalType InstallWorkflowStepApprovalType = "terraform_plan"
	HelmApprovalApprovalType  InstallWorkflowStepApprovalType = "helm_approval"
	ImageApprovalApprovalType InstallWorkflowStepApprovalType = "image_approval"
)

type InstallWorkflowStepApproval struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `gorm:"index:idx_app_install_name,unique" json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	// used for RLS
	OrgID string `json:"org_id,omitzero" gorm:"notnull" swaggerignore:"true" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `json:"-" faker:"-" temporaljson:"org,omitzero,omitempty"`

	// the step that this approval belongs too
	InstallWorkflowStepID string `temporaljson:"install_workflow_step_id,omitzero,omitempty"`

	// the runner job where this approval was created
	RunnerJobID string    `temporaljson:"runner_job_id,omitzero,omitempty"`
	RunnerJob   RunnerJob `temporaljson:"runner_job,omitzero,omitempty"`

	// status of an approval is either pending, awaiting-response or done.

	Status CompositeStatus `json:"status,omitzero" temporaljson:"status,omitzero,omitempty"`

	// the plan and which type it is here
	Type              InstallWorkflowStepApprovalType `json:"type,omitzero" temporaljson:"type,omitzero,omitempty"`
	TerraformPlanJSON string                          `gorm:"jsonb" temporaljson:"terraform_plan_json,omitzero,omitempty"`
	HelmPlanJSON      string                          `gorm:"jsonb" temporaljson:"helm_plan_json,omitzero,omitempty"`
	ImageApprovalJSON string                          `gorm:"jsonb" temporaljson:"image_approval_json,omitzero,omitempty"`

	// the response object must be created by the user in the UI or CLI
	Response *InstallWorkflowStepApprovalResponse `json:"response,omitzero" temporaljson:"response,omitzero,omitempty"`
}
