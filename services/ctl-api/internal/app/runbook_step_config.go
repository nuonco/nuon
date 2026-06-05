package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
)

type RunbookStepType string

const (
	RunbookStepTypeDeploy             RunbookStepType = "deploy"
	RunbookStepTypeAction             RunbookStepType = "action"
	RunbookStepTypeSandboxReprovision RunbookStepType = "sandbox_reprovision"
	RunbookStepTypeSandboxDeprovision RunbookStepType = "sandbox_deprovision"
)

type RunbookStepConfig struct {
	ID          string                `json:"id,omitzero" gorm:"primary_key;check:id_checker,char_length(id)=26" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgID string `json:"org_id,omitzero" gorm:"notnull" swaggerignore:"true" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `json:"-" faker:"-" temporaljson:"org,omitzero,omitempty"`

	RunbookConfigID string        `json:"runbook_config_id,omitzero" gorm:"notnull" temporaljson:"runbook_config_id,omitzero,omitempty"`
	RunbookConfig   RunbookConfig `json:"-" temporaljson:"runbook_config,omitzero,omitempty"`

	Idx  int             `json:"idx" gorm:"notnull;default:0" temporaljson:"idx,omitzero,omitempty"`
	Name string          `json:"name,omitzero" gorm:"notnull" temporaljson:"name,omitzero,omitempty"`
	Type RunbookStepType `json:"type,omitzero" gorm:"notnull" swaggertype:"string" temporaljson:"type,omitzero,omitempty"`

	// deploy fields
	ComponentName      string `json:"component_name,omitzero" temporaljson:"component_name,omitzero,omitempty"`
	DeployDependencies bool   `json:"deploy_dependencies,omitzero" gorm:"default:false" temporaljson:"deploy_dependencies,omitzero,omitempty"`

	// sandbox lifecycle fields
	SkipComponentDeploys bool `json:"skip_component_deploys,omitzero" gorm:"default:false" temporaljson:"skip_component_deploys,omitzero,omitempty"`

	// action reference field
	ActionWorkflowID generics.NullString `json:"action_workflow_id,omitzero" swaggertype:"string" temporaljson:"action_workflow_id,omitzero,omitempty"`

	// inline action fields
	Command        string        `json:"command,omitzero" temporaljson:"command,omitzero,omitempty"`
	InlineContents string        `json:"inline_contents,omitzero" temporaljson:"inline_contents,omitzero,omitempty"`
	EnvVars        pgtype.Hstore `json:"env_vars,omitzero" gorm:"type:hstore" swaggertype:"object,string" temporaljson:"env_vars,omitzero,omitempty"`
	Timeout        time.Duration `json:"timeout,omitzero" gorm:"default:0;not null" swaggertype:"primitive,integer" temporaljson:"timeout,omitzero,omitempty"`
	Role           string        `json:"role,omitzero" temporaljson:"role,omitzero,omitempty"`
}

func (r *RunbookStepConfig) BeforeCreate(tx *gorm.DB) error {
	r.ID = domains.NewRunbookStepConfigID()
	r.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	r.OrgID = orgIDFromContext(tx.Statement.Context)
	return nil
}

func (r *RunbookStepConfig) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &RunbookStepConfig{}, "org_id"),
			Columns: []string{
				"org_id",
			},
		},
	}
}
