package app

import (
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
)

const (
	defaultQueueEmitterWorkflowIDTemplate string = "queue-emitter-%s"
)

type QueueEmitterMode string

const (
	// QueueEmitterModeCron emits signals on a recurring cron schedule
	QueueEmitterModeCron QueueEmitterMode = "cron"
	// QueueEmitterModeScheduled emits a signal once at a scheduled time, then stops
	QueueEmitterModeScheduled QueueEmitterMode = "scheduled"
)

type QueueEmitter struct {
	ID          string  `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string  `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account `json:"-" temporaljson:"created_by,omitzero,omitempty"`

	CreatedAt time.Time             `json:"created_at,omitzero" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt time.Time             `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgID string `json:"org_id,omitzero" gorm:"notnull" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `json:"-" temporaljson:"org,omitzero,omitempty"`

	// Many-to-one: each emitter belongs to exactly one queue
	QueueID string `json:"queue_id,omitzero" gorm:"type:text;not null;index:idx_queue_emitter_queue_id" temporaljson:"queue_id,omitzero,omitempty"`
	Queue   Queue  `json:"-" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"queue,omitzero,omitempty"`

	// Emitter identity
	Name        string `json:"name,omitzero" gorm:"type:text;not null" temporaljson:"name,omitzero,omitempty"`
	Description string `json:"description,omitzero" gorm:"type:text" temporaljson:"description,omitzero,omitempty"`

	// Emitter mode: "cron" for recurring, "scheduled" for one-shot
	Mode QueueEmitterMode `json:"mode,omitzero" gorm:"type:text;not null;default:'cron'" temporaljson:"mode,omitzero,omitempty"`

	// Schedule configuration
	// For cron mode: cron expression (e.g., "0 * * * *")
	CronSchedule string `json:"cron_schedule,omitzero" gorm:"type:text" temporaljson:"cron_schedule,omitzero,omitempty"`
	// For scheduled mode: the time to fire the signal
	ScheduledAt *time.Time `json:"scheduled_at,omitzero" temporaljson:"scheduled_at,omitzero,omitempty"`
	// For scheduled mode: whether the signal has been fired
	Fired bool `json:"fired,omitzero" gorm:"default:false" temporaljson:"fired,omitzero,omitempty"`

	// Signal template - the signal to emit on each tick
	SignalType     signal.SignalType   `json:"signal_type,omitzero" gorm:"type:text;not null" temporaljson:"signal_type,omitzero,omitempty"`
	SignalTemplate signaldb.SignalData `json:"signal_template,omitzero" temporaljson:"-"`

	// Runtime state using shared CompositeStatus
	Status        CompositeStatus `json:"status" temporaljson:"status,omitzero,omitempty"`
	LastEmittedAt *time.Time      `json:"last_emitted_at,omitzero" temporaljson:"last_emitted_at,omitzero,omitempty"`
	NextEmitAt    *time.Time      `json:"next_emit_at,omitzero" temporaljson:"next_emit_at,omitzero,omitempty"`
	EmitCount     int64           `json:"emit_count,omitzero" gorm:"default:0" temporaljson:"emit_count,omitzero,omitempty"`

	// Workflow reference for the emitter's cron workflow
	Workflow signaldb.WorkflowRef `json:"workflow" temporaljson:"workflow,omitzero,omitempty"`

	// Signals emitted by this emitter
	Signals []QueueSignal `json:"-" gorm:"foreignKey:EmitterID" temporaljson:"signals,omitzero,omitempty"`
}

func (r *QueueEmitter) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &QueueEmitter{}, "org_id"),
			Columns: []string{
				"org_id",
			},
		},
		{
			Name: indexes.Name(db, &QueueEmitter{}, "queue_id"),
			Columns: []string{
				"queue_id",
			},
		},
	}
}

func (r *QueueEmitter) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = domains.NewQueueEmitterID()
		r.Workflow.ID = fmt.Sprintf(r.Workflow.IDTemplate, r.ID)
		if r.Workflow.IDTemplate == "" {
			r.Workflow.IDTemplate = defaultQueueEmitterWorkflowIDTemplate
			r.Workflow.ID = fmt.Sprintf(defaultQueueEmitterWorkflowIDTemplate, r.ID)
		}
	}

	if r.CreatedByID == "" {
		r.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	if r.OrgID == "" {
		r.OrgID = orgIDFromContext(tx.Statement.Context)
	}

	return nil
}
