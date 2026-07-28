package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
)

type InstallConfigSync struct {
	ID          string                `gorm:"primarykey;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgID string `json:"org_id,omitzero" gorm:"notnull;default null" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `faker:"-" json:"-" temporaljson:"org,omitzero,omitempty"`

	InstallID string  `json:"install_id,omitzero" gorm:"default:null" temporaljson:"install_id,omitzero,omitempty"`
	Install   Install `faker:"-" json:"-" temporaljson:"install,omitzero,omitempty"`

	AppBranchID string    `json:"app_branch_id,omitzero" gorm:"not null" temporaljson:"app_branch_id,omitzero,omitempty"`
	AppBranch   AppBranch `faker:"-" json:"-" temporaljson:"app_branch,omitzero,omitempty"`

	AppBranchConfigID string          `json:"app_branch_config_id,omitzero" gorm:"not null" temporaljson:"app_branch_config_id,omitzero,omitempty"`
	AppBranchConfig   AppBranchConfig `faker:"-" json:"-" temporaljson:"app_branch_config,omitzero,omitempty"`

	AppBranchRunID *string      `json:"app_branch_run_id,omitempty" temporaljson:"app_branch_run_id,omitzero,omitempty"`
	AppBranchRun   AppBranchRun `faker:"-" json:"-" temporaljson:"app_branch_run,omitzero,omitempty"`

	VCSConnectionCommitID *string              `json:"vcs_connection_commit_id,omitempty" swaggerignore:"true" temporaljson:"vcs_connection_commit_id,omitzero,omitempty"`
	VCSConnectionCommit   *VCSConnectionCommit `json:"vcs_connection_commit,omitempty" temporaljson:"vcs_connection_commit,omitzero,omitempty"`

	TriggeredBy string `json:"triggered_by,omitzero" gorm:"not null;default:'manual'" temporaljson:"triggered_by,omitzero,omitempty"`

	Status CompositeStatus `json:"status,omitzero" gorm:"type:jsonb" temporaljson:"status,omitzero,omitempty"`

	Metadata map[string]string `json:"metadata,omitempty" gorm:"type:jsonb;default:null;serializer:json" temporaljson:"metadata,omitzero,omitempty"`

	Versions []InstallConfigVersion `json:"versions,omitempty" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"versions,omitzero,omitempty"`
}

func (i *InstallConfigSync) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name:    indexes.Name(db, &InstallConfigSync{}, "org_id"),
			Columns: []string{"org_id"},
		},
		{
			Name:    indexes.Name(db, &InstallConfigSync{}, "install_id"),
			Columns: []string{"install_id"},
		},
		{
			Name:    indexes.Name(db, &InstallConfigSync{}, "app_branch_id"),
			Columns: []string{"app_branch_id"},
		},
		{
			Name:    indexes.Name(db, &InstallConfigSync{}, "app_branch_run_id"),
			Columns: []string{"app_branch_run_id"},
		},
	}
}

func (i *InstallConfigSync) BeforeCreate(tx *gorm.DB) error {
	if i.ID == "" {
		i.ID = domains.NewInstallConfigSyncID()
	}

	if i.CreatedByID == "" {
		i.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	if i.OrgID == "" {
		i.OrgID = orgIDFromContext(tx.Statement.Context)
	}

	return nil
}
