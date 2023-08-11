package app

import (
	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"gorm.io/gorm"
)

type UserOrg struct {
	Model

	UserID string
	OrgID  string `gorm:"primaryKey"`
	IsNew  bool   `gorm:"-:all"`
}

func (u *UserOrg) BeforeCreate(tx *gorm.DB) error {
	u.ID = domains.NewUserID()
	return nil
}
