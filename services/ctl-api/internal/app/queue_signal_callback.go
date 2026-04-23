package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
)

type QueueSignalCallback struct {
	ID string `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`

	CreatedAt time.Time             `json:"created_at,omitzero" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt time.Time             `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgID *string `json:"org_id,omitempty" temporaljson:"org_id,omitzero,omitempty"`
	Org   *Org    `json:"-" temporaljson:"org,omitzero,omitempty"`

	QueueSignalID string      `json:"queue_signal_id,omitzero" gorm:"type:text;not null;index:idx_qsc_queue_signal_id" temporaljson:"queue_signal_id,omitzero,omitempty"`
	QueueSignal   QueueSignal `json:"-" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"queue_signal,omitzero,omitempty"`

	Event         string                 `json:"event,omitzero" gorm:"type:text;not null" temporaljson:"event,omitzero,omitempty"`
	UpdateHandler signaldb.UpdateHandler `json:"update_handler" gorm:"type:jsonb;not null" temporaljson:"update_handler,omitzero,omitempty"`
}

func (r *QueueSignalCallback) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &QueueSignalCallback{}, "queue_signal_id"),
			Columns: []string{
				"queue_signal_id",
			},
		},
	}
}

func (r *QueueSignalCallback) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = domains.NewQueueSignalCallbackID()
	}

	if r.OrgID == nil {
		if orgID := orgIDFromContext(tx.Statement.Context); orgID != "" {
			r.OrgID = &orgID
		}
	}

	return nil
}
