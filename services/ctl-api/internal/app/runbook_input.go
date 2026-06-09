package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
)

type RunbookInputType string

const (
	RunbookInputTypeString RunbookInputType = "string"
	RunbookInputTypeNumber RunbookInputType = "number"
	RunbookInputTypeBool   RunbookInputType = "bool"
	RunbookInputTypeList   RunbookInputType = "list"
	RunbookInputTypeJSON   RunbookInputType = "json"
)

type RunbookInput struct {
	ID          string                `gorm:"primarykey;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" gorm:"index:idx_runbook_input_unique_name,unique" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgID string `json:"org_id,omitzero" gorm:"notnull;default null" swaggerignore:"true" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `faker:"-" json:"-" temporaljson:"org,omitzero,omitempty"`

	RunbookConfigID string        `json:"runbook_config_id,omitzero" gorm:"notnull;default null;index:idx_runbook_input_unique_name,unique" temporaljson:"runbook_config_id,omitzero,omitempty"`
	RunbookConfig   RunbookConfig `json:"-" temporaljson:"runbook_config,omitzero,omitempty"`

	Idx         int              `json:"idx" gorm:"notnull;default:0" temporaljson:"idx,omitzero,omitempty"`
	Name        string           `json:"name,omitzero" gorm:"not null;default null;index:idx_runbook_input_unique_name,unique" temporaljson:"name,omitzero,omitempty"`
	DisplayName string           `json:"display_name,omitzero" temporaljson:"display_name,omitzero,omitempty"`
	Description string           `json:"description,omitzero" temporaljson:"description,omitzero,omitempty"`
	Default     string           `json:"default,omitzero" temporaljson:"default,omitzero,omitempty"`
	Required    bool             `json:"required,omitzero" temporaljson:"required,omitzero,omitempty"`
	Sensitive   bool             `json:"sensitive,omitzero" temporaljson:"sensitive,omitzero,omitempty"`
	Type        RunbookInputType `json:"type,omitzero" swaggertype:"string" temporaljson:"type,omitzero,omitempty"`
}

func (r *RunbookInput) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &RunbookInput{}, "org_id"),
			Columns: []string{
				"org_id",
			},
		},
		{
			Name: indexes.Name(db, &RunbookInput{}, "runbook_config_id"),
			Columns: []string{
				"runbook_config_id",
			},
		},
	}
}

func (r *RunbookInput) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = domains.NewRunbookInputID()
	}
	if r.CreatedByID == "" {
		r.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	if r.OrgID == "" {
		r.OrgID = orgIDFromContext(tx.Statement.Context)
	}
	return nil
}
