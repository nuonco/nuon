package app

import (
	"time"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

type AppSecret struct {
	ID          string                `gorm:"primarykey;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string                `json:"created_by_id" gorm:"not null;default:null"`
	CreatedBy   UserToken             `json:"created_by" gorm:"references:Subject"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
	DeletedAt   soft_delete.DeletedAt `json:"-" gorm:"index:idx_app_secret_name,unique"`

	OrgID string `json:"org_id" gorm:"notnull;default null"`
	Org   Org    `faker:"-" json:"-"`

	AppID string `json:"app_id" gorm:"not null;default null;index:idx_app_secret_name,unique"`
	App   App    `json:"-" faker:"-"`

	Name  string `json:"name" gorm:"not null;default null;index:idx_app_secret_name,unique"`
	Value string `json:"-" gorm:"not null;default null"`

	// after query fields
	Length int `json:"length" gorm:"-"`
}

func (a *AppSecret) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = domains.NewAppSecretID()
	}
	if a.CreatedByID == "" {
		a.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	if a.OrgID == "" {
		a.OrgID = orgIDFromContext(tx.Statement.Context)
	}

	return nil
}

func (a *AppSecret) AfterQuery(tx *gorm.DB) error {
	a.Length = len(a.Value)
	return nil
}
