package app

import (
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/migrations"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/links"
)

type WorkflowType string

const (
	WorkflowTypeProvision          WorkflowType = "provision"
	WorkflowTypeDeprovision        WorkflowType = "deprovision"
	WorkflowTypeDeprovisionSandbox WorkflowType = "deprovision_sandbox"

	// day-2 triggers
	WorkflowTypeManualDeploy       WorkflowType = "manual_deploy"
	WorkflowTypeInputUpdate        WorkflowType = "input_update"
	WorkflowTypeDeployComponents   WorkflowType = "deploy_components"
	WorkflowTypeTeardownComponent  WorkflowType = "teardown_component"
	WorkflowTypeTeardownComponents WorkflowType = "teardown_components"
	WorkflowTypeReprovisionSandbox WorkflowType = "reprovision_sandbox"
	WorkflowTypeActionWorkflowRun  WorkflowType = "action_workflow_run"
	WorkflowTypeSyncSecrets        WorkflowType = "sync_secrets"
	WorkflowTypeDriftRun           WorkflowType = "drift_run"

	// app branches workflows
	WorkflowTypeAppBranchesManualUpdate        WorkflowType = "app_branches_manual_update"
	WorkflowTypeAppBranchesConfigRepoUpdate    WorkflowType = "app_branches_config_repo_update"
	WorkflowTypeAppBranchesComponentRepoUpdate WorkflowType = "app_branches_component_repo_update"

	// reprovision everything
	WorkflowTypeReprovision WorkflowType = "reprovision"
)

type WorkflowMetadataKey string

const (
	WorkflowMetadataKeyWorkflowNameSuffix = "workflow-name-suffix"
)

func (i WorkflowType) PastTenseName() string {
	switch i {
	case WorkflowTypeProvision:
		return "Provisioned install"
	case WorkflowTypeReprovision:
		return "Reprovisioned install"
	case WorkflowTypeReprovisionSandbox:
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
	case WorkflowTypeReprovisionSandbox:
		return "Reprovisioning sandbox"
	case WorkflowTypeSyncSecrets:
		return "Syncing secrets"
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
	case WorkflowTypeReprovisionSandbox:
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
	}

	return "unknown"
}

type StepErrorBehavior string

const (
	// abort on error
	StepErrorBehaviorAbort StepErrorBehavior = "abort"

	// continue on error
	StepErrorBehaviorContinue StepErrorBehavior = "continue"
)

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
	DeletedAt   soft_delete.DeletedAt `gorm:"index:idx_app_install_name,unique" json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	OwnerID   string `json:"owner_id,omitzero" gorm:"type:text;check:owner_id_checker,char_length(id)=26" temporaljson:"owner_id,omitzero,omitempty"`
	OwnerType string `json:"owner_type,omitzero" gorm:"type:text;" temporaljson:"owner_type,omitzero,omitempty"`

	// used for RLS
	OrgID string `json:"org_id,omitzero" gorm:"notnull" swaggerignore:"true" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `json:"-" faker:"-" temporaljson:"org,omitzero,omitempty"`

	Type              WorkflowType      `json:"type,omitzero" gorm:"not null;default null" temporaljson:"type,omitzero,omitempty"`
	Metadata          pgtype.Hstore     `json:"metadata,omitzero" gorm:"type:hstore" swaggertype:"object,string" temporaljson:"metadata,omitzero,omitempty"`
	Status            CompositeStatus   `json:"status,omitzero" temporaljson:"status,omitzero,omitempty"`
	StepErrorBehavior StepErrorBehavior `json:"step_error_behavior,omitzero" temporaljson:"step_error_behavior,omitzero,omitempty"`

	ApprovalOption InstallApprovalOption `json:"approval_option,omitzero" gorm:"default 'auto'" temporaljson:"approval_option,omitzero,omitempty"`

	PlanOnly bool `json:"plan_only"`

	StartedAt  time.Time `json:"started_at,omitzero" gorm:"default:null" temporaljson:"started_at,omitzero,omitempty"`
	FinishedAt time.Time `json:"finished_at,omitzero" gorm:"default:null" temporaljson:"finished_at,omitzero,omitempty"`
	Finished   bool      `json:"finished,omitzero" gorm:"-" temporaljson:"finished,omitzero,omitempty"`

	// steps represent each piece of the workflow
	Steps []WorkflowStep `json:"steps,omitzero" gorm:"foreignKey:InstallWorkflowID;constraint:OnDelete:CASCADE;" temporaljson:"steps,omitzero,omitempty"`
	Name  string         `json:"name,omitzero" gorm:"-" temporaljson:"name,omitzero,omitempty"`

	ExecutionTime time.Duration `json:"execution_time,omitzero" gorm:"-" swaggertype:"primitive,integer" temporaljson:"execution_time,omitzero,omitempty"`

	InstallSandboxRuns        []InstallSandboxRun        `json:"install_sandbox_runs,omitzero" gorm:"foreignKey:InstallWorkflowID;constraint:OnDelete:CASCADE;" temporaljson:"install_sandbox_runs,omitzero,omitempty"`
	InstallDeploys            []InstallDeploy            `json:"install_deploys,omitzero" gorm:"foreignKey:InstallWorkflowID;constraint:OnDelete:CASCADE;" temporaljson:"install_deploys,omitzero,omitempty"`
	InstallActionWorkflowRuns []InstallActionWorkflowRun `json:"install_action_workflow_runs,omitzero" gorm:"foreignKey:InstallWorkflowID;constraint:OnDelete:CASCADE;" temporaljson:"install_action_runs,omitzero,omitempty"`

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

func (i *Workflow) BeforeSave(tx *gorm.DB) error {
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
	}
}

func (r *Workflow) AfterQuery(tx *gorm.DB) error {
	r.Links = links.InstallWorkflowLinks(tx.Statement.Context, r.ID)

	r.ExecutionTime = generics.GetTimeDuration(r.StartedAt, r.FinishedAt)
	r.Finished = !r.FinishedAt.IsZero()

	name := r.Type.Name()
	if !r.FinishedAt.IsZero() {
		name = r.Type.PastTenseName()
	}
	r.Name = name
	if component_name, ok := r.Metadata[WorkflowMetadataKeyWorkflowNameSuffix]; ok {
		r.Name = fmt.Sprintf("%s (%s)", r.Name, generics.FromPtrStr(component_name))
	}

	return nil
}
