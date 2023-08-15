package app

import (
	"time"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"gorm.io/gorm"
)

type UserOrg struct {
	ID          string         `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string         `json:"created_by_id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	UserID string
	OrgID  string `gorm:"primaryKey"`
	IsNew  bool   `gorm:"-:all"`
}

func (u *UserOrg) BeforeCreate(tx *gorm.DB) error {
	u.ID = domains.NewUserID()
	return nil
}
