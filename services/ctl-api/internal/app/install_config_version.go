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

type InstallConfigVersion struct {
	ID          string                `gorm:"primarykey;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgID string `json:"org_id,omitzero" gorm:"notnull;default null" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `faker:"-" json:"-" temporaljson:"org,omitzero,omitempty"`

	InstallConfigSyncID string            `json:"install_config_sync_id,omitzero" gorm:"not null" temporaljson:"install_config_sync_id,omitzero,omitempty"`
	InstallConfigSync   InstallConfigSync `faker:"-" json:"-" temporaljson:"install_config_sync,omitzero,omitempty"`

	InstallID   string  `json:"install_id,omitzero" gorm:"not null" temporaljson:"install_id,omitzero,omitempty"`
	Install     Install `faker:"-" json:"-" temporaljson:"install,omitzero,omitempty"`
	InstallName string  `json:"install_name,omitzero" gorm:"not null" temporaljson:"install_name,omitzero,omitempty"`

	FilePath string `json:"file_path,omitempty" temporaljson:"file_path,omitzero,omitempty"`

	Created bool `json:"created,omitzero" gorm:"default:false" temporaljson:"created,omitzero,omitempty"`

	Status CompositeStatus `json:"status,omitzero" gorm:"type:jsonb" temporaljson:"status,omitzero,omitempty"`

	Diff *blobstore.Blob `json:"diff,omitempty" temporaljson:"diff,omitzero,omitempty"`

	Metadata map[string]string `json:"metadata,omitempty" gorm:"type:jsonb;default:null;serializer:json" temporaljson:"metadata,omitzero,omitempty"`
}

func (i *InstallConfigVersion) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name:    indexes.Name(db, &InstallConfigVersion{}, "org_id"),
			Columns: []string{"org_id"},
		},
		{
			Name:    indexes.Name(db, &InstallConfigVersion{}, "install_config_sync_id"),
			Columns: []string{"install_config_sync_id"},
		},
		{
			Name:    indexes.Name(db, &InstallConfigVersion{}, "install_id"),
			Columns: []string{"install_id"},
		},
	}
}

func (i *InstallConfigVersion) BeforeCreate(tx *gorm.DB) error {
	if i.ID == "" {
		i.ID = domains.NewInstallConfigVersionID()
	}

	if i.CreatedByID == "" {
		i.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	if i.OrgID == "" {
		i.OrgID = orgIDFromContext(tx.Statement.Context)
	}

	return nil
}
