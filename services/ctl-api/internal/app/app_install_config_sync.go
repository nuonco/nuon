package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
)

type AppInstallConfigSync struct {
	ID          string                `gorm:"primarykey;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgID string `json:"org_id,omitzero" gorm:"notnull;default null" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `faker:"-" json:"-" temporaljson:"org,omitzero,omitempty"`

	AppID string `json:"app_id,omitzero" gorm:"not null" temporaljson:"app_id,omitzero,omitempty"`
	App   App    `faker:"-" json:"-" temporaljson:"app,omitzero,omitempty"`

	VCSConnectionCommitID *string              `json:"vcs_connection_commit_id,omitempty" swaggerignore:"true" temporaljson:"vcs_connection_commit_id,omitzero,omitempty"`
	VCSConnectionCommit   *VCSConnectionCommit `json:"vcs_connection_commit,omitempty" temporaljson:"vcs_connection_commit,omitzero,omitempty"`

	TriggeredBy string `json:"triggered_by,omitzero" gorm:"not null;default:'manual'" temporaljson:"triggered_by,omitzero,omitempty"`

	QueueSignalID string `json:"queue_signal_id,omitempty" gorm:"default:null" temporaljson:"queue_signal_id,omitzero,omitempty"`
	QueueID       string `json:"queue_id,omitempty" gorm:"default:null" temporaljson:"queue_id,omitzero,omitempty"`

	WorkflowID *string   `json:"workflow_id,omitempty" gorm:"default:null" temporaljson:"workflow_id,omitzero,omitempty"`
	Workflow   *Workflow `json:"workflow,omitempty" gorm:"foreignKey:WorkflowID" temporaljson:"workflow,omitzero,omitempty"`

	Status CompositeStatus `json:"status,omitzero" gorm:"type:jsonb" temporaljson:"status,omitzero,omitempty"`

	InstallConfigSyncs      []InstallConfigSync      `json:"install_config_syncs,omitempty" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"install_config_syncs,omitzero,omitempty"`
	InstallCreationApproval *InstallCreationApproval `json:"install_creation_approval,omitempty" gorm:"foreignKey:AppInstallConfigSyncID" temporaljson:"install_creation_approval,omitzero,omitempty"`
}

func (a *AppInstallConfigSync) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name:    indexes.Name(db, &AppInstallConfigSync{}, "org_id"),
			Columns: []string{"org_id"},
		},
		{
			Name:    indexes.Name(db, &AppInstallConfigSync{}, "app_id"),
			Columns: []string{"app_id"},
		},
	}
}

func (a *AppInstallConfigSync) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = domains.NewAppInstallConfigSyncID()
	}

	if a.CreatedByID == "" {
		a.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	if a.OrgID == "" {
		a.OrgID = orgIDFromContext(tx.Statement.Context)
	}

	return nil
}
