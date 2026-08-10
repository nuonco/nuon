package app

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
)

type InstallCreationApprovalStatus string

const (
	InstallCreationApprovalStatusPending  InstallCreationApprovalStatus = "pending"
	InstallCreationApprovalStatusApproved InstallCreationApprovalStatus = "approved"
	InstallCreationApprovalStatusDenied   InstallCreationApprovalStatus = "denied"
)

type ProposedInstall struct {
	Name     string          `json:"name"`
	FilePath string          `json:"file_path"`
	Config   json.RawMessage `json:"config"`
}

type InstallCreationApproval struct {
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

	AppInstallConfigSyncID string               `json:"app_install_config_sync_id,omitzero" gorm:"not null" temporaljson:"app_install_config_sync_id,omitzero,omitempty"`
	AppInstallConfigSync   AppInstallConfigSync `faker:"-" json:"-" temporaljson:"app_install_config_sync,omitzero,omitempty"`

	ProposedInstalls []ProposedInstall             `json:"proposed_installs,omitempty" gorm:"type:jsonb;serializer:json" temporaljson:"proposed_installs,omitzero,omitempty"`
	Status           InstallCreationApprovalStatus `json:"status,omitzero" gorm:"not null;default:'pending'" temporaljson:"status,omitzero,omitempty"`

	ApprovedAt   *time.Time `json:"approved_at,omitempty" temporaljson:"approved_at,omitzero,omitempty"`
	ApprovedByID *string    `json:"approved_by_id,omitempty" temporaljson:"approved_by_id,omitzero,omitempty"`
}

func (i *InstallCreationApproval) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name:    indexes.Name(db, &InstallCreationApproval{}, "org_id"),
			Columns: []string{"org_id"},
		},
		{
			Name:    indexes.Name(db, &InstallCreationApproval{}, "app_id"),
			Columns: []string{"app_id"},
		},
		{
			Name:    indexes.Name(db, &InstallCreationApproval{}, "app_install_config_sync_id"),
			Columns: []string{"app_install_config_sync_id"},
		},
	}
}

func (i *InstallCreationApproval) BeforeCreate(tx *gorm.DB) error {
	if i.ID == "" {
		i.ID = domains.NewInstallCreationApprovalID()
	}

	if i.CreatedByID == "" {
		i.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	if i.OrgID == "" {
		i.OrgID = orgIDFromContext(tx.Statement.Context)
	}

	return nil
}
