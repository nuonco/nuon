package app

import (
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/migrations"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/queue/signal"
	signaldb "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/queue/signal/db"
)

type QueueSignal struct {
	ID          string  `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string  `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account `json:"-" temporaljson:"created_by,omitzero,omitempty"`

	CreatedAt time.Time             `json:"created_at,omitzero" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt time.Time             `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgID string `json:"org_id,omitzero" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `json:"-" temporaljson:"org,omitzero,omitempty"`

	QueueID string `json:"queue_id,omitzero" gorm:"type:text;check:owner_id_checker,char_length(id)=26" temporaljson:"queue_id,omitzero,omitempty"`
	Queue   Queue  `json:"queue"`

	OwnerID   string `json:"owner_id,omitzero" gorm:"type:text;check:owner_id_checker,char_length(id)=26" temporaljson:"owner_id,omitzero,omitempty"`
	OwnerType string `json:"owner_type,omitzero" gorm:"type:text;" temporaljson:"owner_type,omitzero,omitempty"`

	Status CompositeStatus     `json:"status"`
	Type   signal.SignalType   `json:"type"`
	Signal signaldb.SignalData `json:"signal" temporaljson:"-"`

	Workflow signaldb.WorkflowRef `json:"workflow"`
}

func (r *QueueSignal) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &QueueSignal{}, "org_id"),
			Columns: []string{
				"org_id",
			},
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
	if r.OrgID == "" {
		r.OrgID = orgIDFromContext(tx.Statement.Context)
	}

	return nil
}
