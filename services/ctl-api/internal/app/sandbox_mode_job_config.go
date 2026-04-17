package app

import (
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
	DeletedAt   soft_delete.DeletedAt `json:"-" gorm:"index:idx_sandbox_config_unique,unique"`

	JobType string `json:"job_type,omitzero" gorm:"notnull;index:idx_sandbox_config_unique,unique"`

	Enabled bool `json:"enabled" gorm:"default:true"`

	// Behavior
	Preset       string        `json:"preset,omitzero"`
	Duration     time.Duration `json:"duration,omitzero"`
	FaultRate    float64       `json:"fault_rate"`
	ErrorMessage string        `json:"error_message,omitempty"`
	FailAtStep   string        `json:"fail_at_step,omitempty"`

	// Sleep/timeout/shutdown controls
	SleepDuration   time.Duration `json:"sleep_duration,omitempty"`
	Timeout         time.Duration `json:"timeout,omitempty"`
	TriggerShutdown bool          `json:"trigger_shutdown,omitempty"`

	// Log data — stored in DB, editable from dashboard
	LogLines []byte `json:"log_lines,omitempty" gorm:"type:jsonb" swaggertype:"string"`

	// Plan contents (for plan-type jobs)
	PlanMachineContents string `json:"machine_contents,omitempty" gorm:"type:text"`
	PlanDisplayContents string `json:"display_contents,omitempty" gorm:"type:text"`

	// Custom outputs
	Outputs []byte `json:"outputs,omitempty" gorm:"type:jsonb" swaggertype:"string"`
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
			Name:    indexes.Name(db, &SandboxModeConfig{}, "runner_id"),
			Columns: []string{"runner_id"},
		},
		{
			Name:    indexes.Name(db, &SandboxModeConfig{}, "runner_id_job_type"),
			Columns: []string{"runner_id", "job_type"},
		},
	}
}
