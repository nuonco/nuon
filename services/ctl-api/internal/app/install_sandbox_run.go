package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
)

type SandboxRunType string

const (
	SandboxRunTypeProvision   SandboxRunType = "provision"
	SandboxRunTypeReprovision SandboxRunType = "reprovision"
	SandboxRunTypeDeprovision SandboxRunType = "deprovision"
)

type SandboxRunStatus string

const (
	SandboxRunStatusActive         SandboxRunStatus = "active"
	SandboxRunStatusError          SandboxRunStatus = "error"
	SandboxRunStatusQueued         SandboxRunStatus = "queued"
	SandboxRunStatusDeprovisioned  SandboxRunStatus = "deprovisioned"
	SandboxRunStatusDeprovisioning SandboxRunStatus = "deprovisioning"
	SandboxRunStatusProvisioning   SandboxRunStatus = "provisioning"
	SandboxRunStatusReprovisioning SandboxRunStatus = "reprovisioning"
	SandboxRunStatusAccessError    SandboxRunStatus = "access_error"
	SandboxRunStatusUnknown        SandboxRunStatus = "unknown"
	SandboxRunStatusEmpty          SandboxRunStatus = "empty"
)

type InstallSandboxRun struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string                `json:"created_by_id" gorm:"not null;default:null"`
	CreatedBy   Account               `json:"created_by"`
	CreatedAt   time.Time             `json:"created_at" gorm:"notnull"`
	UpdatedAt   time.Time             `json:"updated_at" gorm:"notnull"`
	DeletedAt   soft_delete.DeletedAt `json:"-"`

	// runner details
	RunnerJob          RunnerJob                  `json:"runner_job" gorm:"polymorphic:Owner;"`
	LogStream          LogStream                  `json:"log_stream" gorm:"polymorphic:Owner;"`
	ActionWorkflowRuns []InstallActionWorkflowRun `json:"action_workflow_runs" gorm:"polymorphic:TriggeredBy;"`

	// used for RLS
	OrgID     string  `json:"org_id" gorm:"notnull" swaggerignore:"true"`
	Org       Org     `json:"-" faker:"-"`
	InstallID string  `json:"install_id" gorm:"not null;default null"`
	Install   Install `swaggerignore:"true" json:"-"`

	RunType           SandboxRunType   `json:"run_type"`
	Status            SandboxRunStatus `json:"status" gorm:"notnull" swaggertype:"string"`
	StatusDescription string           `json:"status_description" gorm:"notnull"`

	AppSandboxConfigID string           `json:"-"`
	AppSandboxConfig   AppSandboxConfig `json:"app_sandbox_config"`
}

func (i *InstallSandboxRun) BeforeCreate(tx *gorm.DB) error {
	if i.ID == "" {
		i.ID = domains.NewSandboxRunID()
	}

	if i.CreatedByID == "" {
		i.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}

	if i.OrgID == "" {
		i.OrgID = orgIDFromContext(tx.Statement.Context)
	}

	return nil
}
