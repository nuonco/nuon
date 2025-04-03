package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/migrations"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/views"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/viewsql"
)

type InstallDeployType string

const (
	InstallDeployTypeRelease  InstallDeployType = "release"
	InstallDeployTypeInstall  InstallDeployType = "install"
	InstallDeployTypeTeardown InstallDeployType = "teardown"
	InstallDeployTypePlanOnly InstallDeployType = "plan-only"
)

func (i InstallDeployType) RunnerJobOperationType() RunnerJobOperationType {
	switch i {
	case InstallDeployTypeTeardown:
		return RunnerJobOperationTypeDestroy
	case InstallDeployTypeRelease,
		InstallDeployTypeInstall:
		return RunnerJobOperationTypeCreate
	case InstallDeployTypePlanOnly:
		return RunnerJobOperationTypePlanOnly
	}

	return RunnerJobOperationTypeUnknown
}

type InstallDeployStatus string

const (
	InstallDeployStatusActive    InstallDeployStatus = "active"
	InstallDeployStatusInactive  InstallDeployStatus = "inactive"
	InstallDeployStatusError     InstallDeployStatus = "error"
	InstallDeployStatusNoop      InstallDeployStatus = "noop"
	InstallDeployStatusPlanning  InstallDeployStatus = "planning"
	InstallDeployStatusSyncing   InstallDeployStatus = "syncing"
	InstallDeployStatusExecuting InstallDeployStatus = "executing"
	InstallDeployStatusUnknown   InstallDeployStatus = "unknown"
	InstallDeployStatusPending   InstallDeployStatus = "pending"
	InstallDeployStatusQueued    InstallDeployStatus = "queued"
)

type InstallDeploy struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string                `json:"created_by_id" gorm:"not null;default:null"`
	CreatedBy   Account               `json:"created_by"`
	CreatedAt   time.Time             `json:"created_at" gorm:"notnull"`
	UpdatedAt   time.Time             `json:"updated_at" gorm:"notnull"`
	DeletedAt   soft_delete.DeletedAt `gorm:"index" json:"-"`

	// used for RLS
	OrgID string `json:"org_id" gorm:"notnull" swaggerignore:"true"`
	Org   Org    `json:"-" faker:"-"`

	// runner details
	RunnerJobs []RunnerJob `json:"runner_jobs" gorm:"polymorphic:Owner;"`
	LogStream  LogStream   `json:"log_stream" gorm:"polymorphic:Owner;"`

	ActionWorkflowRuns []InstallActionWorkflowRun `json:"action_workflow_runs" gorm:"polymorphic:TriggeredBy;"`

	ComponentBuildID string         `json:"build_id" gorm:"notnull"`
	ComponentBuild   ComponentBuild `faker:"-" json:"-"`

	InstallComponentID string           `json:"install_component_id" gorm:"notnull"`
	InstallComponent   InstallComponent `faker:"-" json:"-"`

	ComponentReleaseStepID *string               `json:"release_id"`
	ComponentReleaseStep   *ComponentReleaseStep `json:"-"`

	Status            InstallDeployStatus `json:"status" gorm:"notnull" swaggertype:"string"`
	StatusDescription string              `json:"status_description" gorm:"notnull"`
	Type              InstallDeployType   `json:"install_deploy_type"`

	// Fields that are de-nested at read time using AfterQuery
	InstallID              string `json:"install_id" gorm:"-"`
	ComponentID            string `json:"component_id" gorm:"-"`
	ComponentName          string `json:"component_name" gorm:"-"`
	ComponentConfigVersion int    `gorm:"-" json:"component_config_version"`
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
	return nil
}

func (c *InstallDeploy) IsTornDown() bool {
	return (generics.SliceContains(c.Status, []InstallDeployStatus{InstallDeployStatusActive, InstallDeployStatusInactive})) && c.Type == InstallDeployTypeTeardown
}

func (i *InstallDeploy) Views(db *gorm.DB) []migrations.View {
	return []migrations.View{
		{
			Name: views.CustomViewName(db, &InstallDeploy{}, "latest_view_v1"),
			SQL:  viewsql.InstallDeploysLatestViewV1,
		},
	}
}
