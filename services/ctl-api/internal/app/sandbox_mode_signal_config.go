package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
)

type SandboxModeSignalConfig struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null"`
	CreatedAt   time.Time             `json:"created_at,omitzero" gorm:"notnull"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" gorm:"notnull"`
	DeletedAt   soft_delete.DeletedAt `json:"-" gorm:"index:idx_sandbox_signal_config_unique,unique"`

	SignalType string `json:"signal_type,omitzero" gorm:"notnull;index:idx_sandbox_signal_config_unique,unique"`

	Enabled bool `json:"enabled" gorm:"default:false"`

	// Separate failure mode fields
	DeadlockSleep time.Duration `json:"deadlock_sleep,omitempty"` // blocks in a real sleep
	WorkflowSleep time.Duration `json:"workflow_sleep,omitempty"` // sleeps using workflow.Sleep
	Panic         bool          `json:"panic"`                    // triggers panic
	Error         string        `json:"error,omitempty"`          // returns error from Execute with this message
	ValidateError string        `json:"validate_error,omitempty"` // returns error from Validate with this message
}

func (s *SandboxModeSignalConfig) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = domains.NewSandboxModeSignalConfigID()
	}
	if s.CreatedByID == "" {
		s.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	return nil
}

func (s *SandboxModeSignalConfig) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name:    indexes.Name(db, &SandboxModeSignalConfig{}, "signal_type"),
			Columns: []string{"signal_type"},
		},
	}
}
