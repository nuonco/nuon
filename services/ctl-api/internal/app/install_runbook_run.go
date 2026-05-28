package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
)

type InstallRunbookRunStatus string

const (
	InstallRunbookRunStatusQueued     InstallRunbookRunStatus = "queued"
	InstallRunbookRunStatusInProgress InstallRunbookRunStatus = "in-progress"
	InstallRunbookRunStatusFinished   InstallRunbookRunStatus = "finished"
	InstallRunbookRunStatusError      InstallRunbookRunStatus = "error"
	InstallRunbookRunStatusCancelled  InstallRunbookRunStatus = "cancelled"
	InstallRunbookRunStatusUnknown    InstallRunbookRunStatus = "unknown"
)

type InstallRunbookRun struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"created_by,omitzero" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" gorm:"notnull;index:idx_install_runbook_runs_query,priority:3,sort:desc" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgID string `json:"org_id,omitzero" gorm:"notnull" swaggerignore:"true" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `json:"-" faker:"-" temporaljson:"org,omitzero,omitempty"`

	InstallID string  `json:"install_id,omitzero" gorm:"not null;default null;index:idx_install_runbook_runs_query,priority:1" temporaljson:"install_id,omitzero,omitempty"`
	Install   Install `swaggerignore:"true" json:"-" temporaljson:"install,omitzero,omitempty"`

	InstallRunbookID string         `json:"install_runbook_id,omitzero" gorm:"index:idx_install_runbook_runs_query,priority:2" temporaljson:"install_runbook_id,omitzero,omitempty"`
	InstallRunbook   InstallRunbook `json:"install_runbook,omitzero" temporaljson:"install_runbook,omitzero,omitempty"`

	RunbookConfigID string        `json:"runbook_config_id,omitzero" temporaljson:"runbook_config_id,omitzero,omitempty"`
	RunbookConfig   RunbookConfig `json:"runbook_config,omitzero" temporaljson:"runbook_config,omitzero,omitempty"`

	Status            InstallRunbookRunStatus `json:"status,omitzero" gorm:"notnull" swaggertype:"string" temporaljson:"status,omitzero,omitempty"`
	StatusDescription string                  `json:"status_description,omitzero" gorm:"notnull" temporaljson:"status_description,omitzero,omitempty"`
	StatusV2          CompositeStatus         `json:"status_v2,omitzero" gorm:"type:jsonb" temporaljson:"status_v2,omitzero,omitempty"`

	TriggeredByID string `json:"triggered_by_id,omitzero" gorm:"type:text" temporaljson:"triggered_by_id,omitzero,omitempty"`

	InstallWorkflowID *string   `json:"install_workflow_id" gorm:"default null" temporaljson:"install_workflow_id,omitzero,omitempty"`
	InstallWorkflow   *Workflow `json:"install_workflow,omitzero" gorm:"foreignKey:InstallWorkflowID" temporaljson:"install_workflow,omitzero,omitempty"`

	// after query
	ExecutionTime time.Duration `json:"execution_time,omitzero" gorm:"-" swaggertype:"primitive,integer" temporaljson:"execution_time,omitzero,omitempty"`
}

func (r *InstallRunbookRun) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = domains.NewInstallRunbookRunID()
	}
	if r.CreatedByID == "" {
		r.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	if r.OrgID == "" {
		r.OrgID = orgIDFromContext(tx.Statement.Context)
	}
	return nil
}

func (r *InstallRunbookRun) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &InstallRunbookRun{}, "preload"),
			Columns: []string{
				"install_runbook_id",
				"deleted_at",
				"created_at DESC",
			},
		},
		{
			Name: indexes.Name(db, &InstallRunbookRun{}, "org_id"),
			Columns: []string{
				"org_id",
			},
		},
	}
}

func (r *InstallRunbookRun) AfterQuery(tx *gorm.DB) error {
	if r.StatusV2.Status != "" {
		r.Status = InstallRunbookRunStatus(r.StatusV2.Status)
		r.StatusDescription = r.StatusV2.StatusHumanDescription
	}
	return nil
}
