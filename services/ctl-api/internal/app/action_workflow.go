package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/migrations"
)

type ActionWorkflowStatus string

const (
	ActionWorkflowStatusActive ActionWorkflowStatus = "active"
	// error state
	ActionWorkflowStatusError ActionWorkflowStatus = "error"
	// queued for deletion
	ActionWorkflowStatusDeleteQueued ActionWorkflowStatus = "delete_queued"
)

type ActionWorkflow struct {
	ID string `json:"id" gorm:"primary_key;check:id_checker,char_length(id)=26" temporaljson:"id,omitzero,omitempty"`
	// TODO: change to default null after migration & after initial pr
	Status            ActionWorkflowStatus  `json:"status,omitzero" gorm:"notnull;default:'active'" swaggertype:"string" temporaljson:"status,omitzero,omitempty"`
	StatusDescription string                `json:"status_description,omitzero" temporaljson:"status_description,omitzero,omitempty"`
	CreatedByID       string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy         Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt         time.Time             `json:"created_at,omitzero" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt         time.Time             `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt         soft_delete.DeletedAt `json:"-" gorm:"index:idx_action_workflow_app_id_name,unique" temporaljson:"deleted_at,omitzero,omitempty"`

	// used for RLS
	OrgID string `json:"org_id,omitzero" gorm:"notnull" swaggerignore:"true" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `json:"-" faker:"-" temporaljson:"org,omitzero,omitempty"`

	App   App    `json:"-" swaggerignore:"true" temporaljson:"app,omitzero,omitempty"`
	AppID string `json:"app_id,omitzero" gorm:"index:idx_action_workflow_app_id_name,unique" faker:"-" temporaljson:"app_id,omitzero,omitempty"`

	Configs     []ActionWorkflowConfig `json:"configs" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"configs,omitzero,omitempty"`
	ConfigCount int                    `json:"config_count" gorm:"->;-:migration" temporaljson:"config_count,omitzero,omitempty"`

	// metadata
	Name string `json:"name,omitzero" gorm:"index:idx_action_workflow_app_id_name,unique" temporaljson:"name,omitzero,omitempty"`
}

func (a *ActionWorkflow) BeforeCreate(tx *gorm.DB) error {
	a.ID = domains.NewActionWorkflowID()
	a.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	a.OrgID = orgIDFromContext(tx.Statement.Context)
	return nil
}

func (a *ActionWorkflow) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &ActionWorkflow{}, "org_id"),
			Columns: []string{
				"org_id",
			},
		},
	}
}
