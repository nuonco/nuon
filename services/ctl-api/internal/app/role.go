package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
)

type RoleType string

const (
	// user roles
	RoleTypeOrgAdmin RoleType = "org_admin"

	// service account roles
	RoleTypeInstaller RoleType = "installer"
	RoleTypeRunner    RoleType = "runner"
)

type Role struct {
	ID          string `gorm:"primarykey;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string `json:"created_by_id" gorm:"notnull;defaultnull"`
	CreatedBy   Account
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
	DeletedAt   soft_delete.DeletedAt `json:"-" gorm:"notnull;index:idx_role,unique"`

	AccountRoles []AccountRole `gorm:"many2many:account_roles;constraint:OnDelete:CASCADE;" json:"-"`

	OrgID string `json:"org_id" gorm:"notnull;index:idx_role,unique" swaggerignore:"true"`
	Org   Org    `json:"-" faker:"-"`

	RoleType RoleType `json:"role_type" gorm:"notnull;index:idx_role,unique"`

	Policies []Policy `json:"policies"`
}

func (a *Role) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = domains.NewAccountRoleID()
	}
	if a.CreatedByID == "" {
		a.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}

	return nil
}

func (a *Role) AfterQuery(tx *gorm.DB) error {
	return nil
}
