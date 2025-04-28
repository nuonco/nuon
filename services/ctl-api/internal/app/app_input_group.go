package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
)

type AppInputGroup struct {
	ID          string                `gorm:"primarykey;check:id_checker,char_length(id)=26" json:"id" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgID string `json:"org_id" gorm:"notnull;default null" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `faker:"-" json:"-" temporaljson:"org,omitzero,omitempty"`

	AppInputConfigID string         `json:"app_input_id" gorm:"notnull; default null" temporaljson:"app_input_config_id,omitzero,omitempty"`
	AppInputConfig   AppInputConfig `json:"-" temporaljson:"app_input_config,omitzero,omitempty"`

	Name        string `json:"name" gorm:"not null;default null" temporaljson:"name,omitzero,omitempty"`
	DisplayName string `json:"display_name" temporaljson:"display_name,omitzero,omitempty"`
	Description string `json:"description" gorm:"not null; default null" temporaljson:"description,omitzero,omitempty"`

	AppInputs []AppInput `json:"app_inputs" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"app_inputs,omitzero,omitempty"`
}

func (a *AppInputGroup) BeforeCreate(tx *gorm.DB) error {
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
