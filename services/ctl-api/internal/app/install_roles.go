package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
)

type InstallRoles struct {
	ID          string                `gorm:"primarykey;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`
	OrgID       string                `json:"org_id,omitzero" gorm:"notnull;default null" temporaljson:"org_id,omitzero,omitempty"`
	Org         Org                   `json:"-" faker:"-" temporaljson:"org,omitzero,omitempty"`

	InstallID string  `json:"install_id,omitzero" gorm:"notnull;default null" temporaljson:"install_id,omitzero,omitempty"`
	Install   Install `json:"-" temporaljson:"install,omitzero,omitempty"`

	AppRoleConfigID string              `json:"app_role_config_id,omitzero" gorm:"notnull;default null" temporaljson:"app_role_config_id,omitzero,omitempty"`
	AppRoleConfig   AppAWSIAMRoleConfig `json:"app_role_config,omitzero" temporaljson:"app_role_config,omitzero,omitempty"`

	Enabled bool `json:"enabled,omitzero" gorm:"default:false" temporaljson:"enabled,omitzero,omitempty"`

	Provisioned bool `json:"provisioned,omitzero" gorm:"default:false" temporaljson:"provisioned,omitzero,omitempty"`

	// cloud specific role identifier
	RoleID string `json:"role_id,omitzero" temporaljson:"role_id,omitzero,omitempty"`
}

func (i *InstallRoles) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &InstallRoles{}, "idx_org_id_install_id"),
			Columns: []string{
				"org_id",
				"install_id",
			},
		},
	}
}

func (a *InstallRoles) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = domains.NewAppID()
	}

	a.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	a.OrgID = orgIDFromContext(tx.Statement.Context)
	return nil
}
