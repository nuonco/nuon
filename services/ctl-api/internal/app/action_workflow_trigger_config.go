package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/migrations"
)

type ActionWorkflowTriggerType string

const (
	// this is for manual debugging/triggering in the ui
	ActionWorkflowTriggerTypeManual ActionWorkflowTriggerType = "manual"

	// run on a hook
	ActionWorkflowTriggerTypeCron ActionWorkflowTriggerType = "cron"

	// individaul component ones
	ActionWorkflowTriggerTypePreDeployComponent  ActionWorkflowTriggerType = "pre-deploy-component"
	ActionWorkflowTriggerTypePostDeployComponent ActionWorkflowTriggerType = "post-deploy-component"

	ActionWorkflowTriggerTypePreTeardownComponent  ActionWorkflowTriggerType = "pre-teardown-component"
	ActionWorkflowTriggerTypePostTeardownComponent ActionWorkflowTriggerType = "post-teardown-component"

	// internals
	ActionWorkflowTriggerTypePreSecretsSync  ActionWorkflowTriggerType = "pre-secrets-sync"
	ActionWorkflowTriggerTypePostSecretsSync ActionWorkflowTriggerType = "post-secrets-sync"

	// workflow triggers
	ActionWorkflowTriggerTypePreProvision  ActionWorkflowTriggerType = "pre-provision"
	ActionWorkflowTriggerTypePostProvision ActionWorkflowTriggerType = "post-provision"

	ActionWorkflowTriggerTypePreReprovision  ActionWorkflowTriggerType = "pre-reprovision"
	ActionWorkflowTriggerTypePostReprovision ActionWorkflowTriggerType = "post-reprovision"

	ActionWorkflowTriggerTypePreDeprovision  ActionWorkflowTriggerType = "pre-deprovision"
	ActionWorkflowTriggerTypePostDeprovision ActionWorkflowTriggerType = "post-deprovision"

	ActionWorkflowTriggerTypePreDeployAllComponents  ActionWorkflowTriggerType = "pre-deploy-all-components"
	ActionWorkflowTriggerTypePostDeployAllComponents ActionWorkflowTriggerType = "post-deploy-all-components"

	ActionWorkflowTriggerTypePreTeardownAllComponents  ActionWorkflowTriggerType = "pre-teardown-all-components"
	ActionWorkflowTriggerTypePostTeardownAllComponents ActionWorkflowTriggerType = "post-teardown-all-components"

	ActionWorkflowTriggerTypePreDeprovisionSandbox  ActionWorkflowTriggerType = "pre-deprovision-sandbox"
	ActionWorkflowTriggerTypePostDeprovisionSandbox ActionWorkflowTriggerType = "post-deprovision-sandbox"

	ActionWorkflowTriggerTypePreReprovisionSandbox  ActionWorkflowTriggerType = "pre-reprovision-sandbox"
	ActionWorkflowTriggerTypePostReprovisionSandbox ActionWorkflowTriggerType = "post-reprovision-sandbox"

	ActionWorkflowTriggerTypePreUpdateInputs  ActionWorkflowTriggerType = "pre-update-inputs"
	ActionWorkflowTriggerTypePostUpdateInputs ActionWorkflowTriggerType = "post-update-inputs"
)

// These component types require a component to be passed with them
var AllActionWorkflowComponentTriggerTypes = []ActionWorkflowTriggerType{
	ActionWorkflowTriggerTypePreDeployComponent,
	ActionWorkflowTriggerTypePostDeployComponent,
	ActionWorkflowTriggerTypePreTeardownComponent,
	ActionWorkflowTriggerTypePostTeardownComponent,
}

// All component types
var AllActionWorkflowTriggerTypes = []ActionWorkflowTriggerType{
	ActionWorkflowTriggerTypeManual,
	ActionWorkflowTriggerTypeCron,
	ActionWorkflowTriggerTypePreDeployComponent,
	ActionWorkflowTriggerTypePostDeployComponent,
	ActionWorkflowTriggerTypePreTeardownComponent,
	ActionWorkflowTriggerTypePostTeardownComponent,
	ActionWorkflowTriggerTypePreProvision,
	ActionWorkflowTriggerTypePostProvision,
	ActionWorkflowTriggerTypePreReprovision,
	ActionWorkflowTriggerTypePostReprovision,
	ActionWorkflowTriggerTypePreDeprovision,
	ActionWorkflowTriggerTypePostDeprovision,
	ActionWorkflowTriggerTypePreDeployAllComponents,
	ActionWorkflowTriggerTypePostDeployAllComponents,
	ActionWorkflowTriggerTypePreTeardownAllComponents,
	ActionWorkflowTriggerTypePostTeardownAllComponents,
	ActionWorkflowTriggerTypePreDeprovisionSandbox,
	ActionWorkflowTriggerTypePostDeprovisionSandbox,
	ActionWorkflowTriggerTypePreReprovisionSandbox,
	ActionWorkflowTriggerTypePostReprovisionSandbox,
	ActionWorkflowTriggerTypePreUpdateInputs,
	ActionWorkflowTriggerTypePostUpdateInputs,
}

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

	Index int `json:"index,omitzero" swaggertype:"integer" gorm:"default:0;"`

	// individual fields for different types

	CronSchedule string              `json:"cron_schedule,omitzero,omitempty" temporaljson:"cron_schedule,omitzero,omitempty"`
	ComponentID  generics.NullString `json:"component_id,omitzero" swaggertype:"string" temporaljson:"component_id,omitzero,omitempty"`
	Component    *Component          `json:"component" temporaljson:"component,omitzero,omitempty"`
}

func (a *ActionWorkflowTriggerConfig) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &ActionWorkflowTriggerConfig{}, "org_id"),
			Columns: []string{
				"org_id",
			},
		},
	}
}

func (a *ActionWorkflowTriggerConfig) BeforeCreate(tx *gorm.DB) error {
	a.ID = domains.NewActionWorkflowTriggerConfigID()
	a.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	a.OrgID = orgIDFromContext(tx.Statement.Context)
	return nil
}
