package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

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

	ActionWorkflowTriggerTypePreDeploy  ActionWorkflowTriggerType = "pre-deploy"
	ActionWorkflowTriggerTypePostDeploy ActionWorkflowTriggerType = "post-deploy"
)

type ActionWorkflowTriggerConfig struct {
	ID          string                `json:"id" gorm:"primary_key;check:id_checker,char_length(id)=26"`
	CreatedByID string                `json:"created_by_id" gorm:"not null;default:null"`
	CreatedBy   Account               `json:"-"`
	CreatedAt   time.Time             `json:"created_at" gorm:"notnull"`
	UpdatedAt   time.Time             `json:"updated_at" gorm:"notnull"`
	DeletedAt   soft_delete.DeletedAt `json:"-" gorm:"index:idx_action_workflow_trigger_config_action_workflow_config_id_type,unique"`

	// used for RLS
	OrgID string `json:"org_id" gorm:"notnull" swaggerignore:"true"`
	Org   Org    `json:"-" faker:"-"`

	App   App    `json:"-" swaggerignore:"true"`
	AppID string `json:"app_id" gorm:"notnull;index:idx_app_install_name,unique"`

	// this belongs to an app config id
	AppConfigID string    `json:"app_config_id"`
	AppConfig   AppConfig `json:"-"`

	ActionWorkflowConfigID string               `json:"action_workflow_config_id" gorm:"index:idx_action_workflow_trigger_config_action_workflow_config_id_type,unique"`
	ActionWorkflowConfig   ActionWorkflowConfig `json:"-"`

	// individual fields for different types

	Type         ActionWorkflowTriggerType `json:"type" swaggertype:"string" gorm:"default null;not null;index:idx_action_workflow_trigger_config_action_workflow_config_id_type,unique"`
	CronSchedule string                    `json:"cron_schedule,omitempty"`
}

func (a *ActionWorkflowTriggerConfig) BeforeCreate(tx *gorm.DB) error {
	a.ID = domains.NewActionWorkflowTriggerConfigID()
	a.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	a.OrgID = orgIDFromContext(tx.Statement.Context)
	return nil
}
