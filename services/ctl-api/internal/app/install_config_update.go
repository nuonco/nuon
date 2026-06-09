package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
)

// InstallConfigUpdate tracks a per-install config transition during an app branch run.
// It links an AppBranchRun to the install workflow that diffs and deploys changes.
type InstallConfigUpdate struct {
	ID          string                `gorm:"primarykey;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgID string `json:"org_id,omitzero" gorm:"notnull;default null" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `faker:"-" json:"-" temporaljson:"org,omitzero,omitempty"`

	AppBranchRunID string       `json:"app_branch_run_id,omitempty" temporaljson:"app_branch_run_id,omitzero,omitempty"`
	AppBranchRun   AppBranchRun `faker:"-" json:"-" temporaljson:"app_branch_run,omitzero,omitempty"`

	InstallGroupID string                `json:"install_group_id,omitempty" temporaljson:"install_group_id,omitzero,omitempty"`
	InstallGroup   AppBranchInstallGroup `faker:"-" json:"-" temporaljson:"install_group,omitzero,omitempty"`

	InstallID string  `json:"install_id,omitzero" gorm:"not null" temporaljson:"install_id,omitzero,omitempty"`
	Install   Install `faker:"-" json:"-" temporaljson:"install,omitzero,omitempty"`

	OldAppConfigID string    `json:"old_app_config_id,omitzero" temporaljson:"old_app_config_id,omitzero,omitempty"`
	OldAppConfig   AppConfig `faker:"-" json:"-" temporaljson:"old_app_config,omitzero,omitempty"`

	NewAppConfigID string    `json:"new_app_config_id,omitzero" gorm:"not null" temporaljson:"new_app_config_id,omitzero,omitempty"`
	NewAppConfig   AppConfig `faker:"-" json:"-" temporaljson:"new_app_config,omitzero,omitempty"`

	// WorkflowID links to the install workflow that performs the actual diff and deploy.
	WorkflowID *string   `json:"workflow_id,omitempty" temporaljson:"workflow_id,omitzero,omitempty"`
	Workflow   *Workflow `json:"workflow,omitempty" temporaljson:"workflow,omitzero,omitempty"`

	// Diff stores the serialized config diff result.
	Diff *blobstore.Blob `json:"diff,omitempty" temporaljson:"diff,omitzero,omitempty"`

	Status CompositeStatus `json:"status,omitzero" gorm:"type:jsonb" temporaljson:"status,omitzero,omitempty"`
}

func (i *InstallConfigUpdate) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &InstallConfigUpdate{}, "org_id"),
			Columns: []string{
				"org_id",
			},
		},
		{
			Name: indexes.Name(db, &InstallConfigUpdate{}, "app_branch_run_id"),
			Columns: []string{
				"app_branch_run_id",
			},
		},
		{
			Name: indexes.Name(db, &InstallConfigUpdate{}, "install_id"),
			Columns: []string{
				"install_id",
			},
		},
	}
}

func (i *InstallConfigUpdate) BeforeCreate(tx *gorm.DB) error {
	if i.ID == "" {
		i.ID = domains.NewInstallConfigUpdateID()
	}

	if i.CreatedByID == "" {
		i.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	if i.OrgID == "" {
		i.OrgID = orgIDFromContext(tx.Statement.Context)
	}

	return nil
}
