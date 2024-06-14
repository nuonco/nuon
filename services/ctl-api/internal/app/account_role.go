package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
)

// AccountRole is a many2many table used by gorm under the hood
type AccountRole struct {
	ID string `gorm:"primarykey;check:id_checker,char_length(id)=26" json:"id"`

	CreatedAt time.Time             `json:"created_at" gorm:"notnull"`
	UpdatedAt time.Time             `json:"updated_at" gorm:"notnull"`
	DeletedAt soft_delete.DeletedAt `json:"-" gorm:"index:idx_account_role:unique"`

	OrgID string `json:"org_id" gorm:"notnull" swaggerignore:"true"`
	Org   Org    `json:"-" faker:"-"`

	RoleID string `gorm:"primary_key;index:idx_account_role:unique"`
	Role   Role

	AccountID string `gorm:"primary_key;index:idx_account_role:unique"`
	Account   Account
}

func (c *AccountRole) BeforeSave(tx *gorm.DB) error {
	c.ID = domains.NewAccountID()

	if c.OrgID == "" {
		c.OrgID = orgIDFromContext(tx.Statement.Context)
	}

	return nil
}
