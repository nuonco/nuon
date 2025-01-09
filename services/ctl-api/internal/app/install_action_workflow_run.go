package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
)

type InstallActionWorkflowRunStatus string

const (
	InstallActionRunStatusFinished   InstallActionWorkflowRunStatus = "finished"
	InstallActionRunStatusQueued     InstallActionWorkflowRunStatus = "queued"
	InstallActionRunStatusInProgress InstallActionWorkflowRunStatus = "in-progress"
	InstallActionRunStatusError      InstallActionWorkflowRunStatus = "error"
	InstallActionRunStatusTimedOut   InstallActionWorkflowRunStatus = "timed-out"
	InstallActionRunStatusCancelled  InstallActionWorkflowRunStatus = "cancelled"
	InstallActionRunStatusUnknown    InstallActionWorkflowRunStatus = "unknown"
)

type InstallActionWorkflowRun struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string                `json:"created_by_id" gorm:"not null;default:null"`
	CreatedBy   Account               `json:"-"`
	CreatedAt   time.Time             `json:"created_at" gorm:"notnull"`
	UpdatedAt   time.Time             `json:"updated_at" gorm:"notnull"`
	DeletedAt   soft_delete.DeletedAt `json:"-"`

	RunnerJob *RunnerJob `json:"runner_job" gorm:"polymorphic:Owner;"`

	LogStream LogStream `json:"log_stream" gorm:"polymorphic:Owner;"`

	// used for RLS
	OrgID     string  `json:"org_id" gorm:"notnull" swaggerignore:"true"`
	Org       Org     `json:"-" faker:"-"`
	InstallID string  `json:"install_id" gorm:"not null;default null"`
	Install   Install `swaggerignore:"true" json:"-" temporaljson:"install"`

	Status            InstallActionWorkflowRunStatus `json:"status" gorm:"notnull" swaggertype:"string"`
	StatusDescription string                         `json:"status_description" gorm:"notnull"`

	TriggerType ActionWorkflowTriggerType `json:"trigger_type" gorm:"notnull;default:''"`

	ActionWorkflowConfigID string               `json:"action_workflow_config_id" gorm:"notnull"`
	ActionWorkflowConfig   ActionWorkflowConfig `json:"config"`

	Steps []InstallActionWorkflowRunStep `json:"steps" gorm:"constraint:OnDelete:CASCADE;"`

	// after query
	// TODO: update runner to track start and finish timestamps
	ExecutionTime time.Duration `json:"execution_time" gorm:"-" swaggertype:"primitive,integer"`
}

func (i *InstallActionWorkflowRun) BeforeCreate(tx *gorm.DB) error {
	if i.ID == "" {
		i.ID = domains.NewInstallActionWorkflowRunID()
	}

	if i.CreatedByID == "" {
		i.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	if i.OrgID == "" {
		i.OrgID = orgIDFromContext(tx.Statement.Context)
	}
	return nil
}

func (i *InstallActionWorkflowRun) AfterQuery(tx *gorm.DB) error {
	if i.RunnerJob != nil {
		i.ExecutionTime = i.RunnerJob.ExecutionTime
	}
	return nil
}
