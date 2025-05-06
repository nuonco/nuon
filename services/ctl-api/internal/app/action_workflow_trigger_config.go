package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/pkg/shortid/domains"
)

type ActionWorkflowTriggerType string

const (
	// this is for manual debugging/triggering in the ui
	ActionWorkflowTriggerTypeManual ActionWorkflowTriggerType = "manual"

	// run on a hook
	ActionWorkflowTriggerTypeCron ActionWorkflowTriggerType = "cron"

	// can add workflow triggers for different types of events
	ActionWorkflowTriggerTypePreSandboxRun  ActionWorkflowTriggerType = "pre-sandbox-run"
	ActionWorkflowTriggerTypePostSandboxRun ActionWorkflowTriggerType = "post-sandbox-run"

	// triggers that run on a specific component deploy
	ActionWorkflowTriggerTypePreDeployComponent  ActionWorkflowTriggerType = "pre-component-deploy"
	ActionWorkflowTriggerTypePostDeployComponent ActionWorkflowTriggerType = "post-component-deploy"

	// triggers that are run on delete
	ActionWorkflowTriggerTypePreTeardownComponent  ActionWorkflowTriggerType = "pre-component-delete"
	ActionWorkflowTriggerTypePostTeardownComponent ActionWorkflowTriggerType = "post-component-delete"

	// NOTE(jm): the following triggers are going to be deprecated
	// triggers that run on _every_ component deploy
	ActionWorkflowTriggerTypePreDeployAll  ActionWorkflowTriggerType = "pre-deploy"
	ActionWorkflowTriggerTypePostDeployAll ActionWorkflowTriggerType = "post-deploy"
)

type ActionWorkflowTriggerConfig struct {
	ID          string                `json:"id" gorm:"primary_key;check:id_checker,char_length(id)=26" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" gorm:"index:idx_action_workflow_trigger_config_action_workflow_config_id_type,unique" temporaljson:"deleted_at,omitzero,omitempty"`

	// used for RLS
	OrgID string `json:"org_id,omitzero" gorm:"notnull" swaggerignore:"true" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `json:"-" faker:"-" temporaljson:"org,omitzero,omitempty"`

	App   App    `json:"-" swaggerignore:"true" temporaljson:"app,omitzero,omitempty"`
	AppID string `json:"app_id,omitzero" gorm:"notnull;index:idx_app_install_name,unique" temporaljson:"app_id,omitzero,omitempty"`

	// this belongs to an app config id
	AppConfigID string    `json:"app_config_id,omitzero" temporaljson:"app_config_id,omitzero,omitempty"`
	AppConfig   AppConfig `json:"-" temporaljson:"app_config,omitzero,omitempty"`

	ActionWorkflowConfigID string               `json:"action_workflow_config_id,omitzero" gorm:"index:idx_action_workflow_trigger_config_action_workflow_config_id_type,unique" temporaljson:"action_workflow_config_id,omitzero,omitempty"`
	ActionWorkflowConfig   ActionWorkflowConfig `json:"-" temporaljson:"action_workflow_config,omitzero,omitempty"`

	Type ActionWorkflowTriggerType `json:"type,omitzero" swaggertype:"string" gorm:"default null;not null;index:idx_action_workflow_trigger_config_action_workflow_config_id_type,unique" temporaljson:"type,omitzero,omitempty"`

	// individual fields for different types

	CronSchedule string              `json:"cron_schedule,omitzero,omitempty" temporaljson:"cron_schedule,omitzero,omitempty"`
	ComponentID  generics.NullString `json:"component_id,omitzero" swaggertype:"string" temporaljson:"component_id,omitzero,omitempty"`
}

func (a *ActionWorkflowTriggerConfig) BeforeCreate(tx *gorm.DB) error {
	a.ID = domains.NewActionWorkflowTriggerConfigID()
	a.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	a.OrgID = orgIDFromContext(tx.Statement.Context)
	return nil
}
