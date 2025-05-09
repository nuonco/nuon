package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/migrations"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/views"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/viewsql"
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
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"created_by,omitzero" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	// runner details
	RunnerJob          RunnerJob                  `json:"runner_job,omitzero" gorm:"polymorphic:Owner;" temporaljson:"runner_job,omitzero,omitempty"`
	LogStream          LogStream                  `json:"log_stream,omitzero" gorm:"polymorphic:Owner;" temporaljson:"log_stream,omitzero,omitempty"`
	ActionWorkflowRuns []InstallActionWorkflowRun `json:"action_workflow_runs,omitzero" gorm:"polymorphic:TriggeredBy;" temporaljson:"action_workflow_runs,omitzero,omitempty"`

	// used for RLS
	OrgID     string  `json:"org_id,omitzero" gorm:"notnull" swaggerignore:"true" temporaljson:"org_id,omitzero,omitempty"`
	Org       Org     `json:"-" faker:"-" temporaljson:"org,omitzero,omitempty"`
	InstallID string  `json:"install_id,omitzero" gorm:"not null;default null" temporaljson:"install_id,omitzero,omitempty"`
	Install   Install `swaggerignore:"true" json:"-" temporaljson:"install,omitzero,omitempty"`

	// TODO: once we run a backfill we can make this non pointer
	InstallSandboxID *string         `json:"install_sandbox_id,omitzero" gorm:"default null" temporaljson:"install_sandbox_id,omitzero,omitempty"`
	InstallSandbox   *InstallSandbox `swaggerignore:"true" json:"-" temporaljson:"install_sandbox,omitzero,omitempty"`

	InstallWorkflowID *string          `json:"install_workflow_id,omitzero" gorm:"default null" temporaljson:"install_sandbox_id,omitzero,omitempty"`
	InstallWorkflow   *InstallWorkflow `swaggerignore:"true" json:"-" temporaljson:"install_workflow,omitzero,omitempty"`

	RunType           SandboxRunType   `json:"run_type,omitzero" temporaljson:"run_type,omitzero,omitempty"`
	Status            SandboxRunStatus `json:"status,omitzero" gorm:"notnull" swaggertype:"string" temporaljson:"status,omitzero,omitempty"`
	StatusDescription string           `json:"status_description,omitzero" gorm:"notnull" temporaljson:"status_description,omitzero,omitempty"`

	AppSandboxConfigID string           `json:"-" temporaljson:"app_sandbox_config_id,omitzero,omitempty"`
	AppSandboxConfig   AppSandboxConfig `json:"app_sandbox_config,omitzero" temporaljson:"app_sandbox_config,omitzero,omitempty"`
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

	if i.InstallWorkflowID == nil {
		workflow := installWorkflowFromContext(tx.Statement.Context)
		if workflow != nil {
			i.InstallWorkflowID = &workflow.ID
		}
	}

	return nil
}

func (i *InstallSandboxRun) Views(db *gorm.DB) []migrations.View {
	return []migrations.View{
		{
			Name:          views.CustomViewName(db, &InstallSandboxRun{}, "state_view_v1"),
			SQL:           viewsql.InstallSandboxRunsStateViewV1,
			AlwaysReapply: true,
		},
	}
}
