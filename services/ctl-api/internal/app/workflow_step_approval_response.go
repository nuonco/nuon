package app

import (
	"database/sql"
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/migrations"
)

type WorkflowStepResponseType string

const (
	WorkflowStepApprovalResponseTypeDeny                     WorkflowStepResponseType = "deny"
	WorkflowStepApprovalResponseTypeApprove                  WorkflowStepResponseType = "approve"
	WorkflowStepApprovalResponseTypeSkipCurrent              WorkflowStepResponseType = "deny-skip-current"
	WorkflowStepApprovalResponseTypeSkipCurrentAndDependents WorkflowStepResponseType = "deny-skip-current-and-dependents"
	WorkflowStepApprovalResponseTypeRetryPlan                WorkflowStepResponseType = "retry"

	// auto approve is when the workflow uses auto-approve
	WorkflowStepApprovalResponseTypeAutoApprove WorkflowStepResponseType = "auto-approve"
)

type WorkflowStepApprovalResponse struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `gorm:"index:idx_app_install_name,unique" json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	// used for RLS
	OrgID string `json:"org_id,omitzero" gorm:"notnull" swaggerignore:"true" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `json:"-" faker:"-" temporaljson:"org,omitzero,omitempty"`

	// the approval the response belongs to

	InstallWorkflowStepApprovalID string               `json:"install_workflow_step_approval_id,omitzero" temporaljson:"install_workflow_step_approval_id,omitzero,omitempty"`
	InstallWorkflowStepApproval   WorkflowStepApproval `json:"-" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"install_workflow_step_approval,omitzero,omitempty"`

	// the response type
	Type WorkflowStepResponseType `json:"type,omitzero" temporaljson:"type,omitzero,omitempty" swaggertype:"string"`

	Note string `json:"note,omitzero" temporaljson:"note,omitzero,omitempty"`
}

func (c *WorkflowStepApprovalResponse) TableName() string {
	return "install_workflow_step_approval_responses"
}

func (c *WorkflowStepApprovalResponse) BeforeCreate(tx *gorm.DB) error {
	c.ID = domains.NewWorkflowStepApprovalID()

	if c.CreatedByID == "" {
		c.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}

	if c.OrgID == "" {
		c.OrgID = orgIDFromContext(tx.Statement.Context)
	}
	return nil
}

func (c *WorkflowStepApprovalResponse) AfterQuery(tx *gorm.DB) error {
	return nil
}

func (c *WorkflowStepApprovalResponse) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: "idx_install_workflow_step_approval_responses_uq",
			Columns: []string{
				"install_workflow_step_approval_id",
				"deleted_at",
			},
			UniqueValue: sql.NullBool{Bool: true, Valid: true},
		},
		{
			Name: indexes.Name(db, &WorkflowStepApprovalResponse{}, "org_id"),
			Columns: []string{
				"org_id",
			},
		},
	}
}
