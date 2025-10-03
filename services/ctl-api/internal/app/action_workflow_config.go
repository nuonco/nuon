package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/lib/pq"

	"github.com/powertoolsdev/mono/pkg/config/refs"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/migrations"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/views"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/viewsql"
)

type ActionWorkflowConfig struct {
	ID          string                `json:"id,omitzero" gorm:"primary_key;check:id_checker,char_length(id)=26" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" gorm:"index:idx_action_workflow_id_app_config_id,unique" temporaljson:"deleted_at,omitzero,omitempty"`

	// used for RLS
	OrgID string `json:"org_id,omitzero" gorm:"notnull" swaggerignore:"true" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `json:"-" faker:"-" temporaljson:"org,omitzero,omitempty"`

	App   App    `json:"-" swaggerignore:"true" temporaljson:"app,omitzero,omitempty"`
	AppID string `json:"app_id,omitzero" gorm:"notnull;index:idx_app_install_name,unique" temporaljson:"app_id,omitzero,omitempty"`

	AppConfigID string    `json:"app_config_id,omitzero" gorm:"index:idx_action_workflow_id_app_config_id,unique" temporaljson:"app_config_id,omitzero,omitempty"`
	AppConfig   AppConfig `json:"-" temporaljson:"app_config,omitzero,omitempty"`

	ActionWorkflowID string         `json:"action_workflow_id,omitzero" gorm:"index:idx_action_workflow_id_app_config_id,unique" temporaljson:"action_workflow_id,omitzero,omitempty"`
	ActionWorkflow   ActionWorkflow `json:"-" temporaljson:"action_workflow,omitzero,omitempty"`

	// INFO: if adding new associations here, ensure they are added to the batch delete activity
	Triggers []ActionWorkflowTriggerConfig `json:"triggers,omitzero" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"triggers,omitzero,omitempty"`
	Steps    []ActionWorkflowStepConfig    `json:"steps,omitzero" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"steps,omitzero,omitempty"`
	Runs     []InstallActionWorkflowRun    `json:"-" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"runs,omitzero,omitempty"`

	Timeout time.Duration `json:"timeout,omitzero" gorm:"default null;not null" swaggertype:"primitive,integer" temporaljson:"timeout,omitzero,omitempty"`

	ComponentDependencyIDs pq.StringArray `json:"component_dependency_ids" temporaljson:"component_dependency_ids" swaggertype:"array,string" gorm:"type:text[]"`
	References             pq.StringArray `json:"references" temporaljson:"references" swaggertype:"array,string" gorm:"type:text[]"`

	// after query fields

	Refs              []refs.Ref                    `gorm:"-"`
	CronTrigger       *ActionWorkflowTriggerConfig  `json:"-" temporaljson:"cron_trigger,omitzero,omitempty"`
	LifecycleTriggers []ActionWorkflowTriggerConfig `json:"-" temporaljson:"lifecycle_triggers,omitzero,omitempty"`

	BreakGlassRoleARN generics.NullString `json:"break_glass_role_arn,omitzero" temporaljson:"break_glass_role_arn,omitzero,omitempty" swaggertype:"string"`
}

func (a *ActionWorkflowConfig) BeforeCreate(tx *gorm.DB) error {
	a.ID = domains.NewActionWorkflowConfigID()
	a.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	a.OrgID = orgIDFromContext(tx.Statement.Context)
	return nil
}

func (a *ActionWorkflowConfig) AfterQuery(tx *gorm.DB) error {
	cRefs := make([]refs.Ref, 0)
	for _, ref := range a.References {
		cRefs = append(cRefs, refs.NewFromString(ref))
	}
	a.Refs = cRefs

	a.LifecycleTriggers = make([]ActionWorkflowTriggerConfig, 0)

	for _, trigger := range a.Triggers {
		switch trigger.Type {
		case ActionWorkflowTriggerTypeManual:
			continue
		case ActionWorkflowTriggerTypeCron:
			a.CronTrigger = &trigger
		default:
			a.LifecycleTriggers = append(a.LifecycleTriggers, trigger)
		}
	}

	return nil
}

func (i *ActionWorkflowConfig) Views(db *gorm.DB) []migrations.View {
	return []migrations.View{
		{
			Name:          views.CustomViewName(db, &ActionWorkflowConfig{}, "latest_view_v1"),
			SQL:           viewsql.ActionWorkflowConfigsViewV1,
			AlwaysReapply: true,
		},
	}
}

func (a *ActionWorkflowConfig) WorkflowConfigCanTriggerManually() bool {
	for _, trigger := range a.Triggers {
		if trigger.Type == ActionWorkflowTriggerTypeManual {
			return true
		}
	}

	return false
}

func (a *ActionWorkflowConfig) HasComponentTrigger(typ ActionWorkflowTriggerType, componentID string) bool {
	for _, trigger := range a.Triggers {
		if trigger.Type == typ && trigger.ComponentID.ValueString() == componentID {
			return true
		}
	}

	return false
}

func (a *ActionWorkflowConfig) HasTrigger(typ ActionWorkflowTriggerType) bool {
	for _, trigger := range a.Triggers {
		if trigger.Type == typ {
			return true
		}
	}

	return false
}

func (a *ActionWorkflowConfig) GetTriggerIndex(typ ActionWorkflowTriggerType) int {
	for _, trigger := range a.Triggers {
		if trigger.Type == typ {
			return trigger.Index
		}
	}

	return 0
}

func (a *ActionWorkflowConfig) GetComponentTriggerIndex(typ ActionWorkflowTriggerType, componentID string) int {
	for _, trigger := range a.Triggers {
		if trigger.Type == typ && trigger.ComponentID.ValueString() == componentID {
			return trigger.Index
		}
	}

	return 0
}
