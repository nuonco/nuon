package app

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
)

type RunnerProcessShutdownStatus string

const (
	RunnerProcessShutdownStatusRequested  RunnerProcessShutdownStatus = "requested"
	RunnerProcessShutdownStatusInProgress RunnerProcessShutdownStatus = "in-progress"
	RunnerProcessShutdownStatusCompleted  RunnerProcessShutdownStatus = "completed"
	RunnerProcessShutdownStatusFailed     RunnerProcessShutdownStatus = "failed"
	RunnerProcessShutdownStatusCancelled  RunnerProcessShutdownStatus = "cancelled"
)

type RunnerProcessShutdownType string

const (
	RunnerProcessShutdownTypeGraceful RunnerProcessShutdownType = "graceful"
	RunnerProcessShutdownTypeForce    RunnerProcessShutdownType = "force"
	RunnerProcessShutdownTypeRestart  RunnerProcessShutdownType = "restart"
)

type RunnerProcessShutdown struct {
	ID          string  `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero"`
	CreatedByID string  `gorm:"not null;default:null" json:"created_by_id,omitzero"`
	CreatedBy   Account `json:"-"`

	CreatedAt time.Time             `json:"created_at,omitzero" gorm:"notnull"`
	UpdatedAt time.Time             `json:"updated_at,omitzero" gorm:"notnull"`
	DeletedAt soft_delete.DeletedAt `json:"-"`

	OrgID string `json:"org_id,omitzero" gorm:"index"`
	Org   Org    `json:"-"`

	RunnerProcessID string `json:"runner_process_id,omitzero" gorm:"index;not null"`

	Type            RunnerProcessShutdownType `json:"type,omitzero" gorm:"not null"`
	CompositeStatus CompositeStatus           `json:"composite_status,omitzero" gorm:"type:jsonb"`

	// Status and StatusDescription are computed from CompositeStatus via AfterQuery.
	Status            RunnerProcessShutdownStatus `json:"status,omitzero" gorm:"-"`
	StatusDescription string                      `json:"status_description,omitzero" gorm:"-"`

	Metadata pgtype.Hstore `json:"metadata,omitempty" gorm:"type:hstore"`
}

func (r *RunnerProcessShutdown) AfterQuery(tx *gorm.DB) error {
	if r.CompositeStatus.Status != "" {
		r.Status = RunnerProcessShutdownStatus(r.CompositeStatus.Status)
		r.StatusDescription = r.CompositeStatus.StatusHumanDescription
	}
	return nil
}

func (r *RunnerProcessShutdown) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = domains.NewRunnerProcessShutdownID()
	}

	if r.CreatedByID == "" {
		r.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}

	if r.OrgID == "" {
		r.OrgID = orgIDFromContext(tx.Statement.Context)
	}

	return nil
}

func (r *RunnerProcessShutdown) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name:    indexes.Name(db, &RunnerProcessShutdown{}, "process_id"),
			Columns: []string{"runner_process_id", "deleted_at", "created_at DESC"},
		},
		{
			Name:    indexes.Name(db, &RunnerProcessShutdown{}, "org_id"),
			Columns: []string{"org_id"},
		},
	}
}
