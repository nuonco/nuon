package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
)

// InstallScheduledReconcileStatus is the lifecycle status of a scheduled
// reconcile. The pinned app config is read at fire time, not at schedule
// time, so the row never stores a config ID.
type InstallScheduledReconcileStatus string

const (
	InstallScheduledReconcileStatusPending   InstallScheduledReconcileStatus = "pending"
	InstallScheduledReconcileStatusFired     InstallScheduledReconcileStatus = "fired"
	InstallScheduledReconcileStatusNoChanges InstallScheduledReconcileStatus = "no-changes"
	InstallScheduledReconcileStatusCancelled InstallScheduledReconcileStatus = "cancelled"
	InstallScheduledReconcileStatusFailed    InstallScheduledReconcileStatus = "failed"
)

// InstallScheduledReconcile is a user-scheduled reconcile-app-version request
// for an install. An emitter fires a scheduledreconcile signal at
// ScheduledAt, which loads this row and dispatches the reconcile.
type InstallScheduledReconcile struct {
	ID          string                `gorm:"primarykey;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgID string `json:"org_id,omitzero" gorm:"notnull;default null" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `faker:"-" json:"-" temporaljson:"org,omitzero,omitempty"`

	InstallID string  `json:"install_id,omitzero" gorm:"not null" temporaljson:"install_id,omitzero,omitempty"`
	Install   Install `faker:"-" json:"-" temporaljson:"install,omitzero,omitempty"`

	ScheduledAt time.Time `json:"scheduled_at,omitzero" gorm:"not null" temporaljson:"scheduled_at,omitzero,omitempty"`
	PlanOnly    bool      `json:"plan_only" temporaljson:"plan_only,omitzero,omitempty"`

	// EmitterID is the fire-once queue emitter that dispatches this schedule.
	EmitterID string `json:"emitter_id,omitzero" gorm:"default:null" temporaljson:"emitter_id,omitzero,omitempty"`

	Status       InstallScheduledReconcileStatus `json:"status,omitzero" gorm:"not null;default:pending" temporaljson:"status,omitzero,omitempty"`
	StatusReason string                          `json:"status_reason,omitempty" gorm:"default:null" temporaljson:"status_reason,omitzero,omitempty"`

	// WorkflowID is set once the reconcile workflow has been created (status fired).
	WorkflowID *string `json:"workflow_id,omitempty" temporaljson:"workflow_id,omitzero,omitempty"`
}

func (i *InstallScheduledReconcile) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &InstallScheduledReconcile{}, "org_id"),
			Columns: []string{
				"org_id",
			},
		},
		{
			Name: indexes.Name(db, &InstallScheduledReconcile{}, "install_id"),
			Columns: []string{
				"install_id",
			},
		},
	}
}

func (i *InstallScheduledReconcile) BeforeCreate(tx *gorm.DB) error {
	if i.ID == "" {
		i.ID = domains.NewInstallScheduledReconcileID()
	}

	if i.CreatedByID == "" {
		i.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	if i.OrgID == "" {
		i.OrgID = orgIDFromContext(tx.Statement.Context)
	}
	if i.Status == "" {
		i.Status = InstallScheduledReconcileStatusPending
	}

	return nil
}
