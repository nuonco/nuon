package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
)

type InstallAppBranchConnection struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `gorm:"index" json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgID string `json:"org_id,omitzero" gorm:"notnull;default null" swaggerignore:"true" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `json:"-" faker:"-" temporaljson:"org,omitzero,omitempty"`

	InstallID string  `json:"install_id,omitzero" gorm:"notnull" temporaljson:"install_id,omitzero,omitempty"`
	Install   Install `json:"-" faker:"-" temporaljson:"install,omitzero,omitempty"`

	AppBranchID string    `json:"app_branch_id,omitzero" gorm:"notnull" temporaljson:"app_branch_id,omitzero,omitempty"`
	AppBranch   AppBranch `json:"app_branch,omitzero" faker:"-" temporaljson:"app_branch,omitzero,omitempty"`

	Active        bool       `json:"active" gorm:"default:true" temporaljson:"active,omitempty"`
	ActivatedAt   time.Time  `json:"activated_at,omitzero" gorm:"notnull" temporaljson:"activated_at,omitzero,omitempty"`
	DeactivatedAt *time.Time `json:"deactivated_at,omitempty" gorm:"default:null" temporaljson:"deactivated_at,omitzero,omitempty"`
}

func (c *InstallAppBranchConnection) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &InstallAppBranchConnection{}, "install_id"),
			Columns: []string{
				"install_id",
			},
		},
		{
			Name: indexes.Name(db, &InstallAppBranchConnection{}, "app_branch_id"),
			Columns: []string{
				"app_branch_id",
			},
		},
		{
			Name: indexes.Name(db, &InstallAppBranchConnection{}, "install_branch_active"),
			Columns: []string{
				"install_id",
				"app_branch_id",
				"active",
			},
		},
	}
}

func (c *InstallAppBranchConnection) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = domains.NewInstallAppBranchConnectionID()
	}
	if c.CreatedByID == "" {
		c.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	if c.OrgID == "" {
		c.OrgID = orgIDFromContext(tx.Statement.Context)
	}
	if c.ActivatedAt.IsZero() {
		c.ActivatedAt = time.Now()
	}
	return nil
}
