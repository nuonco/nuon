package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/views"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/viewsql"
)

type RunbookConfig struct {
	ID          string                `json:"id,omitzero" gorm:"primary_key;check:id_checker,char_length(id)=26" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" gorm:"index:idx_runbook_config_runbook_id_app_config_id,unique" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgID string `json:"org_id,omitzero" gorm:"notnull" swaggerignore:"true" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `json:"-" faker:"-" temporaljson:"org,omitzero,omitempty"`

	App   App    `json:"-" swaggerignore:"true" temporaljson:"app,omitzero,omitempty"`
	AppID string `json:"app_id,omitzero" gorm:"notnull" temporaljson:"app_id,omitzero,omitempty"`

	AppConfigID string    `json:"app_config_id,omitzero" gorm:"index:idx_runbook_config_runbook_id_app_config_id,unique" temporaljson:"app_config_id,omitzero,omitempty"`
	AppConfig   AppConfig `json:"-" temporaljson:"app_config,omitzero,omitempty"`

	RunbookID string  `json:"runbook_id,omitzero" gorm:"index:idx_runbook_config_runbook_id_app_config_id,unique" temporaljson:"runbook_id,omitzero,omitempty"`
	Runbook   Runbook `json:"-" temporaljson:"runbook,omitzero,omitempty"`

	Readme string              `json:"readme,omitzero" temporaljson:"readme,omitzero,omitempty"`
	Steps  []RunbookStepConfig `json:"steps,omitzero" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"steps,omitzero,omitempty"`
	Inputs []RunbookInput      `json:"inputs,omitzero" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"inputs,omitzero,omitempty"`
}

func (r *RunbookConfig) BeforeCreate(tx *gorm.DB) error {
	r.ID = domains.NewRunbookConfigID()
	r.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	r.OrgID = orgIDFromContext(tx.Statement.Context)
	return nil
}

func (r *RunbookConfig) Views(db *gorm.DB) []migrations.View {
	return []migrations.View{
		{
			Name:          views.CustomViewName(db, &RunbookConfig{}, "latest_view_v1"),
			SQL:           viewsql.RunbookConfigsLatestViewV1,
			AlwaysReapply: true,
		},
	}
}

func (r *RunbookConfig) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &RunbookConfig{}, "org_id"),
			Columns: []string{
				"org_id",
			},
		},
	}
}
