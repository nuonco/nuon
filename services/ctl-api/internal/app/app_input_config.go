package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/migrations"
)

type AppInputConfig struct {
	ID          string                `gorm:"primarykey;check:id_checker,char_length(id)=26" json:"id" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgID       string `json:"org_id" gorm:"notnull;default null" temporaljson:"org_id,omitzero,omitempty"`
	Org         Org    `faker:"-" json:"-" temporaljson:"org,omitzero,omitempty"`
	AppID       string `json:"app_id" temporaljson:"app_id,omitzero,omitempty"`
	AppConfigID string `json:"app_config_id" temporaljson:"app_config_id,omitzero,omitempty"`

	AppInputs      []AppInput      `json:"inputs" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"app_inputs,omitzero,omitempty"`
	AppInputGroups []AppInputGroup `json:"input_groups" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"app_input_groups,omitzero,omitempty"`

	InstallInputs []InstallInputs `json:"install_inputs" gorm:"constraint:OnDelete:CASCADE" temporaljson:"install_inputs,omitzero,omitempty"`
}

func (a *AppInputConfig) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &AppInputConfig{}, "preload"),
			Columns: []string{
				"app_id",
				"deleted_at",
				"created_at DESC",
			},
		},
	}
}

func (a *AppInputConfig) BeforeCreate(tx *gorm.DB) error {
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
