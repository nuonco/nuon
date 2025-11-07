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

type WorkflowStepApprovalType string

const (
	NoopApprovalType               WorkflowStepApprovalType = "noop"
	ApproveAllApprovalType         WorkflowStepApprovalType = "approve-all"
	TerraformPlanApprovalType      WorkflowStepApprovalType = "terraform_plan"
	KubernetesManifestApprovalType WorkflowStepApprovalType = "kubernetes_manifest_approval"
	HelmApprovalApprovalType       WorkflowStepApprovalType = "helm_approval"
)

type WorkflowStepApproval struct {
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
	InstallWorkflowStepID string       `gorm:"install_workflow_step_id,notnull" temporaljson:"install_workflow_step_id,omitzero,omitempty"`
	InstallWorkflowStep   WorkflowStep `temporaljson:"install_workflow_step,omitzero,omitempty"`

	// the runner job where this approval was created
	RunnerJobID *string    `json:"runner_job_id,omitzero" temporaljson:"runner_job_id,omitzero,omitempty"`
	RunnerJob   *RunnerJob `json:"runner_job,omitzero" temporaljson:"runner_job,omitzero,omitempty"`

	OwnerID   string `json:"owner_id,omitzero" gorm:"type:text;check:owner_id_checker,char_length(id)=26;index:idx_runner_jobs_owner_id,priority:1" temporaljson:"owner_id,omitzero,omitempty"`
	OwnerType string `json:"owner_type,omitzero" gorm:"type:text;" temporaljson:"owner_type,omitzero,omitempty"`

	Contents string `json:"-" temporaljson:"-"`

	Type WorkflowStepApprovalType `json:"type"`

	// the response object must be created by the user in the UI or CLI

	Response *WorkflowStepApprovalResponse `gorm:"foreignKey:InstallWorkflowStepApprovalID" json:"response,omitzero" temporaljson:"response,omitzero,omitempty" swaggertype:"object,string"`

	// afterquery
	WorkflowStepID string       `json:"workflow_step_id,omitzero" gorm:"-" temporaljson:"workflow_step_id,omitzero,omitempty"`
	WorkflowStep   WorkflowStep `json:"workflow_step,omitzero" gorm:"-" temporaljson:"workflow_step,omitzero,omitempty"`
}

func (c *WorkflowStepApproval) TableName() string {
	return "install_workflow_step_approvals"
}

func (c *WorkflowStepApproval) BeforeCreate(tx *gorm.DB) error {
	c.ID = domains.NewWorkflowStepApprovalID()

	if c.CreatedByID == "" {
		c.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}

	if c.OrgID == "" {
		c.OrgID = orgIDFromContext(tx.Statement.Context)
	}
	return nil
}

func (c *WorkflowStepApproval) AfterQuery(tx *gorm.DB) error {
	c.WorkflowStepID = c.InstallWorkflowStep.ID
	c.WorkflowStep = c.InstallWorkflowStep
	return nil
}

func (c *WorkflowStepApproval) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: "idx_install_workflow_step_approvals_uq",
			Columns: []string{
				"install_workflow_step_id",
				"deleted_at",
			},
			UniqueValue: sql.NullBool{Bool: true, Valid: true},
		},
		{
			Name: indexes.Name(db, &WorkflowStepApproval{}, "org_id"),
			Columns: []string{
				"org_id",
			},
		},
	}
}
