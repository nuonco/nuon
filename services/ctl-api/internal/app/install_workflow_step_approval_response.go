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
	InstallWorkflowStepApprovalID string `json:"install_workflow_step_approval_id,omitzero" temporaljson:"install_workflow_step_approval_id,omitzero,omitempty"`

	// the response type
	Type InstallWorkflowStepResponseType `json:"type,omitzero" temporaljson:"type,omitzero,omitempty"`
}
