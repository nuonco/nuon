package app

import (
	"database/sql"
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
)

type InstallActionWorkflow struct {
	ID          string                `json:"id,omitzero" gorm:"primary_key;check:id_checker,char_length(id)=26" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	// used for RLS
	OrgID string `json:"org_id,omitzero" gorm:"notnull" swaggerignore:"true" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `json:"-" faker:"-" temporaljson:"org,omitzero,omitempty"`

	Install   Install `json:"-" swaggerignore:"true" temporaljson:"install,omitzero,omitempty"`
	InstallID string  `json:"install_id,omitzero" faker:"-" temporaljson:"install_id,omitzero,omitempty"`

	ActionWorkflow   ActionWorkflow `json:"action_workflow,omitzero" temporaljson:"action_workflow,omitzero,omitempty"`
	ActionWorkflowID string         `json:"action_workflow_id,omitzero" temporaljson:"action_workflow_id,omitzero,omitempty"`

	Runs []InstallActionWorkflowRun `faker:"-" gorm:"constraint:OnDelete:CASCADE;" json:"runs,omitzero" temporaljson:"runs,omitzero,omitempty"`

	// after query fields filled in after querying
	Status InstallActionWorkflowRunStatus `json:"status,omitzero" gorm:"-" swaggertype:"string" temporaljson:"status,omitzero,omitempty"`
}

func (a *InstallActionWorkflow) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &InstallActionWorkflow{}, "org_id"),
			Columns: []string{
				"org_id",
			},
		},
		{
			Name:        "idx_install_action_workflow_id",
			Columns:     []string{"deleted_at", "install_id", "action_workflow_id"},
			UniqueValue: sql.NullBool{Bool: true, Valid: true},
		},
	}
}

func (a *InstallActionWorkflow) BeforeCreate(tx *gorm.DB) error {
	a.ID = domains.NewInstallActionWorkflowConfigID()
	a.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	a.OrgID = orgIDFromContext(tx.Statement.Context)
	return nil
}

func (c *InstallActionWorkflow) AfterQuery(tx *gorm.DB) error {
	c.Status = InstallActionRunStatusUnknown
	if len(c.Runs) > 0 {
		c.Status = c.Runs[0].Status
	}

	return nil
}
