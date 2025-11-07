package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/migrations"
)

type RoleType string

const (
	// user roles
	RoleTypeOrgAdmin RoleType = "org_admin"

	// service account roles
	RoleTypeInstaller       RoleType = "installer"
	RoleTypeRunner          RoleType = "runner"
	RoleTypeHostedInstaller RoleType = "hosted-installer"
)

type Role struct {
	ID          string                `gorm:"primarykey;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"notnull;defaultnull" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `temporaljson:"created_by,omitzero,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	Accounts []Account `gorm:"many2many:account_roles;constraint:OnDelete:CASCADE;" json:"-" temporaljson:"accounts,omitzero,omitempty"`

	// NOTE: not all roles have to belong to an org, this is mainly for historical reasons.
	OrgID generics.NullString `json:"org_id,omitzero" swaggerignore:"true" temporaljson:"org_id,omitzero,omitempty"`
	Org   *Org                `json:"-" faker:"-" temporaljson:"org,omitzero,omitempty"`

	RoleType RoleType `json:"role_type,omitzero" gorm:"defaultnull;notnull" temporaljson:"role_type,omitzero,omitempty"`

	Policies []Policy `json:"policies,omitzero" temporaljson:"policies,omitzero,omitempty"`
}

func (a *Role) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{}
}

func (a *Role) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = domains.NewRoleID()
	}
	if a.CreatedByID == "" {
		a.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}

	return nil
}

func (a *Role) AfterQuery(tx *gorm.DB) error {
	return nil
}
