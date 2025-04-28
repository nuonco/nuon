package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
)

type AppInput struct {
	ID          string                `gorm:"primarykey;check:id_checker,char_length(id)=26" json:"id" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" gorm:"index:idx_app_input_unique_name,unique" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgID string `json:"org_id" gorm:"notnull;default null" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `faker:"-" json:"-" temporaljson:"org,omitzero,omitempty"`

	AppInputConfigID string         `json:"app_input_id" gorm:"notnull; default null;index:idx_app_input_unique_name,unique" temporaljson:"app_input_config_id,omitzero,omitempty"`
	AppInputConfig   AppInputConfig `json:"-" temporaljson:"app_input_config,omitzero,omitempty"`

	AppInputGroup   AppInputGroup `json:"group" temporaljson:"app_input_group,omitzero,omitempty"`
	AppInputGroupID string        `json:"group_id" temporaljson:"app_input_group_id,omitzero,omitempty"`

	Name        string `json:"name" gorm:"not null;default null;index:idx_app_input_unique_name,unique" temporaljson:"name,omitzero,omitempty"`
	DisplayName string `json:"display_name" temporaljson:"display_name,omitzero,omitempty"`
	Description string `json:"description" gorm:"not null; default null" temporaljson:"description,omitzero,omitempty"`
	Default     string `json:"default" temporaljson:"default,omitzero,omitempty"`
	Required    bool   `json:"required" temporaljson:"required,omitzero,omitempty"`
	Sensitive   bool   `json:"sensitive" temporaljson:"sensitive,omitzero,omitempty"`
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
