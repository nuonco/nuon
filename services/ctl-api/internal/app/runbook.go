package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
)

type RunbookStatus string

const (
	RunbookStatusActive       RunbookStatus = "active"
	RunbookStatusError        RunbookStatus = "error"
	RunbookStatusDeleteQueued RunbookStatus = "delete_queued"
)

type Runbook struct {
	ID string `json:"id" gorm:"primary_key;check:id_checker,char_length(id)=26" temporaljson:"id,omitzero,omitempty"`

	Status            RunbookStatus   `json:"status,omitzero" gorm:"notnull;default:'active'" swaggertype:"string" temporaljson:"status,omitzero,omitempty"`
	StatusDescription string          `json:"status_description,omitzero" temporaljson:"status_description,omitzero,omitempty"`
	StatusV2          CompositeStatus `json:"status_v2,omitzero" gorm:"type:jsonb" temporaljson:"status_v2,omitzero,omitempty"`

	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" gorm:"index:idx_runbook_app_id_name,unique" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgID string `json:"org_id,omitzero" gorm:"notnull" swaggerignore:"true" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `json:"-" faker:"-" temporaljson:"org,omitzero,omitempty"`

	App   App    `json:"-" swaggerignore:"true" temporaljson:"app,omitzero,omitempty"`
	AppID string `json:"app_id,omitzero" gorm:"index:idx_runbook_app_id_name,unique" faker:"-" temporaljson:"app_id,omitzero,omitempty"`

	Configs     []RunbookConfig `json:"configs" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"configs,omitzero,omitempty"`
	ConfigCount int             `json:"config_count" gorm:"->;-:migration" temporaljson:"config_count,omitzero,omitempty"`

	Name        string `json:"name,omitzero" gorm:"index:idx_runbook_app_id_name,unique" temporaljson:"name,omitzero,omitempty"`
	Description string `json:"description,omitzero" temporaljson:"description,omitzero,omitempty"`
	labels.Labeled
}

func (r *Runbook) BeforeCreate(tx *gorm.DB) error {
	r.ID = domains.NewRunbookID()
	r.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	r.OrgID = orgIDFromContext(tx.Statement.Context)
	return nil
}

func (r *Runbook) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &Runbook{}, "org_id"),
			Columns: []string{
				"org_id",
			},
		},
	}
}
