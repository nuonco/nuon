package app

import (
	"database/sql"
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
)

type SandboxModeConfig struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null"`
	CreatedAt   time.Time             `json:"created_at,omitzero" gorm:"notnull"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" gorm:"notnull"`
	DeletedAt   soft_delete.DeletedAt `json:"-"`

	JobType string `json:"job_type,omitzero" gorm:"notnull"`
	Enabled bool   `json:"enabled" gorm:"default:true"`

	// Timing
	Duration      time.Duration `json:"duration,omitzero"`
	SleepDuration time.Duration `json:"sleep_duration,omitempty"`

	// Failure modes (simple toggles)
	ShouldError     bool `json:"should_error"`
	Panic           bool `json:"panic"`
	TriggerShutdown bool `json:"trigger_shutdown"`

	// Template references (keys into the templates package)
	LogTemplate    string `json:"log_template,omitempty"`
	PlanTemplate   string `json:"plan_template,omitempty"`
	OutputTemplate string `json:"output_template,omitempty"`
}

func (s *SandboxModeConfig) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = domains.NewSandboxModeConfigID()
	}
	if s.CreatedByID == "" {
		s.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	return nil
}

func (s *SandboxModeConfig) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name:        indexes.Name(db, &SandboxModeConfig{}, "job_type"),
			Columns:     []string{"job_type", "deleted_at"},
			UniqueValue: sql.NullBool{Bool: true, Valid: true},
		},
	}
}
