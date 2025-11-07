package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/migrations"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/views"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/viewsql"
)

type InstallDeployType string

const (
	InstallDeployTypeSync     InstallDeployType = "sync-image"
	InstallDeployTypeApply    InstallDeployType = "apply"
	InstallDeployTypeTeardown InstallDeployType = "teardown"
)

type InstallDeployStatus string

const (
	InstallDeployStatusActive          InstallDeployStatus = "active"
	InstallDeployStatusInactive        InstallDeployStatus = "inactive"
	InstallDeployStatusError           InstallDeployStatus = "error"
	InstallDeployStatusNoop            InstallDeployStatus = "noop"
	InstallDeployStatusPlanning        InstallDeployStatus = "planning"
	InstallDeployStatusSyncing         InstallDeployStatus = "syncing"
	InstallDeployStatusExecuting       InstallDeployStatus = "executing"
	InstallDeployStatusCancelled       InstallDeployStatus = "cancelled"
	InstallDeployStatusUnknown         InstallDeployStatus = "unknown"
	InstallDeployStatusPending         InstallDeployStatus = "pending"
	InstallDeployStatusQueued          InstallDeployStatus = "queued"
	InstallDeployStatusPendingApproval InstallDeployStatus = "pending-approval"
	InstallDeployStatusDriftDetected   InstallDeployStatus = "drift-detected"
	InstallDeployStatusAutoSkipped     InstallDeployStatus = "auto-skipped"
	InstallDeployStatusNoDrift         InstallDeployStatus = "no-drift"
	InstallDeployApprovalDenied        InstallDeployStatus = "approval-denied"
)

type InstallDeploy struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"created_by,omitzero" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `gorm:"index" json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	// used for RLS
	OrgID string `json:"org_id,omitzero" gorm:"notnull" swaggerignore:"true" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `json:"-" faker:"-" temporaljson:"org,omitzero,omitempty"`

	// runner details
	RunnerJobs  []RunnerJob `json:"runner_jobs,omitzero" gorm:"polymorphic:Owner;" temporaljson:"runner_jobs,omitzero,omitempty"`
	OCIArtifact OCIArtifact `json:"oci_artifact,omitzero,omitempty" gorm:"polymorphic:Owner;" temporaljson:"oci_artifact,omitempty"`
	LogStream   LogStream   `json:"log_stream,omitzero" gorm:"polymorphic:Owner;" temporaljson:"log_stream,omitzero,omitempty"`

	ActionWorkflowRuns []InstallActionWorkflowRun `json:"action_workflow_runs,omitzero" gorm:"polymorphic:TriggeredBy;" temporaljson:"action_workflow_runs,omitzero,omitempty"`

	ComponentBuildID string         `json:"build_id,omitzero" gorm:"notnull" temporaljson:"component_build_id,omitzero,omitempty"`
	ComponentBuild   ComponentBuild `faker:"-" json:"component_build,omitzero" temporaljson:"component_build,omitzero,omitempty"`

	InstallComponentID string           `json:"install_component_id,omitzero" gorm:"notnull" temporaljson:"install_component_id,omitzero,omitempty"`
	InstallComponent   InstallComponent `faker:"-" json:"-" temporaljson:"install_component,omitzero,omitempty"`

	ComponentReleaseStepID *string               `json:"release_id,omitzero" temporaljson:"component_release_step_id,omitzero,omitempty"`
	ComponentReleaseStep   *ComponentReleaseStep `json:"-" temporaljson:"component_release_step,omitzero,omitempty"`

	Status            InstallDeployStatus `json:"status,omitzero" gorm:"notnull" swaggertype:"string" temporaljson:"status,omitzero,omitempty"`
	StatusDescription string              `json:"status_description,omitzero" gorm:"notnull" temporaljson:"status_description,omitzero,omitempty"`
	Type              InstallDeployType   `json:"install_deploy_type,omitzero" temporaljson:"type,omitzero,omitempty"`
	StatusV2          CompositeStatus     `json:"status_v2,omitzero" gorm:"type:jsonb" temporaljson:"status_v2,omitzero,omitempty"`

	// DEPRECATED: use WorkflowID
	InstallWorkflowID *string   `json:"install_workflow_id,omitzero" gorm:"default null" temporaljson:"install_sandbox_id,omitzero,omitempty"`
	InstallWorkflow   *Workflow `swaggerignore:"true" json:"-" temporaljson:"install_workflow,omitzero,omitempty"`

	// Fields that are de-nested at read time using AfterQuery
	InstallID              string    `json:"install_id" gorm:"-" temporaljson:"install_id,omitzero,omitempty"`
	ComponentID            string    `json:"component_id,omitzero" gorm:"-" temporaljson:"component_id,omitzero,omitempty"`
	ComponentName          string    `json:"component_name,omitzero" gorm:"-" temporaljson:"component_name,omitzero,omitempty"`
	ComponentConfigVersion int       `gorm:"-" json:"component_config_version,omitzero" temporaljson:"component_config_version,omitzero,omitempty"`
	WorkflowID             *string   `json:"workflow_id,omitzero" gorm:"-" temporaljson:"workflow_step_id,omitzero,omitempty"`
	Workflow               *Workflow `json:"workflow,omitzero" gorm:"-" temporaljson:"workflow_step,omitzero,omitempty"`
	PlanOnly               bool      `json:"plan_only" gorm:"-" temporaljson:"plan_only,omitzero,omitempty"`

	Outputs map[string]any `json:"outputs,omitzero" gorm:"-" temporaljson:"outputs,omitzero,omitempty"`
}

func (c *InstallDeploy) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &InstallDeploy{}, "org_id"),
			Columns: []string{
				"org_id",
			},
		},
	}
}

func (c *InstallDeploy) BeforeCreate(tx *gorm.DB) error {
	c.ID = domains.NewDeployID()
	if c.CreatedByID == "" {
		c.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	if c.OrgID == "" {
		c.OrgID = orgIDFromContext(tx.Statement.Context)
	}

	return nil
}

func (c *InstallDeploy) AfterQuery(tx *gorm.DB) error {
	c.InstallID = c.InstallComponent.InstallID
	c.ComponentID = c.InstallComponent.ComponentID
	c.ComponentName = c.InstallComponent.Component.Name
	c.ComponentConfigVersion = c.ComponentBuild.ComponentConfigVersion
	c.WorkflowID = c.InstallWorkflowID
	c.Workflow = c.InstallWorkflow

	outputs := make(map[string]any, 0)
	for _, rj := range c.RunnerJobs {
		// NOTE: omit the create-apply-plan jobs from the outputs
		if rj.Operation != RunnerJobOperationTypeCreateApplyPlan {
			outputs = generics.MergeMaps(outputs, rj.ParsedOutputs)
		}
	}
	c.Outputs = outputs

	if c.StatusV2.Status != "" {
		c.Status = InstallDeployStatus(c.StatusV2.Status)
		c.StatusDescription = c.StatusV2.StatusHumanDescription
	}

	return nil
}

func (c *InstallDeploy) IsTornDown() bool {
	return (generics.SliceContains(c.Status, []InstallDeployStatus{InstallDeployStatusActive, InstallDeployStatusInactive})) && c.Type == InstallDeployTypeTeardown
}

func (i *InstallDeploy) Views(db *gorm.DB) []migrations.View {
	return []migrations.View{
		{
			Name:          views.CustomViewName(db, &InstallDeploy{}, "latest_view_v1"),
			SQL:           viewsql.InstallDeploysLatestViewV1,
			AlwaysReapply: true,
		},
		{
			Name:          views.CustomViewName(db, &InstallDeploy{}, "state_view_v1"),
			SQL:           viewsql.InstallDeploysStateViewV1,
			AlwaysReapply: true,
		},
	}
}
