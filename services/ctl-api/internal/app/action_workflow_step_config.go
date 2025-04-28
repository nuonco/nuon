package app

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
)

type ActionWorkflowStepConfig struct {
	ID          string                `json:"id" gorm:"primary_key;check:id_checker,char_length(id)=26" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" gorm:"index:idx_action_workflow_step_config_action_workflow_config_id_name,unique" temporaljson:"deleted_at,omitzero,omitempty"`

	// used for RLS
	OrgID string `json:"org_id" gorm:"notnull" swaggerignore:"true" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `json:"-" faker:"-" temporaljson:"org,omitzero,omitempty"`

	App   App    `json:"-" swaggerignore:"true" temporaljson:"app,omitzero,omitempty"`
	AppID string `json:"app_id" gorm:"notnull;index:idx_app_install_name,unique" temporaljson:"app_id,omitzero,omitempty"`

	// this belongs to an app config id
	AppConfigID string    `json:"app_config_id" temporaljson:"app_config_id,omitzero,omitempty"`
	AppConfig   AppConfig `json:"-" temporaljson:"app_config,omitzero,omitempty"`

	ActionWorkflowConfigID string               `json:"action_workflow_config_id" gorm:"index:idx_action_workflow_step_config_action_workflow_config_id_name,unique" temporaljson:"action_workflow_config_id,omitzero,omitempty"`
	ActionWorkflowConfig   ActionWorkflowConfig `json:"-" temporaljson:"action_workflow_config,omitzero,omitempty"`

	// metadata
	Name           string `json:"name" gorm:"index:idx_action_workflow_step_config_action_workflow_config_id_name,unique" temporaljson:"name,omitzero,omitempty"`
	PreviousStepID string `json:"previous_step_id" temporaljson:"previous_step_id,omitzero,omitempty"`
	Idx            int    `json:"idx" temporaljson:"idx,omitzero,omitempty"`

	// all the details needed for a step
	PublicGitVCSConfig       *PublicGitVCSConfig       `gorm:"polymorphic:ComponentConfig;constraint:OnDelete:CASCADE;" json:"public_git_vcs_config,omitempty" temporaljson:"public_git_vcs_config,omitzero,omitempty"`
	ConnectedGithubVCSConfig *ConnectedGithubVCSConfig `gorm:"polymorphic:ComponentConfig;constraint:OnDelete:CASCADE;" json:"connected_github_vcs_config,omitempty" temporaljson:"connected_github_vcs_config,omitzero,omitempty"`
	VCSConnectionType        VCSConnectionType         `json:"-" gorm:"-" temporaljson:"vcs_connection_type,omitzero,omitempty"`

	EnvVars        pgtype.Hstore `json:"env_vars" gorm:"type:hstore" swaggertype:"object,string" temporaljson:"env_vars,omitzero,omitempty"`
	Command        string        `json:"command" temporaljson:"command,omitzero,omitempty"`
	InlineContents string        `json:"inline_contents" temporaljson:"inline_contents,omitzero,omitempty"`
}

func (a *ActionWorkflowStepConfig) BeforeCreate(tx *gorm.DB) error {
	a.ID = domains.NewActionWorkflowStepConfigID()
	a.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	a.OrgID = orgIDFromContext(tx.Statement.Context)
	return nil
}

func (a *ActionWorkflowStepConfig) AfterQuery(tx *gorm.DB) error {
	if a.EnvVars == nil {
		a.EnvVars = pgtype.Hstore{}
	}

	// set the vcs connection type correctly
	if a.ConnectedGithubVCSConfig != nil {
		a.VCSConnectionType = VCSConnectionTypeConnectedRepo
	} else if a.PublicGitVCSConfig != nil {
		a.VCSConnectionType = VCSConnectionTypePublicRepo
	} else {
		a.VCSConnectionType = VCSConnectionTypeNone
	}

	return nil
}
