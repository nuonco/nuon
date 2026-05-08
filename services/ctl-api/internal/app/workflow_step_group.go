package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
)

type WorkflowStepGroup struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgID string `json:"org_id,omitzero" gorm:"notnull" swaggerignore:"true" temporaljson:"org_id,omitzero,omitempty"`

	WorkflowID string `json:"workflow_id,omitzero" gorm:"not null" temporaljson:"workflow_id,omitzero,omitempty"`

	GroupIdx int  `json:"group_idx" temporaljson:"group_idx,omitzero,omitempty"`
	Parallel bool `json:"parallel,omitzero" gorm:"default:false" temporaljson:"parallel,omitzero,omitempty"`

	// EagerExecution marks this group for early execution during step generation.
	// When set, the group is returned via the "eager-step-groups" update handler
	// before all groups have been generated, allowing execution to start sooner.
	// This field is not persisted to DB — it is only used during generation.
	EagerExecution bool `json:"eager_execution,omitzero" gorm:"-" temporaljson:"eager_execution,omitzero,omitempty"`

	Status          CompositeStatus `json:"status,omitzero" temporaljson:"status,omitzero,omitempty"`
	Name            string          `json:"name,omitzero" temporaljson:"name,omitzero,omitempty"`
	ResultDirective string          `json:"result_directive,omitzero" gorm:"type:text;default:''" temporaljson:"result_directive,omitzero,omitempty"`

	// Timeout is the execution timeout for this group, derived from its
	// steps' timeouts. For sequential groups: sum of step timeouts. For parallel
	// groups: max of step timeouts. Zero means use default fallback.
	Timeout time.Duration `json:"timeout,omitzero" gorm:"default:0" swaggertype:"string" temporaljson:"timeout,omitzero,omitempty"`

	QueueSignal *QueueSignal `json:"queue_signal,omitempty" gorm:"polymorphic:Owner;" temporaljson:"queue_signal,omitzero,omitempty"`

	Steps []WorkflowStep `json:"steps,omitzero" gorm:"foreignKey:WorkflowStepGroupID;constraint:OnDelete:CASCADE;" temporaljson:"steps,omitzero,omitempty"`
}

func (g *WorkflowStepGroup) TableName() string {
	return "workflow_step_groups"
}

func (g *WorkflowStepGroup) BeforeCreate(tx *gorm.DB) error {
	if g.ID == "" {
		g.ID = domains.NewWorkflowStepGroupID()
	}

	if g.CreatedByID == "" {
		g.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}

	if g.OrgID == "" {
		g.OrgID = orgIDFromContext(tx.Statement.Context)
	}

	return nil
}

func (g *WorkflowStepGroup) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &WorkflowStepGroup{}, "workflow_id_deleted_at"),
			Columns: []string{
				"workflow_id",
				"deleted_at",
			},
		},
		{
			Name: indexes.Name(db, &WorkflowStepGroup{}, "org_id"),
			Columns: []string{
				"org_id",
			},
		},
	}
}
