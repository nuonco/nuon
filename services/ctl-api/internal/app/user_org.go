package app

import (
	"time"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"gorm.io/gorm"
)

type UserOrg struct {
	ID          string         `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string         `json:"created_by_id" gorm:"notnull"`
	CreatedAt   time.Time      `json:"created_at" gorm:"notnull"`
	UpdatedAt   time.Time      `json:"updated_at" gorm:"notnull"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// parent relationship
	OrgID string `gorm:"notnull"`
	Org   Org    `gorm:"constraint:OnDelete:CASCADE;"`

	UserID string `gorm:"notnull"`
}

func (u *UserOrg) BeforeCreate(tx *gorm.DB) error {
	u.ID = domains.NewUserID()
	u.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	return nil
}
