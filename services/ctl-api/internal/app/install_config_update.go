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

type InstallAppConfigVersion struct {
	ID          string                `gorm:"primarykey;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgID string `json:"org_id,omitzero" gorm:"notnull;default null" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `faker:"-" json:"-" temporaljson:"org,omitzero,omitempty"`

	AppBranchRunID *string      `json:"app_branch_run_id,omitempty" temporaljson:"app_branch_run_id,omitzero,omitempty"`
	AppBranchRun   AppBranchRun `faker:"-" json:"app_branch_run,omitempty" temporaljson:"app_branch_run,omitzero,omitempty"`

	InstallGroupID *string               `json:"install_group_id,omitempty" temporaljson:"install_group_id,omitzero,omitempty"`
	InstallGroup   AppBranchInstallGroup `faker:"-" json:"-" temporaljson:"install_group,omitzero,omitempty"`

	AppReleaseID *string     `json:"app_release_id,omitempty" gorm:"index" temporaljson:"app_release_id,omitzero,omitempty"`
	AppRelease   *AppRelease `faker:"-" json:"-" gorm:"constraint:OnDelete:RESTRICT;" temporaljson:"app_release,omitzero,omitempty"`

	OperatingModelID *string                `json:"operating_model_id,omitempty" gorm:"index" temporaljson:"operating_model_id,omitzero,omitempty"`
	OperatingModel   *InstallOperatingModel `faker:"-" json:"-" gorm:"constraint:OnDelete:RESTRICT;" temporaljson:"operating_model,omitzero,omitempty"`

	InstallID string  `json:"install_id,omitzero" gorm:"not null" temporaljson:"install_id,omitzero,omitempty"`
	Install   Install `faker:"-" json:"-" temporaljson:"install,omitzero,omitempty"`

	OldAppConfigID string    `json:"old_app_config_id,omitzero" temporaljson:"old_app_config_id,omitzero,omitempty"`
	OldAppConfig   AppConfig `faker:"-" json:"-" temporaljson:"old_app_config,omitzero,omitempty"`

	NewAppConfigID string    `json:"new_app_config_id,omitzero" gorm:"not null" temporaljson:"new_app_config_id,omitzero,omitempty"`
	NewAppConfig   AppConfig `faker:"-" json:"-" temporaljson:"new_app_config,omitzero,omitempty"`

	WorkflowID *string   `json:"workflow_id,omitempty" temporaljson:"workflow_id,omitzero,omitempty"`
	Workflow   *Workflow `json:"workflow,omitempty" temporaljson:"workflow,omitzero,omitempty"`

	Diff *blobstore.Blob `json:"diff,omitempty" temporaljson:"diff,omitzero,omitempty"`

	Metadata map[string]string `json:"metadata,omitempty" gorm:"type:jsonb;default:null;serializer:json" temporaljson:"metadata,omitzero,omitempty"`

	Status CompositeStatus `json:"status,omitzero" gorm:"type:jsonb" temporaljson:"status,omitzero,omitempty"`
}

func (i *InstallAppConfigVersion) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &InstallAppConfigVersion{}, "org_id"),
			Columns: []string{
				"org_id",
			},
		},
		{
			Name: indexes.Name(db, &InstallAppConfigVersion{}, "app_branch_run_id"),
			Columns: []string{
				"app_branch_run_id",
			},
		},
		{
			Name: indexes.Name(db, &InstallAppConfigVersion{}, "install_id"),
			Columns: []string{
				"install_id",
			},
		},
		{
			Name:    indexes.Name(db, &InstallAppConfigVersion{}, "app_release_id"),
			Columns: []string{"app_release_id"},
		},
	}
}

func (i *InstallAppConfigVersion) BeforeCreate(tx *gorm.DB) error {
	if i.ID == "" {
		i.ID = domains.NewInstallAppConfigVersionID()
	}

	if i.CreatedByID == "" {
		i.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	if i.OrgID == "" {
		i.OrgID = orgIDFromContext(tx.Statement.Context)
	}

	return nil
}
