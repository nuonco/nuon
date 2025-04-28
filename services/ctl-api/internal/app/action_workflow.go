package app

import (
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
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
	Status            ActionWorkflowStatus  `json:"status" gorm:"notnull;default:'active'" swaggertype:"string" temporaljson:"status,omitzero,omitempty"`
	StatusDescription string                `json:"status_description" temporaljson:"status_description,omitzero,omitempty"`
	CreatedByID       string                `json:"created_by_id" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy         Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt         time.Time             `json:"created_at" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt         time.Time             `json:"updated_at" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt         soft_delete.DeletedAt `json:"-" gorm:"index:idx_action_workflow_app_id_name,unique" temporaljson:"deleted_at,omitzero,omitempty"`

	// used for RLS
	OrgID string `json:"org_id" gorm:"notnull" swaggerignore:"true" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `json:"-" faker:"-" temporaljson:"org,omitzero,omitempty"`

	App   App    `json:"-" swaggerignore:"true" temporaljson:"app,omitzero,omitempty"`
	AppID string `json:"app_id" gorm:"index:idx_action_workflow_app_id_name,unique" faker:"-" temporaljson:"app_id,omitzero,omitempty"`

	Configs     []ActionWorkflowConfig `json:"configs" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"configs,omitzero,omitempty"`
	ConfigCount int                    `json:"config_count" gorm:"->;-:migration" temporaljson:"config_count,omitzero,omitempty"`

	// metadata
	Name string `json:"name" gorm:"index:idx_action_workflow_app_id_name,unique" temporaljson:"name,omitzero,omitempty"`
}

func (a *ActionWorkflow) BeforeCreate(tx *gorm.DB) error {
	a.ID = domains.NewActionWorkflowID()
	a.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	a.OrgID = orgIDFromContext(tx.Statement.Context)
	return nil
}

func (a *ActionWorkflow) BeforeDelete(tx *gorm.DB) error {
	if a.ID == "" {
		return nil
	}

	configs := []ActionWorkflowConfig{}
	resp := tx.Find(&configs, " action_workflow_id = ?", a.ID)
	if resp.Error != nil {
		return resp.Error
	}

	for _, config := range configs {
		installActionWorkflowRuns := []InstallActionWorkflowRun{}
		resp = tx.Select(clause.Associations).Delete(&installActionWorkflowRuns, " action_workflow_config_id = ?", config.ID)
		if resp.Error != nil {
			return fmt.Errorf("error deleting install action workflow runs: %w", resp.Error)
		}

		triggers := []ActionWorkflowTriggerConfig{}
		resp := tx.Delete(&triggers, " action_workflow_config_id = ?", config.ID)
		if resp.Error != nil {
			return fmt.Errorf("error deleting action workflow triggers: %w", resp.Error)
		}

		steps := []ActionWorkflowStepConfig{}
		resp = tx.Delete(&steps, " action_workflow_config_id = ?", config.ID)
		if resp.Error != nil {
			return fmt.Errorf("error deleting action workflow steps: %w", resp.Error)
		}

		resp = tx.Delete(&config)
		if resp.Error != nil {
			return fmt.Errorf("error deleting action workflow config: %w", resp.Error)
		}
	}

	installActionWorkflows := []InstallActionWorkflow{}
	resp = tx.Delete(&installActionWorkflows, " action_workflow_id = ?", a.ID)
	if resp.Error != nil {
		return fmt.Errorf("error deleting install action workflows: %w", resp.Error)
	}
	return nil
}
