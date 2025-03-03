package app

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/migrations"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/views"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/viewsql"
)

type RunnerJobStatus string

const (
	// all jobs are set as queued to start, and the event loop should update them to available.
	RunnerJobStatusQueued RunnerJobStatus = "queued"
	// the runner queries jobs that are available, to find something to work on
	RunnerJobStatusAvailable RunnerJobStatus = "available"
	// once a runner is actively working on the job
	RunnerJobStatusInProgress RunnerJobStatus = "in-progress"
	// once a runner has finished the job
	RunnerJobStatusFinished RunnerJobStatus = "finished"

	// once a runner has failed the job
	RunnerJobStatusFailed RunnerJobStatus = "failed"
	// once the job has timed out
	RunnerJobStatusTimedOut RunnerJobStatus = "timed-out"
	// not attempted is when the runner can not attempt
	RunnerJobStatusNotAttempted RunnerJobStatus = "not-attempted"
	// cancelled
	RunnerJobStatusCancelled RunnerJobStatus = "cancelled"
	// status is not known
	RunnerJobStatusUnknown RunnerJobStatus = "unknown"
)

type RunnerJobGroup string

const (
	// a health check is a runner health check, not to be confused with a heart beat.
	RunnerJobGroupHealthChecks RunnerJobGroup = "health-checks"

	// component groups for builds, syncing and deploys
	RunnerJobGroupSync   RunnerJobGroup = "sync"
	RunnerJobGroupBuild  RunnerJobGroup = "build"
	RunnerJobGroupDeploy RunnerJobGroup = "deploy"

	// sandbox jobs such as provision, deprovision.
	RunnerJobGroupSandbox RunnerJobGroup = "sandbox"

	// runner jobs such as provision, deprovision and pre-flight checks.
	RunnerJobGroupRunner RunnerJobGroup = "runner"

	// operations jobs such as shutdown, restart, noop and update settings.
	RunnerJobGroupOperations RunnerJobGroup = "operations"

	// actions workflows
	RunnerJobGroupActions RunnerJobGroup = "actions"

	RunnerJobGroupUnknown RunnerJobGroup = ""
	RunnerJobGroupAny     RunnerJobGroup = "any"
)

type RunnerJobType string

const (
	// a health check is a runner health check, not to be confused with a heart beat
	RunnerJobTypeHealthCheck RunnerJobType = "health-check"

	// build job types
	RunnerJobTypeDockerBuild          RunnerJobType = "docker-build"
	RunnerJobTypeContainerImageBuild  RunnerJobType = "container-image-build"
	RunnerJobTypeTerraformModuleBuild RunnerJobType = "terraform-module-build"
	RunnerJobTypeHelmChartBuild       RunnerJobType = "helm-chart-build"
	RunnerJobTypeNOOPBuild            RunnerJobType = "noop-build"

	// sync job types
	RunnerJobTypeOCISync  RunnerJobType = "oci-sync"
	RunnerJobTypeNOOPSync RunnerJobType = "noop-sync"

	// deploy job types
	RunnerJobTypeTerraformDeploy RunnerJobType = "terraform-deploy"
	RunnerJobTypeHelmChartDeploy RunnerJobType = "helm-chart-deploy"
	RunnerJobTypeJobDeploy       RunnerJobType = "job-deploy"
	RunnerJobTypeJobNOOPDeploy   RunnerJobType = "noop-deploy"

	// operations job types
	RunnerJobTypeShutDown      RunnerJobType = "shut-down"
	RunnerJobTypeUpdateVersion RunnerJobType = "update-version"
	RunnerJobTypeNOOP          RunnerJobType = "noop"

	// sandbox job types
	RunnerJobTypeSandboxTerraform RunnerJobType = "sandbox-terraform"

	// runner job types
	RunnerJobTypeRunnerHelm      RunnerJobType = "runner-helm"
	RunnerJobTypeRunnerTerraform RunnerJobType = "runner-terraform"
	RunnerJobTypeRunnerLocal     RunnerJobType = "runner-local"

	// actions job types
	RunnerJobTypeActionsWorkflowRun RunnerJobType = "actions-workflow"

	// unknown
	RunnerJobTypeUnknown = "unknown"
)

func (r RunnerJobType) Group() RunnerJobGroup {
	switch r {

	// builds
	case RunnerJobTypeDockerBuild,
		RunnerJobTypeContainerImageBuild,
		RunnerJobTypeNOOPBuild,
		RunnerJobTypeTerraformModuleBuild,
		RunnerJobTypeHelmChartBuild:
		return RunnerJobGroupBuild

		// syncing
	case RunnerJobTypeOCISync,
		RunnerJobTypeNOOPSync:
		return RunnerJobGroupSync

		// deploys
	case RunnerJobTypeHelmChartDeploy,
		RunnerJobTypeTerraformDeploy,
		RunnerJobTypeJobDeploy,
		RunnerJobTypeJobNOOPDeploy:
		return RunnerJobGroupDeploy

		// runners
	case RunnerJobTypeRunnerHelm, RunnerJobTypeRunnerTerraform:
		return RunnerJobGroupRunner

		// sandboxes
	case RunnerJobTypeSandboxTerraform:
		return RunnerJobGroupSandbox

		// health checks
	case RunnerJobTypeHealthCheck:
		return RunnerJobGroupHealthChecks

		// operations
	case RunnerJobTypeNOOP, RunnerJobTypeShutDown, RunnerJobTypeUpdateVersion:
		return RunnerJobGroupOperations

	case RunnerJobTypeActionsWorkflowRun:
		return RunnerJobGroupActions

	default:
		return RunnerJobGroupUnknown
	}
}

// operation types that correspond to the type of operation
type RunnerJobOperationType string

const (
	// exec is used for shut down, scripts and more. It is mainly ignored as those job types do not really need to
	// think about operations
	RunnerJobOperationTypeExec RunnerJobOperationType = "exec"

	// the following operations are for common use cases for things such as helm, terraform and other jobs that have
	// multiple operation types.
	RunnerJobOperationTypeCreate   RunnerJobOperationType = "apply"
	RunnerJobOperationTypeDestroy  RunnerJobOperationType = "destroy"
	RunnerJobOperationTypePlanOnly RunnerJobOperationType = "plan-only"
	RunnerJobOperationTypeBuild    RunnerJobOperationType = "build"

	RunnerJobOperationTypeUnknown RunnerJobOperationType = "unknown"
)

type RunnerJob struct {
	ID          string  `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string  `json:"created_by_id" gorm:"not null;default:null"`
	CreatedBy   Account `json:"-"`

	CreatedAt time.Time             `json:"created_at" gorm:"notnull;index:idx_runner_jobs_query,priority:4,sort:desc;index:idx_runner_jobs_owner_id,priority:2,sort:desc"`
	UpdatedAt time.Time             `json:"updated_at" gorm:"notnull"`
	DeletedAt soft_delete.DeletedAt `json:"-" gorm:"index:idx_runner_name,unique;"`

	OrgID string `json:"org_id" gorm:"index:idx_app_name,unique"`
	Org   Org    `json:"-"`

	RunnerID    string  `json:"runner_id" gorm:"index:idx_runner_name,unique;index:idx_runner_jobs_query,priority:1"`
	OwnerID     string  `json:"owner_id" gorm:"type:text;check:owner_id_checker,char_length(id)=26;index:idx_runner_jobs_owner_id,priority:1"`
	OwnerType   string  `json:"owner_type" gorm:"type:text;"`
	LogStreamID *string `json:"log_stream_id"`

	// queue timeout is how long a job can be queued, before being made available
	QueueTimeout time.Duration `json:"queue_timeout" gorm:"default null;not null" swaggertype:"primitive,integer"`
	// available timeout is how long a job can be marked as "available" before being requeued
	AvailableTimeout time.Duration `json:"available_timeout" gorm:"default null;not null" swaggertype:"primitive,integer"`
	// execution timeout is how long a job can be marked as "exeucuting" before being requeued
	ExecutionTimeout time.Duration `json:"execution_timeout" gorm:"default null;not null" swaggertype:"primitive,integer"`

	// overall timeout is how long a job can be attempted, before being cancelled
	OverallTimeout time.Duration `json:"overall_timeout" gorm:"default null;not null" swaggertype:"primitive,integer"`

	MaxExecutions int `json:"max_executions" gorm:"not null;default null"`

	Status            RunnerJobStatus `json:"status" gorm:"not null;default null;index:idx_runner_jobs_query,priority:3"`
	StatusDescription string          `json:"status_description" gorm:"not null;default null"`

	Type      RunnerJobType          `json:"type" gorm:"default null;not null"`
	Group     RunnerJobGroup         `json:"group" gorm:"default:null;not null;index:idx_runner_jobs_query,priority:2"`
	Operation RunnerJobOperationType `json:"operation" gorm:"default:null;not null"`

	Executions []RunnerJobExecution `json:"executions" gorm:"constraint:OnDelete:CASCADE;"`
	Plan       RunnerJobPlan        `json:"-" gorm:"constraint:OnDelete:CASCADE;"`

	StartedAt  time.Time `json:"started_at"  gorm:"default:null"`
	FinishedAt time.Time `json:"finished_at" gorm:"default:null"`

	Metadata pgtype.Hstore `json:"metadata" gorm:"type:hstore" swaggertype:"object,string"`

	// read only fields from view

	ExecutionCount            int    `json:"execution_count" gorm:"->;-:migration"`
	FinalRunnerJobExecutionID string `json:"final_runner_job_execution_id" gorm:"->;-:migration"`
	Outputs                   []byte `json:"outputs_json" temporaljson:"outputs_json" gorm:"->;-:migration;type:jsonb" swaggertype:"primitive,string"`

	// read only fields from gorm AfterQuery

	ExecutionTime time.Duration          `json:"execution_time" gorm:"-" swaggertype:"primitive,integer"`
	Execution     *RunnerJobExecution    `json:"-" gorm:"-"`
	ParsedOutputs map[string]interface{} `json:"outputs" gorm:"-"`
}

func (*RunnerJob) UseView() bool {
	return true
}

func (*RunnerJob) ViewVersion() string {
	return "v1"
}

func (i *RunnerJob) Views(db *gorm.DB) []migrations.View {
	return []migrations.View{
		{
			Name: views.DefaultViewName(db, &RunnerJob{}, 1),
			SQL:  viewsql.RunnerJobViewV1,
		},
	}
}

func (r *RunnerJob) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = domains.NewRunnerJobID()
	}

	if r.CreatedByID == "" {
		r.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}

	if r.Group == RunnerJobGroupUnknown {
		r.Group = r.Type.Group()
	}

	if r.OrgID == "" {
		r.OrgID = orgIDFromContext(tx.Statement.Context)
	}

	if r.LogStreamID == nil {
		r.LogStreamID = generics.ToPtr(logstreamIDFromContext(tx.Statement.Context))
	}

	// the overall timeout can be derived by combining the various lower level timeouts.
	if r.OverallTimeout == 0 {
		r.OverallTimeout = r.QueueTimeout + time.Duration(r.MaxExecutions)*(r.AvailableTimeout+r.ExecutionTimeout)
	}

	return nil
}

func (r *RunnerJob) AfterQuery(tx *gorm.DB) error {
	r.ExecutionTime = generics.GetTimeDuration(r.StartedAt, r.FinishedAt)

	if len(r.Outputs) > 0 {
		var outputs map[string]interface{}
		if err := json.Unmarshal(r.Outputs, &outputs); err != nil {
			return errors.Wrap(err, "unable to parse outputs json")
		}
		r.ParsedOutputs = outputs
	}

	return nil
}
