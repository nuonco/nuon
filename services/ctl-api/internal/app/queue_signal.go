package app

import (
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
	queuecctx "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
)

type QueueSignal struct {
	ID          string  `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string  `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account `json:"-" temporaljson:"created_by,omitzero,omitempty"`

	CreatedAt time.Time             `json:"created_at,omitzero" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt time.Time             `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgID *string `json:"org_id,omitempty" temporaljson:"org_id,omitzero,omitempty"`
	Org   *Org    `json:"-" temporaljson:"org,omitzero,omitempty"`

	QueueID string `json:"queue_id,omitzero" gorm:"type:text;check:owner_id_checker,char_length(id)=26" temporaljson:"queue_id,omitzero,omitempty"`
	Queue   Queue  `json:"queue"`

	// Optional: if this signal was emitted by an emitter
	EmitterID *string       `json:"emitter_id,omitzero" gorm:"type:text;index:idx_queue_signal_emitter_id" temporaljson:"emitter_id,omitzero,omitempty"`
	Emitter   *QueueEmitter `json:"-" gorm:"constraint:OnDelete:SET NULL;" temporaljson:"emitter,omitzero,omitempty"`

	OwnerID   string `json:"owner_id,omitzero" gorm:"type:text;check:owner_id_checker,char_length(id)=26" temporaljson:"owner_id,omitzero,omitempty"`
	OwnerType string `json:"owner_type,omitzero" gorm:"type:text;" temporaljson:"owner_type,omitzero,omitempty"`

	Status CompositeStatus     `json:"status"`
	Type   signal.SignalType   `json:"type"`
	Signal signaldb.SignalData `json:"signal" temporaljson:"-"`

	Workflow signaldb.WorkflowRef `json:"workflow"`

	Enqueued bool `json:"enqueued" gorm:"default:false;not null" temporaljson:"enqueued,omitzero,omitempty"`

	SignalContext queuecctx.SignalContext `json:"signal_context" gorm:"type:jsonb;default:null" temporaljson:"signal_context,omitzero,omitempty"`

	ExecutionCount int `json:"execution_count" gorm:"default:0;not null" temporaljson:"execution_count,omitzero,omitempty"`
}

func (r *QueueSignal) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &QueueSignal{}, "org_id"),
			Columns: []string{
				"org_id",
			},
		},
		{
			Name: indexes.Name(db, &QueueSignal{}, "enqueued"),
			Columns: []string{
				"enqueued",
				"created_at",
			},
		},
		{
			Name: indexes.Name(db, &QueueSignal{}, "owner_type_owner_id_type_deleted_at"),
			Columns: []string{
				"owner_type",
				"owner_id",
				"type",
				"deleted_at",
				"created_at",
			},
		},
		{
			Name: indexes.Name(db, &QueueSignal{}, "owner_type_owner_id_deleted_at"),
			Columns: []string{
				"owner_type",
				"owner_id",
				"deleted_at",
			},
		},
		{
			Name: indexes.Name(db, &QueueSignal{}, "queue_id"),
			Columns: []string{
				"queue_id",
			},
		},
		{
			Name: indexes.Name(db, &QueueSignal{}, "emitter_id_queue_id_inflight"),
			Columns: []string{
				"emitter_id",
				"queue_id",
			},
			Option: "WHERE deleted_at = 0 AND (status->>'status') IN ('queued','in_progress')",
		},
		{
			Name: indexes.Name(db, &QueueSignal{}, "org_id_type_deleted_at"),
			Columns: []string{
				"org_id",
				"type",
				"deleted_at",
			},
		},
		{
			Name: indexes.Name(db, &QueueSignal{}, "org_id_status_deleted_at"),
			Columns: []string{
				"org_id",
				"(status->>'status')",
				"deleted_at",
			},
		},
		{
			Name: indexes.Name(db, &QueueSignal{}, "created_at_deleted_at"),
			Columns: []string{
				"created_at DESC",
				"deleted_at",
			},
		},
		{
			Name: indexes.Name(db, &QueueSignal{}, "unenqueued_active"),
			Columns: []string{
				"created_at DESC",
				"type",
			},
			Option: "WHERE deleted_at = 0 AND enqueued = false",
		},
	}
}

func (r *QueueSignal) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = domains.NewQueueSignalID()

		// NOTE: we set the ID here, to avoid having to update the object after creation, while still having a
		// 1:1 mapping between id and workflow-id (with the template, of course).
		r.Workflow.ID = fmt.Sprintf(r.Workflow.IDTemplate, r.ID)
	}

	if r.CreatedByID == "" {
		r.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	if r.OrgID == nil {
		if orgID := orgIDFromContext(tx.Statement.Context); orgID != "" {
			r.OrgID = &orgID
		}
	}

	return nil
}
