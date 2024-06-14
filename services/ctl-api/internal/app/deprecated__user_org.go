package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
)

type UserOrg_deprecated struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string                `json:"created_by_id" gorm:"not null;default:null"`
	CreatedBy   Account               `json:"created_by"`
	CreatedAt   time.Time             `json:"created_at" gorm:"notnull"`
	UpdatedAt   time.Time             `json:"updated_at" gorm:"notnull"`
	DeletedAt   soft_delete.DeletedAt `gorm:"index" json:"-"`

	// parent relationship
	OrgID string `gorm:"notnull"`
	Org   Org    `gorm:"constraint:OnDelete:CASCADE;" json:"-"`

	UserID string `gorm:"notnull"`
}

func (u *UserOrg_deprecated) TableName() string {
	return "user_orgs"
}

func (u *UserOrg_deprecated) BeforeCreate(tx *gorm.DB) error {
	u.ID = domains.NewUserID()
	if u.CreatedByID == "" {
		u.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	return nil
}
