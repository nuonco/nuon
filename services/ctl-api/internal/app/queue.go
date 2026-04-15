package app

import (
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
)

type Queue struct {
	ID          string  `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string  `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account `json:"-" temporaljson:"created_by,omitzero,omitempty"`

	CreatedAt time.Time             `json:"created_at,omitzero" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt time.Time             `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgID *string `json:"org_id,omitempty" temporaljson:"org_id,omitzero,omitempty"`
	Org   *Org    `json:"-" temporaljson:"org,omitzero,omitempty"`

	OwnerID   string `json:"owner_id,omitzero" gorm:"type:text;check:owner_id_checker,char_length(id)=26;" temporaljson:"owner_id,omitzero,omitempty"`
	OwnerType string `json:"owner_type,omitzero" gorm:"type:text;" temporaljson:"owner_type,omitzero,omitempty"`

	Name        string        `json:"name,omitzero" gorm:"default:''" temporaljson:"name,omitzero,omitempty"`
	MaxDepth    int           `json:"max_depth,omitzero"`
	MaxInFlight int           `json:"max_in_flight,omitzero"`
	IdleTimeout int64         `json:"idle_timeout,omitzero" gorm:"default:0" swaggertype:"primitive,integer"`
	Metadata    pgtype.Hstore `json:"metadata,omitzero" gorm:"type:hstore" swaggertype:"object,string" temporaljson:"metadata,omitzero,omitempty"`

	Workflow signaldb.WorkflowRef `json:"workflow"`

	Signals  []QueueSignal  `json:"queue_signal,omitzero" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"signals,omitzero,omitempty"`
	Emitters []QueueEmitter `json:"emitters,omitzero" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"emitters,omitzero,omitempty"`
}

func (r *Queue) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &Queue{}, "org_id"),
			Columns: []string{
				"org_id",
			},
		},
		{
			Name: indexes.Name(db, &Queue{}, "owner"),
			Columns: []string{
				"owner_id",
				"owner_type",
			},
		},
	}
}

func (r *Queue) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = domains.NewQueueID()
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
