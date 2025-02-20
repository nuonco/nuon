package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
)

type InstallActionWorkflow struct {
	ID          string                `json:"id" gorm:"primary_key;check:id_checker,char_length(id)=26"`
	CreatedByID string                `json:"created_by_id" gorm:"not null;default:null"`
	CreatedBy   Account               `json:"-"`
	CreatedAt   time.Time             `json:"created_at" gorm:"notnull"`
	UpdatedAt   time.Time             `json:"updated_at" gorm:"notnull"`
	DeletedAt   soft_delete.DeletedAt `json:"-" gorm:"index:idx_install_action_workflow_id,unique;index:idx_iaw_org_id_install_id,priority:3"`

	// used for RLS
	OrgID string `json:"org_id" gorm:"notnull;index:idx_iaw_org_id_install_id,priority:1" swaggerignore:"true"`
	Org   Org    `json:"-" faker:"-"`

	Install   Install `json:"-" swaggerignore:"true"`
	InstallID string  `json:"install_id" gorm:"index:idx_install_action_workflow_id,unique;index:idx_iaw_org_id_install_id,priority:2" faker:"-"`

	ActionWorkflow   ActionWorkflow `json:"action_workflow"`
	ActionWorkflowID string         `json:"action_workflow_id" gorm:"index:idx_install_action_workflow_id,unique"`

	Runs []InstallActionWorkflowRun `faker:"-" gorm:"constraint:OnDelete:CASCADE;" json:"runs"`

	// after query fields filled in after querying
	Status InstallActionWorkflowRunStatus `json:"status" gorm:"-" swaggertype:"string"`
}

func (a *InstallActionWorkflow) BeforeCreate(tx *gorm.DB) error {
	a.ID = domains.NewInstallActionWorkflowConfigID()
	a.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	a.OrgID = orgIDFromContext(tx.Statement.Context)
	return nil
}

func (c *InstallActionWorkflow) AfterQuery(tx *gorm.DB) error {
	c.Status = InstallActionRunStatusUnknown
	if len(c.Runs) > 0 {
		c.Status = c.Runs[0].Status
	}

	return nil
}
