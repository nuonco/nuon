package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
)

type AppInput struct {
	ID          string                `gorm:"primarykey;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string                `json:"created_by_id" gorm:"not null;default:null"`
	CreatedBy   Account               `json:"created_by"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
	DeletedAt   soft_delete.DeletedAt `json:"-" gorm:"index:idx_app_input_unique_name,unique"`

	OrgID string `json:"org_id" gorm:"notnull;default null"`
	Org   Org    `faker:"-" json:"-"`

	AppInputConfigID string         `json:"app_input_id" gorm:"notnull; default null;index:idx_app_input_unique_name,unique"`
	AppInputConfig   AppInputConfig `json:"-"`

	AppInputGroup   AppInputGroup `json:"group"`
	AppInputGroupID string        `json:"group_id"`

	Name        string `json:"name" gorm:"not null;default null;index:idx_app_input_unique_name,unique"`
	DisplayName string `json:"display_name"`
	Description string `json:"description" gorm:"not null; default null"`
	Default     string `json:"default"`
	Required    bool   `json:"required"`
	Sensitive   bool   `json:"sensitive"`
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
