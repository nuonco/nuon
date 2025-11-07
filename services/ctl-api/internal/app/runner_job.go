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
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/indexes"
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
	RunnerJobGroupManagement RunnerJobGroup = "management"

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
	RunnerJobTypeTerraformDeploy          RunnerJobType = "terraform-deploy"
	RunnerJobTypeHelmChartDeploy          RunnerJobType = "helm-chart-deploy"
	RunnerJobTypeJobDeploy                RunnerJobType = "job-deploy"
	RunnerJobTypeKubrenetesManifestDeploy RunnerJobType = "kubernetes-manifest-deploy"
	RunnerJobTypeJobNOOPDeploy            RunnerJobType = "noop-deploy"

	// operations job types
	RunnerJobTypeShutDown      RunnerJobType = "shut-down"
	RunnerJobTypeUpdateVersion RunnerJobType = "update-version"
	RunnerJobTypeNOOP          RunnerJobType = "noop"

	// TODO(fd): revisit these names
	// management job types
	// RunnerJobTypeMngVMStats             RunnerJobType = "mng-vm-stats"          // log some vm stats/metrics
	RunnerJobTypeMngVMShutDown          RunnerJobType = "mng-vm-shut-down"          // shut down the vm
	RunnerJobTypeMngShutDown            RunnerJobType = "mng-shut-down"             // shutdown the runner mng process (usually triggers restart)
	RunnerJobTypeMngRunnerUpdateVersion RunnerJobType = "mng-runner-update-version" // update the runner image/version (check for changes and update)
	RunnerJobTypeMngRunnerRestart       RunnerJobType = "mng-runner-restart"        // restart the runner systemctl service (technically, a duplicate. runner can restart self.)

	// sandbox job types
	RunnerJobTypeSandboxTerraform     RunnerJobType = "sandbox-terraform"
	RunnerJobTypeSandboxTerraformPlan RunnerJobType = "sandbox-terraform-plan"
	RunnerJobTypeSandboxSyncSecrets   RunnerJobType = "sandbox-sync-secrets"

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
		RunnerJobTypeKubrenetesManifestDeploy,
		RunnerJobTypeJobNOOPDeploy:
		return RunnerJobGroupDeploy

		// runners
	case RunnerJobTypeRunnerHelm, RunnerJobTypeRunnerTerraform:
		return RunnerJobGroupRunner

		// sandboxes
	case RunnerJobTypeSandboxTerraform,
		RunnerJobTypeSandboxTerraformPlan,
		RunnerJobTypeSandboxSyncSecrets:
		return RunnerJobGroupSandbox

		// health checks
	case RunnerJobTypeHealthCheck:
		return RunnerJobGroupHealthChecks

		// operations
	case RunnerJobTypeNOOP, RunnerJobTypeShutDown, RunnerJobTypeUpdateVersion:
		return RunnerJobGroupOperations

		// management
	case RunnerJobTypeMngVMShutDown, RunnerJobTypeMngShutDown, RunnerJobTypeMngRunnerUpdateVersion, RunnerJobTypeMngRunnerRestart:
		return RunnerJobGroupManagement

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

	// update build
	RunnerJobOperationTypeBuild RunnerJobOperationType = "build"

	// the following operations are for common use cases for things such as helm, terraform and other jobs that have
	// multiple operation types.
	RunnerJobOperationTypeCreateApplyPlan    RunnerJobOperationType = "create-apply-plan"
	RunnerJobOperationTypeCreateTeardownPlan RunnerJobOperationType = "create-teardown-plan"
	RunnerJobOperationTypeApplyPlan          RunnerJobOperationType = "apply-plan"

	RunnerJobOperationTypeUnknown RunnerJobOperationType = "unknown"
)

type RunnerJob struct {
	ID          string  `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string  `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account `json:"-" temporaljson:"created_by,omitzero,omitempty"`

	CreatedAt time.Time             `json:"created_at,omitzero" gorm:"notnull;index:idx_runner_jobs_query,priority:4,sort:desc;index:idx_runner_jobs_owner_id,priority:2,sort:desc" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt time.Time             `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt soft_delete.DeletedAt `json:"-" gorm:"index:idx_runner_name,unique;" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgID string `json:"org_id,omitzero" gorm:"index:idx_app_name,unique" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `json:"-" temporaljson:"org,omitzero,omitempty"`

	RunnerID    string  `json:"runner_id,omitzero" gorm:"index:idx_runner_name,unique;index:idx_runner_jobs_query,priority:1" temporaljson:"runner_id,omitzero,omitempty"`
	OwnerID     string  `json:"owner_id,omitzero" gorm:"type:text;check:owner_id_checker,char_length(id)=26;index:idx_runner_jobs_owner_id,priority:1" temporaljson:"owner_id,omitzero,omitempty"`
	OwnerType   string  `json:"owner_type,omitzero" gorm:"type:text;" temporaljson:"owner_type,omitzero,omitempty"`
	LogStreamID *string `json:"log_stream_id,omitzero" temporaljson:"log_stream_id,omitzero,omitempty"`

	// queue timeout is how long a job can be queued, before being made available
	QueueTimeout time.Duration `json:"queue_timeout,omitzero" gorm:"default null;not null" swaggertype:"primitive,integer" temporaljson:"queue_timeout,omitzero,omitempty"`
	// available timeout is how long a job can be marked as "available" before being requeued
	AvailableTimeout time.Duration `json:"available_timeout,omitzero" gorm:"default null;not null" swaggertype:"primitive,integer" temporaljson:"available_timeout,omitzero,omitempty"`
	// execution timeout is how long a job can be marked as "exeucuting" before being requeued
	ExecutionTimeout time.Duration `json:"execution_timeout,omitzero" gorm:"default null;not null" swaggertype:"primitive,integer" temporaljson:"execution_timeout,omitzero,omitempty"`

	// overall timeout is how long a job can be attempted, before being cancelled
	OverallTimeout time.Duration `json:"overall_timeout,omitzero" gorm:"default null;not null" swaggertype:"primitive,integer" temporaljson:"overall_timeout,omitzero,omitempty"`

	MaxExecutions int `json:"max_executions,omitzero" gorm:"not null;default null" temporaljson:"max_executions,omitzero,omitempty"`

	Status            RunnerJobStatus `json:"status,omitzero" gorm:"not null;default null;index:idx_runner_jobs_query,priority:3" temporaljson:"status,omitzero,omitempty"`
	StatusDescription string          `json:"status_description,omitzero" gorm:"not null;default null" temporaljson:"status_description,omitzero,omitempty"`

	Type      RunnerJobType          `json:"type,omitzero" gorm:"default null;not null" temporaljson:"type,omitzero,omitempty"`
	Group     RunnerJobGroup         `json:"group,omitzero" gorm:"default:null;not null;index:idx_runner_jobs_query,priority:2" temporaljson:"group,omitzero,omitempty"`
	Operation RunnerJobOperationType `json:"operation,omitzero" gorm:"default:null;not null" temporaljson:"operation,omitzero,omitempty"`

	Executions []RunnerJobExecution `json:"executions,omitzero" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"executions,omitzero,omitempty"`
	Plan       RunnerJobPlan        `json:"-" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"plan,omitzero,omitempty"`

	StartedAt  time.Time `json:"started_at,omitzero" gorm:"default:null" temporaljson:"started_at,omitzero,omitempty"`
	FinishedAt time.Time `json:"finished_at,omitzero" gorm:"default:null" temporaljson:"finished_at,omitzero,omitempty"`

	Metadata pgtype.Hstore `json:"metadata,omitzero" gorm:"type:hstore" swaggertype:"object,string" temporaljson:"metadata,omitzero,omitempty"`

	// read only fields from view

	ExecutionCount            int    `json:"execution_count,omitzero" gorm:"->;-:migration" temporaljson:"execution_count,omitzero,omitempty"`
	FinalRunnerJobExecutionID string `json:"final_runner_job_execution_id,omitzero" gorm:"->;-:migration" temporaljson:"final_runner_job_execution_id,omitzero,omitempty"`
	Outputs                   []byte `json:"outputs_json,omitzero" gorm:"->;-:migration;type:jsonb" swaggertype:"primitive,string" temporaljson:"outputs,omitzero,omitempty"`

	// read only fields from gorm AfterQuery

	ExecutionTime time.Duration          `json:"execution_time,omitzero" gorm:"-" swaggertype:"primitive,integer" temporaljson:"execution_time,omitzero,omitempty"`
	Execution     *RunnerJobExecution    `json:"-" gorm:"-" temporaljson:"execution,omitzero,omitempty"`
	ParsedOutputs map[string]interface{} `json:"outputs,omitzero" gorm:"-" temporaljson:"parsed_outputs,omitzero,omitempty"`

	// foreign keys
	States      []TerraformWorkspaceState     `faker:"-" json:"-" swaggerignore:"true" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"states,omitzero,omitempty"`
	LockHistory []TerraformWorkspaceLock      `faker:"-" json:"-" swaggerignore:"true" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"lock_history,omitzero,omitempty"`
	StateJSON   []TerraformWorkspaceStateJSON `json:"-" swaggerignore:"true" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"statesjson,omitzero,omitempty"`
}

func (*RunnerJob) UseView() bool {
	return true
}

func (*RunnerJob) ViewVersion() string {
	return "v2"
}

func (i *RunnerJob) Views(db *gorm.DB) []migrations.View {
	return []migrations.View{
		{
			Name: views.DefaultViewName(db, &RunnerJob{}, 1),
			SQL:  viewsql.RunnerJobViewV1,
		},
		{
			Name: views.DefaultViewName(db, &RunnerJob{}, 2),
			SQL:  viewsql.RunnerJobViewV2,
		},
	}
}

func (a *RunnerJob) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &RunnerJob{}, "find_type"),
			Columns: []string{
				"runner_id",
				"type",
				"deleted_at",
				"created_at DESC",
			},
		},
		{
			Name: indexes.Name(db, &RunnerJob{}, "org_id"),
			Columns: []string{
				"org_id",
			},
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

	if len(r.Executions) > 0 {
		r.Execution = &r.Executions[0]
	}

	return nil
}
