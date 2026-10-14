package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
)

type RunbookTriggerType string

const (
	RunbookTriggerTypeManual RunbookTriggerType = "manual"
	RunbookTriggerTypeCron   RunbookTriggerType = "cron"
)

type RunbookTriggerConfig struct {
	ID          string                `json:"id" gorm:"primary_key;check:id_checker,char_length(id)=26" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" gorm:"index:idx_runbook_trigger_config_runbook_config_id_type,unique" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgID           string         `json:"org_id,omitzero" gorm:"notnull" swaggerignore:"true" temporaljson:"org_id,omitzero,omitempty"`
	AppID           string         `json:"app_id,omitzero" gorm:"notnull" temporaljson:"app_id,omitzero,omitempty"`
	AppConfigID     string         `json:"app_config_id,omitzero" gorm:"notnull" temporaljson:"app_config_id,omitzero,omitempty"`
	RunbookConfigID string         `json:"runbook_config_id,omitzero" gorm:"notnull;index:idx_runbook_trigger_config_runbook_config_id_type,unique" temporaljson:"runbook_config_id,omitzero,omitempty"`
	RunbookConfig   *RunbookConfig `json:"-" temporaljson:"runbook_config,omitzero,omitempty"`

	Type         RunbookTriggerType `json:"type" gorm:"notnull;index:idx_runbook_trigger_config_runbook_config_id_type,unique;check:runbook_trigger_type_checker,type IN ('manual','cron')" temporaljson:"type,omitzero,omitempty"`
	CronSchedule string             `json:"cron_schedule,omitempty" temporaljson:"cron_schedule,omitzero,omitempty"`
}

func (r *RunbookTriggerConfig) BeforeCreate(tx *gorm.DB) error {
	r.ID = domains.NewRunbookTriggerConfigID()
	r.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	r.OrgID = orgIDFromContext(tx.Statement.Context)
	return nil
}

func (r *RunbookTriggerConfig) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{Name: indexes.Name(db, r, "org_id"), Columns: []string{"org_id"}},
		{Name: indexes.Name(db, r, "app_config_id"), Columns: []string{"app_config_id"}},
	}
}
