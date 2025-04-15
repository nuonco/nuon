package app

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
)

type ActionWorkflowStepConfig struct {
	ID          string                `json:"id" gorm:"primary_key;check:id_checker,char_length(id)=26"`
	CreatedByID string                `json:"created_by_id" gorm:"not null;default:null"`
	CreatedBy   Account               `json:"-"`
	CreatedAt   time.Time             `json:"created_at" gorm:"notnull"`
	UpdatedAt   time.Time             `json:"updated_at" gorm:"notnull"`
	DeletedAt   soft_delete.DeletedAt `json:"-" gorm:"index:idx_action_workflow_step_config_action_workflow_config_id_name,unique"`

	// used for RLS
	OrgID string `json:"org_id" gorm:"notnull" swaggerignore:"true"`
	Org   Org    `json:"-" faker:"-"`

	App   App    `json:"-" swaggerignore:"true"`
	AppID string `json:"app_id" gorm:"notnull;index:idx_app_install_name,unique"`

	// this belongs to an app config id
	AppConfigID string    `json:"app_config_id"`
	AppConfig   AppConfig `json:"-"`

	ActionWorkflowConfigID string               `json:"action_workflow_config_id" gorm:"index:idx_action_workflow_step_config_action_workflow_config_id_name,unique"`
	ActionWorkflowConfig   ActionWorkflowConfig `json:"-"`

	// metadata
	Name           string `json:"name" gorm:"index:idx_action_workflow_step_config_action_workflow_config_id_name,unique"`
	PreviousStepID string `json:"previous_step_id"`
	Idx            int    `json:"idx"`

	// all the details needed for a step
	PublicGitVCSConfig       *PublicGitVCSConfig       `gorm:"polymorphic:ComponentConfig;constraint:OnDelete:CASCADE;" json:"public_git_vcs_config,omitempty"`
	ConnectedGithubVCSConfig *ConnectedGithubVCSConfig `gorm:"polymorphic:ComponentConfig;constraint:OnDelete:CASCADE;" json:"connected_github_vcs_config,omitempty"`
	VCSConnectionType        VCSConnectionType         `json:"-" gorm:"-"`

	EnvVars        pgtype.Hstore `json:"env_vars" gorm:"type:hstore" swaggertype:"object,string"`
	Command        string        `json:"command"`
	InlineContents string        `json:"inline_contents"`
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
