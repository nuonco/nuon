package app

import (
	"time"

	"gorm.io/plugin/soft_delete"
)

type InstallWorkflowApprovalResponseType string

const (
	InstallWorkflowApprovalResponseTypeDenyAll    InstallWorkflowApprovalResponseType = "deny_all"
	InstallWorkflowApprovalResponseTypeApproveAll InstallWorkflowApprovalResponseType = "approve_all"
	InstallWorkflowApprovalResponseTypeManual     InstallWorkflowApprovalResponseType = "manual"
	InstallWorkflowApprovalResponseTypeNone       InstallWorkflowApprovalResponseType = "none"
	InstallWorkflowApprovalResponseUnknown        InstallWorkflowApprovalResponseType = ""
)

type InstallWorkflowApprovalResponse struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `gorm:"index:idx_app_install_name,unique" json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	// used for RLS
	OrgID string `json:"org_id" gorm:"notnull" swaggerignore:"true" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `json:"-" faker:"-" temporaljson:"org,omitzero,omitempty"`

	InstallWorkflowApprovalID string `json:"install_workflow_approval_id" gorm:"notnull;default null" temporaljson:"install_workflow_approval_id,omitzero,omitempty"`
}
