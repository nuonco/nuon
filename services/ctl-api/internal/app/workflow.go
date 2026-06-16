package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/links"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
)

type WorkflowType string

const (
	WorkflowTypeProvision          WorkflowType = "provision"
	WorkflowTypeDeprovision        WorkflowType = "deprovision"
	WorkflowTypeDeprovisionSandbox WorkflowType = "deprovision_sandbox"

	// day-2 triggers
	WorkflowTypeManualDeploy               WorkflowType = "manual_deploy"
	WorkflowTypeInputUpdate                WorkflowType = "input_update"
	WorkflowTypeDeployComponents           WorkflowType = "deploy_components"
	WorkflowTypeTeardownComponent          WorkflowType = "teardown_component"
	WorkflowTypeTeardownComponents         WorkflowType = "teardown_components"
	WorkflowTypeReprovisionSandbox         WorkflowType = "reprovision_sandbox"
	WorkflowTypeDriftRunReprovisionSandbox WorkflowType = "drift_run_reprovision_sandbox"
	WorkflowTypeActionWorkflowRun          WorkflowType = "action_workflow_run"
	WorkflowTypeSyncSecrets                WorkflowType = "sync_secrets"
	WorkflowTypeDriftRun                   WorkflowType = "drift_run"

	// app branches workflows
	WorkflowTypeAppBranchesRun                 WorkflowType = "app_branches_manual_update"
	WorkflowTypeAppBranchesConfigRepoUpdate    WorkflowType = "app_branches_config_repo_update"
	WorkflowTypeAppBranchesComponentRepoUpdate WorkflowType = "app_branches_component_repo_update"
	WorkflowTypeAppBranchConfigUpdate          WorkflowType = "app_branch_config_update"

	// reprovision everything
	WorkflowTypeReprovision WorkflowType = "reprovision"

	// app config builds
	WorkflowTypeAppConfigBuild WorkflowType = "app_config_build"

	// runbooks
	WorkflowTypeRunbookRun WorkflowType = "runbook_run"
)

type WorkflowMetadataKey string

const (
	WorkflowMetadataKeyWorkflowNameSuffix = "workflow-name-suffix"
	WorkflowMetadataKeyRole               = "role"
	WorkflowMetadataKeyOwnerName          = "owner_name"
	WorkflowMetadataKeyChangedInputValues = "changed_input_values"
)

func (i WorkflowType) PastTenseName() string {
	switch i {
	case WorkflowTypeProvision:
		return "Provisioned install"
	case WorkflowTypeReprovision:
		return "Reprovisioned install"
	case WorkflowTypeReprovisionSandbox, WorkflowTypeDriftRunReprovisionSandbox:
		return "Reprovisioned sandbox"
	case WorkflowTypeDeprovision:
		return "Deprovisioned install"
	case WorkflowTypeManualDeploy, WorkflowTypeDriftRun:
		return "Deployed to install"
	case WorkflowTypeInputUpdate:
		return "Updated Input"
	case WorkflowTypeTeardownComponents:
		return "Tore down all components"
	case WorkflowTypeDeployComponents:
		return "Deployed all components"
	case WorkflowTypeSyncSecrets:
		return "Synced secrets"
	case WorkflowTypeActionWorkflowRun:
		return "Action run"
	case WorkflowTypeAppConfigBuild:
		return "Built app config components"
	case WorkflowTypeRunbookRun:
		return "Runbook run"
	default:
	}

	return ""
}

func (i WorkflowType) Name() string {
	switch i {
	case WorkflowTypeProvision:
		return "Provisioning install"
	case WorkflowTypeReprovision:
		return "Reprovisioning install"
	case WorkflowTypeDeprovision:
		return "Deprovisioning install"
	case WorkflowTypeManualDeploy, WorkflowTypeDriftRun:
		return "Deploying to install"
	case WorkflowTypeInputUpdate:
		return "Input Update"
	case WorkflowTypeTeardownComponents:
		return "Tearing down all components"
	case WorkflowTypeDeployComponents:
		return "Deploying all components"
	case WorkflowTypeReprovisionSandbox, WorkflowTypeDriftRunReprovisionSandbox:
		return "Reprovisioning sandbox"
	case WorkflowTypeSyncSecrets:
		return "Syncing secrets"
	case WorkflowTypeActionWorkflowRun:
		return "Action run"
	case WorkflowTypeAppConfigBuild:
		return "Building app config components"
	case WorkflowTypeRunbookRun:
		return "Running runbook"
	default:
	}

	return ""
}

func (i WorkflowType) Description() string {
	switch i {
	case WorkflowTypeProvision:
		return "Creates a runner stack, waits for it to be applied and then provisions the sandbox and deploys all components."
	case WorkflowTypeReprovision:
		return "Creates a new runner stack, waits for it to be applied and then reprovisions the sandbox and deploys all components."
	case WorkflowTypeReprovisionSandbox, WorkflowTypeDriftRunReprovisionSandbox:
		return "Reprovisions the sandbox and redeploys everything on top of it."
	case WorkflowTypeDeprovision:
		return "Deprovisions all components, deprovisions the sandbox and then waits for the cloudformation stack to be deleted."
	case WorkflowTypeManualDeploy, WorkflowTypeActionWorkflowRun:
		return "Deploys a single component."
	case WorkflowTypeInputUpdate:
		return "Depending on which input was changed, will reprovision the sandbox and deploy one or all components."
	case WorkflowTypeDeployComponents:
		return "Deploy all components in the order of their dependencies."
	case WorkflowTypeTeardownComponents:
		return "Teardown components in the reverse order of their dependencies."
	case WorkflowTypeSyncSecrets:
		return "Syncing customer secrets into cluster."
	case WorkflowTypeAppConfigBuild:
		return "Builds all components defined in an app config."
	case WorkflowTypeRunbookRun:
		return "Executes an ordered set of deploy and action steps defined in a runbook."
	}

	return "unknown"
}

// DEPRECATED: this is no longer used, but kept for historical data integrity
type StepErrorBehavior string

const (
	// abort on error
	StepErrorBehaviorAbort StepErrorBehavior = "abort"

	// continue on error
	// DEPRECATED: this is no longer used, but kept for historical data integrity
	// StepErrorBehaviorContinue StepErrorBehavior = "continue"
)

// GenerateStepsResult holds the output of a workflow step generator.
type GenerateStepsResult struct {
	Steps  []*WorkflowStep      `json:"steps"`
	Groups []*WorkflowStepGroup `json:"groups"`
}

// TODO(jm): make install workflows a top level concept called a "workflow", and they belong to either an app or an
// install.
//
// We start with this to make it easier to iterate on them, for now.
type Workflow struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"created_by" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	OwnerID   string `json:"owner_id,omitzero" gorm:"type:text;check:owner_id_checker,char_length(id)=26" temporaljson:"owner_id,omitzero,omitempty"`
	OwnerType string `json:"owner_type,omitzero" gorm:"type:text;" temporaljson:"owner_type,omitzero,omitempty"`
	// OwnerName is a derived, non-persisted convenience field. It is
	// populated by activities that need a human-readable owner label
	// (e.g. workflow lifecycle webhooks) via a small switch on OwnerType
	// — see PkgWorkflowsFlowGetFlow. Empty unless the loading path
	// explicitly fills it.
	OwnerName string `json:"owner_name,omitzero" gorm:"-" temporaljson:"owner_name,omitzero,omitempty"`

	// used for RLS
	OrgID string `json:"org_id,omitzero" gorm:"notnull" swaggerignore:"true" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `json:"-" faker:"-" temporaljson:"org,omitzero,omitempty"`

	Type     WorkflowType    `json:"type,omitzero" gorm:"not null;default null" temporaljson:"type,omitzero,omitempty"`
	Metadata pgtype.Hstore   `json:"metadata,omitzero" gorm:"type:hstore" swaggertype:"object,string" temporaljson:"metadata,omitzero,omitempty"`
	Status   CompositeStatus `json:"status,omitzero" temporaljson:"status,omitzero,omitempty"`

	Role string `json:"role,omitzero,omitempty" temporaljson:"role" gorm:"column:role"`

	// DEPRECATED: for now we always abort on step errors
	StepErrorBehavior StepErrorBehavior `json:"step_error_behavior,omitzero" temporaljson:"step_error_behavior,omitzero,omitempty" swaggertype:"string"`

	ApprovalOption InstallApprovalOption `json:"approval_option,omitzero" gorm:"default 'prompt'" temporaljson:"approval_option,omitzero,omitempty"`

	PlanOnly bool `json:"plan_only"`

	// GenerateStepsSignal is an optional queue signal that generates workflow steps.
	// When set, the conductor enqueues this signal and calls its "FetchSteps" update
	// handler instead of using the hardcoded Generators map.
	GenerateStepsSignal *signaldb.SignalData `json:"generate_steps_signal,omitempty" gorm:"type:jsonb"`

	// ResultDirective is set by the currently executing group signal to communicate
	// the group outcome back to the flow signal. Values: continue, stop, retry-group,
	// skip-group, await-approval.
	ResultDirective string `json:"result_directive,omitzero" gorm:"type:text;default:''" temporaljson:"result_directive,omitzero,omitempty"`

	StartedAt  time.Time `json:"started_at,omitzero" gorm:"default:null" temporaljson:"started_at,omitzero,omitempty"`
	FinishedAt time.Time `json:"finished_at,omitzero" gorm:"default:null" temporaljson:"finished_at,omitzero,omitempty"`
	Finished   bool      `json:"finished,omitzero" gorm:"-" temporaljson:"finished,omitzero,omitempty"`

	// step groups represent logical groupings of steps within the workflow
	StepGroups []WorkflowStepGroup `json:"step_groups,omitzero" gorm:"foreignKey:WorkflowID;constraint:OnDelete:CASCADE;" temporaljson:"step_groups,omitzero,omitempty"`

	// steps represent each piece of the workflow
	Steps []WorkflowStep `json:"steps,omitzero" gorm:"foreignKey:InstallWorkflowID;constraint:OnDelete:CASCADE;" temporaljson:"steps,omitzero,omitempty"`
	// Name is the human-readable workflow title shown in the UI (e.g.
	// "Deploying to install (rds_cluster_temporal)"). Populated by
	// BeforeSave via computeWorkflowName — callers that mutate Type,
	// Metadata, or FinishedAt must go through GORM (Save / struct-based
	// Updates) so the hook fires.
	Name string `json:"name,omitzero" gorm:"column:name;index" temporaljson:"name,omitzero,omitempty"`

	ExecutionTime time.Duration `json:"execution_time,omitzero" gorm:"-" swaggertype:"primitive,integer" temporaljson:"execution_time,omitzero,omitempty"`

	InstallSandboxRuns        []InstallSandboxRun        `json:"install_sandbox_runs,omitzero" gorm:"foreignKey:InstallWorkflowID;constraint:OnDelete:CASCADE;" temporaljson:"install_sandbox_runs,omitzero,omitempty"`
	InstallDeploys            []InstallDeploy            `json:"install_deploys,omitzero" gorm:"foreignKey:InstallWorkflowID;constraint:OnDelete:CASCADE;" temporaljson:"install_deploys,omitzero,omitempty"`
	InstallActionWorkflowRuns []InstallActionWorkflowRun `json:"install_action_workflow_runs,omitzero" gorm:"foreignKey:InstallWorkflowID;constraint:OnDelete:CASCADE;" temporaljson:"install_action_runs,omitzero,omitempty"`
	AppBranchRuns             []AppBranchRun             `json:"app_branch_runs,omitzero" gorm:"foreignKey:WorkflowID;constraint:OnDelete:CASCADE;" temporaljson:"app_branch_runs,omitzero,omitempty"`
	WorkflowRuns              []WorkflowRun              `json:"workflow_runs,omitzero" gorm:"foreignKey:WorkflowID;constraint:OnDelete:CASCADE;" temporaljson:"workflow_runs,omitzero,omitempty"`

	Links map[string]any `json:"links,omitzero,omitempty" temporaljson:"-" gorm:"-"`
}

func (i *Workflow) TableName() string {
	// Workflows used to be called InstallWorkflows
	return "install_workflows"
}

func (i *Workflow) BeforeCreate(tx *gorm.DB) error {
	if i.ID == "" {
		i.ID = domains.NewWorkflowID()
	}

	i.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	i.OrgID = orgIDFromContext(tx.Statement.Context)

	return nil
}

// BeforeSave keeps install_workflows.name in sync with computeWorkflowName.
// Callers that mutate type, metadata, or finished_at must pass a fully
// populated struct (load-then-Save) so this hook can see the post-update
// state. Partial Updates(map / struct-with-only-finished_at) and raw SQL
// bypass the hook's view and will leave name stale.
func (i *Workflow) BeforeSave(tx *gorm.DB) error {
	i.Name = computeWorkflowName(i)
	return nil
}

func (i *Workflow) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: "idx_install_workflows_owner",
			Columns: []string{
				"owner_id",
				"created_at DESC",
			},
		},
		{
			Name: indexes.Name(db, &Workflow{}, "org_id"),
			Columns: []string{
				"org_id",
			},
		},
		{
			Name: "idx_install_workflows_org_created_at",
			Columns: []string{
				"org_id",
				"created_at DESC",
			},
		},
	}
}

func (r *Workflow) AfterQuery(tx *gorm.DB) error {
	r.Links = links.InstallWorkflowLinks(tx.Statement.Context, r.ID)

	r.ExecutionTime = generics.GetTimeDuration(r.StartedAt, r.FinishedAt)
	r.Finished = !r.FinishedAt.IsZero()

	return nil
}
