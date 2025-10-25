package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/migrations"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/views"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/viewsql"
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
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"created_by,omitzero" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" gorm:"notnull;index:idx_install_action_runs_query,priority:3,sort:desc" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	RunnerJob *RunnerJob `json:"runner_job,omitzero" gorm:"polymorphic:Owner;" temporaljson:"runner_job,omitzero,omitempty"`

	LogStream LogStream `json:"log_stream,omitzero" gorm:"polymorphic:Owner;" temporaljson:"log_stream,omitzero,omitempty"`

	// used for RLS
	OrgID string `json:"org_id,omitzero" gorm:"notnull" swaggerignore:"true" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `json:"-" faker:"-" temporaljson:"org,omitzero,omitempty"`

	InstallID string  `json:"install_id,omitzero" gorm:"not null;default null;index:idx_install_action_runs_query,priority:1" temporaljson:"install_id,omitzero,omitempty"`
	Install   Install `swaggerignore:"true" json:"-" temporaljson:"install,omitzero,omitempty"`

	InstallActionWorkflowID string                `json:"install_action_workflow_id,omitzero" gorm:"index:idx_install_action_runs_query,priority:2" temporaljson:"install_action_workflow_id,omitzero,omitempty"`
	InstallActionWorkflow   InstallActionWorkflow `json:"install_action_workflow,omitzero" temporaljson:"install_action_workflow,omitzero,omitempty"`

	Status            InstallActionWorkflowRunStatus `json:"status,omitzero" gorm:"notnull" swaggertype:"string" temporaljson:"status,omitzero,omitempty"`
	StatusDescription string                         `json:"status_description,omitzero" gorm:"notnull" temporaljson:"status_description,omitzero,omitempty"`
	StatusV2          CompositeStatus                `json:"status_v2,omitzero" gorm:"type:jsonb" temporaljson:"status_v2,omitzero,omitempty"`

	TriggerType ActionWorkflowTriggerType `json:"trigger_type,omitzero" gorm:"notnull;default:''" temporaljson:"trigger_type,omitzero,omitempty"`

	TriggeredByID   string `json:"triggered_by_id,omitzero" gorm:"type:text;check:triggered_by_id_checker,char_length(id)=26" temporaljson:"triggered_by_id,omitzero,omitempty"`
	TriggeredByType string `json:"triggered_by_type,omitzero" gorm:"type:text;" temporaljson:"triggered_by_type,omitzero,omitempty"`

	ActionWorkflowConfigID string               `json:"action_workflow_config_id,omitzero" gorm:"notnull" temporaljson:"action_workflow_config_id,omitzero,omitempty"`
	ActionWorkflowConfig   ActionWorkflowConfig `json:"config,omitzero" temporaljson:"action_workflow_config,omitzero,omitempty"`

	Steps []InstallActionWorkflowRunStep `json:"steps,omitzero" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"steps,omitzero,omitempty"`

	RunEnvVars pgtype.Hstore `json:"run_env_vars,omitzero" gorm:"type:hstore" swaggertype:"object,string" temporaljson:"run_env_vars,omitzero,omitempty"`

	InstallWorkflowID *string   `json:"install_workflow_id" gorm:"default null" temporaljson:"install_sandbox_id,omitzero,omitempty"`
	InstallWorkflow   *Workflow `swaggerignore:"true" json:"-" temporaljson:"install_workflow,omitzero,omitempty"`

	// after query
	ExecutionTime time.Duration          `json:"execution_time,omitzero" gorm:"-" swaggertype:"primitive,integer" temporaljson:"execution_time,omitzero,omitempty"`
	Outputs       map[string]interface{} `json:"outputs,omitzero" gorm:"-" temporaljson:"outputs,omitzero,omitempty"`
	WorkflowID    *string                `json:"workflow_id,omitzero" gorm:"-" temporaljson:"workflow_step_id,omitzero,omitempty"`
	Workflow      *Workflow              `json:"workflow,omitzero" gorm:"-" temporaljson:"workflow_step,omitzero,omitempty"`
}

func (i *InstallActionWorkflowRun) Views(db *gorm.DB) []migrations.View {
	return []migrations.View{
		{
			Name:          views.CustomViewName(db, &InstallActionWorkflowRun{}, "latest_view_v1"),
			SQL:           viewsql.InstallActionWorkflowLatestRunsViewV1,
			AlwaysReapply: true,
		},
		{
			Name:          views.CustomViewName(db, &InstallActionWorkflowRun{}, "state_view_v1"),
			SQL:           viewsql.InstallActionWorkflowLatestRunsViewV1,
			AlwaysReapply: true,
		},
	}
}

func (i *InstallActionWorkflowRun) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &InstallActionWorkflowRun{}, "preload"),
			Columns: []string{
				"install_action_workflow_id",
				"deleted_at",
				"created_at DESC",
			},
		},
		{
			Name: indexes.Name(db, &InstallActionWorkflowRun{}, "triggered"),
			Columns: []string{
				"triggered_by_type",
				"triggered_by_id",
				"deleted_at",
			},
		},
	}
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

		if len(i.RunnerJob.ParsedOutputs) > 0 {
			i.Outputs = i.RunnerJob.ParsedOutputs
		}
	}

	if i.StatusV2.Status != "" {
		i.Status = InstallActionWorkflowRunStatus(i.StatusV2.Status)
		i.StatusDescription = i.StatusV2.StatusHumanDescription
	}

	i.WorkflowID = i.InstallWorkflowID
	i.Workflow = i.InstallWorkflow

	return nil
}
