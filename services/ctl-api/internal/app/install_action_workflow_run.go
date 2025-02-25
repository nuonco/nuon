package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
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
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string                `json:"created_by_id" gorm:"not null;default:null"`
	CreatedBy   Account               `json:"-"`
	CreatedAt   time.Time             `json:"created_at" gorm:"notnull;index:idx_iawr_iaw_id_delete_id_created_at,priority:3,sort:desc;index:idx_install_action_runs_query,priority:3,sort:desc"`
	UpdatedAt   time.Time             `json:"updated_at" gorm:"notnull"`
	DeletedAt   soft_delete.DeletedAt `json:"-" gorm:"index:idx_iawr_iaw_id_delete_id_created_at,priority:2"`

	RunnerJob *RunnerJob `json:"runner_job" gorm:"polymorphic:Owner;"`

	LogStream LogStream `json:"log_stream" gorm:"polymorphic:Owner;"`

	// used for RLS
	OrgID string `json:"org_id" gorm:"notnull" swaggerignore:"true"`
	Org   Org    `json:"-" faker:"-"`

	InstallID string  `json:"install_id" gorm:"not null;default null;index:idx_install_action_runs_query,priority:1"`
	Install   Install `swaggerignore:"true" json:"-" temporaljson:"install"`

	InstallActionWorkflowID string                `json:"install_action_workflow_id" gorm:"index:idx_iawr_iaw_id_delete_id_created_at,priority:1;index:idx_install_action_runs_query,priority:2"`
	InstallActionWorkflow   InstallActionWorkflow `json:"install_action_workflow"`

	Status            InstallActionWorkflowRunStatus `json:"status" gorm:"notnull" swaggertype:"string"`
	StatusDescription string                         `json:"status_description" gorm:"notnull"`

	TriggerType ActionWorkflowTriggerType `json:"trigger_type" gorm:"notnull;default:''"`

	TriggeredByID   string `json:"triggered_by_id" gorm:"type:text;check:triggered_by_id_checker,char_length(id)=26"`
	TriggeredByType string `json:"triggered_by_type" gorm:"type:text;"`

	ActionWorkflowConfigID string               `json:"action_workflow_config_id" gorm:"notnull"`
	ActionWorkflowConfig   ActionWorkflowConfig `json:"config"`

	Steps []InstallActionWorkflowRunStep `json:"steps" gorm:"constraint:OnDelete:CASCADE;"`

	RunEnvVars pgtype.Hstore `json:"run_env_vars" gorm:"type:hstore" swaggertype:"object,string"`

	// after query

	ExecutionTime time.Duration          `json:"execution_time" gorm:"-" swaggertype:"primitive,integer"`
	Outputs       map[string]interface{} `json:"outputs" gorm:"-"`
}

func (i *InstallActionWorkflowRun) Views(db *gorm.DB) []migrations.View {
	return []migrations.View{
		{
			Name: views.CustomViewName(db, &InstallActionWorkflowRun{}, "latest_view_v1"),
			SQL:  viewsql.InstallActionWorkflowLatestRunsViewV1,
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
	return nil
}
