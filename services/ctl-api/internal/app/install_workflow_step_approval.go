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
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string                `json:"created_by_id" gorm:"not null;default:null"`
	CreatedBy   Account               `json:"-"`
	CreatedAt   time.Time             `json:"created_at" gorm:"notnull"`
	UpdatedAt   time.Time             `json:"updated_at" gorm:"notnull"`
	DeletedAt   soft_delete.DeletedAt `gorm:"index:idx_app_install_name,unique" json:"-"`

	// used for RLS
	OrgID string `json:"org_id" gorm:"notnull" swaggerignore:"true"`
	Org   Org    `json:"-" faker:"-"`

	// the step that this approval belongs too
	InstallWorkflowStepID string

	// the runner job where this approval was created
	RunnerJobID string
	RunnerJob   RunnerJob

	// status of an approval is either pending, awaiting-response or done.

	Status CompositeStatus `json:"status"`

	// the plan and which type it is here
	Type              InstallWorkflowStepApprovalType `json:"type"`
	TerraformPlanJSON string                          `gorm:"jsonb"`
	HelmPlanJSON      string                          `gorm:"jsonb"`
	ImageApprovalJSON string                          `gorm:"jsonb"`

	// the response object must be created by the user in the UI or CLI
	Response *InstallWorkflowStepApprovalResponse `json:"response"`
}
