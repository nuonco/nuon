package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
)

type ActionWorkflowConfig struct {
	ID          string                `json:"id" gorm:"primary_key;check:id_checker,char_length(id)=26"`
	CreatedByID string                `json:"created_by_id" gorm:"not null;default:null"`
	CreatedBy   Account               `json:"-"`
	CreatedAt   time.Time             `json:"created_at" gorm:"notnull"`
	UpdatedAt   time.Time             `json:"updated_at" gorm:"notnull"`
	DeletedAt   soft_delete.DeletedAt `json:"-" gorm:"index:idx_action_workflow_id_app_config_id,unique"`

	// used for RLS
	OrgID string `json:"org_id" gorm:"notnull" swaggerignore:"true"`
	Org   Org    `json:"-" faker:"-"`

	App   App    `json:"-" swaggerignore:"true" json:"app"`
	AppID string `json:"app_id" gorm:"notnull;index:idx_app_install_name,unique"`

	// this belongs to an app config id
	AppConfigID string    `json:"app_config_id" gorm:"index:idx_action_workflow_id_app_config_id,unique"`
	AppConfig   AppConfig `json:"-"`

	ActionWorkflowID string         `json:"action_workflow_id" gorm:"index:idx_action_workflow_id_app_config_id,unique"`
	ActionWorkflow   ActionWorkflow `json:"-"`

	Triggers []ActionWorkflowTriggerConfig `json:"triggers"`
	Steps    []ActionWorkflowStepConfig    `json:"steps"`
}

func (a *ActionWorkflowConfig) BeforeCreate(tx *gorm.DB) error {
	a.ID = domains.NewActionWorkflowID()
	a.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	a.OrgID = orgIDFromContext(tx.Statement.Context)
	return nil
}
