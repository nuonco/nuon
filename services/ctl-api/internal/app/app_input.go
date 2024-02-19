package app

import (
	"time"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

type AppInput struct {
	ID          string                `gorm:"primarykey;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string                `json:"created_by_id" gorm:"not null;default:null"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
	DeletedAt   soft_delete.DeletedAt `json:"-"`

	OrgID            string `json:"org_id" gorm:"notnull;default null"`
	Org              Org    `faker:"-" json:"-"`
	AppInputConfigID string `json:"app_input_id" gorm:"notnull; default null"`

	Name        string `json:"name" gorm:"not null;default null"`
	DisplayName string `json:"display_name"`
	Description string `json:"description" gorm:"not null; default null"`
	Default     string `json:"default"`
	Required    bool   `json:"required"`
}

func (a *AppInput) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = domains.NewAppID()
	}
	if a.CreatedByID == "" {
		a.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	if a.OrgID == "" {
		a.OrgID = orgIDFromContext(tx.Statement.Context)
	}
	return nil
}
