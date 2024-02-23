package app

import (
	"time"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

type UserOrg struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string                `json:"created_by_id" gorm:"not null;default:null"`
	CreatedBy   UserToken             `json:"created_by" gorm:"references:Subject"`
	CreatedAt   time.Time             `json:"created_at" gorm:"notnull"`
	UpdatedAt   time.Time             `json:"updated_at" gorm:"notnull"`
	DeletedAt   soft_delete.DeletedAt `gorm:"index" json:"-"`

	// parent relationship
	OrgID string `gorm:"notnull"`
	Org   Org    `gorm:"constraint:OnDelete:CASCADE;" json:"-"`

	UserID string `gorm:"notnull"`
}

func (u *UserOrg) BeforeCreate(tx *gorm.DB) error {
	u.ID = domains.NewUserID()
	if u.CreatedByID == "" {
		u.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	return nil
}
