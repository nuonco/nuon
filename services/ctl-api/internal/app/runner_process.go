package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
)

type RunnerProcessStatus string

const (
	RunnerProcessStatusActive       RunnerProcessStatus = "active"
	RunnerProcessStatusShuttingDown RunnerProcessStatus = "shutting-down"
	RunnerProcessStatusShutDown     RunnerProcessStatus = "shut-down"
	RunnerProcessStatusError        RunnerProcessStatus = "error"
	RunnerProcessStatusUnknown      RunnerProcessStatus = "unknown"
)


type RunnerProcess struct {
	ID          string  `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero"`
	CreatedByID string  `gorm:"not null;default:null" json:"created_by_id,omitzero"`
	CreatedBy   Account `json:"-"`

	CreatedAt time.Time             `json:"created_at,omitzero" gorm:"notnull"`
	UpdatedAt time.Time             `json:"updated_at,omitzero" gorm:"notnull"`
	DeletedAt soft_delete.DeletedAt `json:"-"`

	OrgID string `json:"org_id,omitzero" gorm:"index"`
	Org   Org    `json:"-"`

	RunnerID string `json:"runner_id,omitzero" gorm:"index"`
	Runner   Runner `json:"-"`

	Type RunnerProcessType `json:"type,omitzero" gorm:"not null"`

	Status            RunnerProcessStatus `json:"status,omitzero" gorm:"not null;default:'active'"`
	StatusDescription string              `json:"status_description,omitzero"`
	CompositeStatus   CompositeStatus     `json:"composite_status,omitzero" gorm:"type:jsonb"`

	LogStreamID *string    `json:"log_stream_id,omitempty"`
	LogStream   *LogStream `json:"-"`

	Version   string     `json:"version,omitzero"`
	StartedAt *time.Time `json:"started_at,omitempty"`

	Shutdowns []RunnerProcessShutdown `json:"shutdowns,omitempty" gorm:"constraint:OnDelete:CASCADE;"`
}

func (r *RunnerProcess) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = domains.NewRunnerProcessID()
	}

	if r.CreatedByID == "" {
		r.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}

	if r.OrgID == "" {
		r.OrgID = orgIDFromContext(tx.Statement.Context)
	}

	return nil
}

func (r *RunnerProcess) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name:    indexes.Name(db, &RunnerProcess{}, "runner_type_created"),
			Columns: []string{"runner_id", "type", "deleted_at", "created_at DESC"},
		},
		{
			Name:    indexes.Name(db, &RunnerProcess{}, "org_id"),
			Columns: []string{"org_id"},
		},
	}
}
