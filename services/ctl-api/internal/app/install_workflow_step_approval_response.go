package app

import (
	"time"

	"gorm.io/plugin/soft_delete"
)

type InstallWorkflowStepResponseType string

const (
	InstallWorkflowStepApprovalResponseTypeDeny    InstallWorkflowApprovalResponseType = "deny"
	InstallWorkflowStepApprovalResponseTypeApprove InstallWorkflowApprovalResponseType = "approve"
	InstallWorkflowStepApprovalResponseTypeSkip    InstallWorkflowApprovalResponseType = "skip"

	// auto approve is when the workflow uses auto-approve
	InstallWorkflowStepApprovalResponseTypeAutoApprove InstallWorkflowApprovalResponseType = "auto-approve"
)

type InstallWorkflowStepApprovalResponse struct {
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
	InstallWorkflowStepApprovalID string

	// the response type
	Type InstallWorkflowStepResponseType
}
