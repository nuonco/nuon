package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/links"
)

type InstallWorkflowType string

const (
	InstallWorkflowTypeProvision          InstallWorkflowType = "provision"
	InstallWorkflowTypeDeprovision        InstallWorkflowType = "deprovision"
	InstallWorkflowTypeDeprovisionSandbox InstallWorkflowType = "deprovision_sandbox"

	// day-2 triggers
	InstallWorkflowTypeManualDeploy       InstallWorkflowType = "manual_deploy"
	InstallWorkflowTypeInputUpdate        InstallWorkflowType = "input_update"
	InstallWorkflowTypeDeployComponents   InstallWorkflowType = "deploy_components"
	InstallWorkflowTypeTeardownComponents InstallWorkflowType = "teardown_components"
	InstallWorkflowTypeReprovisionSandbox InstallWorkflowType = "reprovision_sandbox"

	// reprovision everything
	InstallWorkflowTypeReprovision InstallWorkflowType = "reprovision"
)

func (i InstallWorkflowType) PastTenseName() string {
	switch i {
	case InstallWorkflowTypeProvision:
		return "Provisioned install"
	case InstallWorkflowTypeReprovision:
		return "Reprovisioned install"
	case InstallWorkflowTypeReprovisionSandbox:
		return "Reprovisioned sandbox"
	case InstallWorkflowTypeDeprovision:
		return "Deprovisioned install"
	case InstallWorkflowTypeManualDeploy:
		return "Deployed to install"
	case InstallWorkflowTypeInputUpdate:
		return "Updated Input"
	case InstallWorkflowTypeTeardownComponents:
		return "Tore down all components"
	case InstallWorkflowTypeDeployComponents:
		return "Deployed all components"
	default:
	}

	return ""
}

func (i InstallWorkflowType) Name() string {
	switch i {
	case InstallWorkflowTypeProvision:
		return "Provisioning install"
	case InstallWorkflowTypeReprovision:
		return "Reprovisioning install"
	case InstallWorkflowTypeDeprovision:
		return "Deprovisioning install"
	case InstallWorkflowTypeManualDeploy:
		return "Deploying to install"
	case InstallWorkflowTypeInputUpdate:
		return "Input Update"
	case InstallWorkflowTypeTeardownComponents:
		return "Tearing down all components"
	case InstallWorkflowTypeDeployComponents:
		return "Deploying all components"
	case InstallWorkflowTypeReprovisionSandbox:
		return "Reprovisioning sandbox"
	default:
	}

	return ""
}

func (i InstallWorkflowType) Description() string {
	switch i {
	case InstallWorkflowTypeProvision:
		return "Creates a runner stack, waits for it to be applied and then provisions the sandbox and deploys all components."
	case InstallWorkflowTypeReprovision:
		return "Creates a new runner stack, waits for it to be applied and then reprovisions the sandbox and deploys all components."
	case InstallWorkflowTypeReprovisionSandbox:
		return "Reprovisions the sandbox and redeploys everything on top of it."
	case InstallWorkflowTypeDeprovision:
		return "Deprovisions all components, deprovisions the sandbox and then waits for the cloudformation stack to be deleted."
	case InstallWorkflowTypeManualDeploy:
		return "Deploys a single component."
	case InstallWorkflowTypeInputUpdate:
		return "Depending on which input was changed, will reprovision the sandbox and deploy one or all components."
	case InstallWorkflowTypeDeployComponents:
		return "Deploy all components in the order of their dependencies."
	case InstallWorkflowTypeTeardownComponents:
		return "Teardown components in the reverse order of their dependencies."
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
type InstallWorkflow struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `gorm:"index:idx_app_install_name,unique" json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	// used for RLS
	OrgID string `json:"org_id" gorm:"notnull" swaggerignore:"true" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `json:"-" faker:"-" temporaljson:"org,omitzero,omitempty"`

	Install   Install `swaggerignore:"true" json:"-" temporaljson:"install,omitzero,omitempty"`
	InstallID string  `json:"install_id" gorm:"notnull;default null" temporaljson:"install_id,omitzero,omitempty"`

	Type              InstallWorkflowType `json:"type" gorm:"not null;default null" temporaljson:"type,omitzero,omitempty"`
	Metadata          pgtype.Hstore       `json:"metadata" gorm:"type:hstore" swaggertype:"object,string" temporaljson:"metadata,omitzero,omitempty"`
	Status            CompositeStatus     `json:"status" temporaljson:"status,omitzero,omitempty"`
	StepErrorBehavior StepErrorBehavior   `json:"step_error_behavior" temporaljson:"step_error_behavior,omitzero,omitempty"`

	StartedAt  time.Time `json:"started_at" gorm:"default:null" temporaljson:"started_at,omitzero,omitempty"`
	FinishedAt time.Time `json:"finished_at" gorm:"default:null" temporaljson:"finished_at,omitzero,omitempty"`
	Finished   bool      `json:"finished" gorm:"-" temporaljson:"finished,omitzero,omitempty"`

	// steps represent each piece of the workflow
	Steps []InstallWorkflowStep `json:"steps" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"steps,omitzero,omitempty"`
	Name  string                `json:"name" gorm:"-" temporaljson:"name,omitzero,omitempty"`

	ExecutionTime time.Duration `json:"execution_time" gorm:"-" swaggertype:"primitive,integer" temporaljson:"execution_time,omitzero,omitempty"`

	InstallSandboxRuns        []InstallSandboxRun        `json:"install_sandbox_runs" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"install_sandbox_runs,omitzero,omitempty"`
	InstallDeploys            []InstallDeploy            `json:"install_deploys" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"install_deploys,omitzero,omitempty"`
	InstallActionWorkflowRuns []InstallActionWorkflowRun `json:"install_action_workflow_runs" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"install_action_runs,omitzero,omitempty"`

	Links map[string]any `json:"links,omitempty" temporaljson:"-" gorm:"-"`
}

func (i *InstallWorkflow) BeforeCreate(tx *gorm.DB) error {
	if i.ID == "" {
		i.ID = domains.NewInstallWorkflowID()
	}

	i.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	i.OrgID = orgIDFromContext(tx.Statement.Context)

	return nil
}

func (r *InstallWorkflow) AfterQuery(tx *gorm.DB) error {
	cfg := configFromContext(tx.Statement.Context)
	if cfg != nil {
		r.Links = links.InstallWorkflowStepLinks(cfg, r.ID)
	}

	r.ExecutionTime = generics.GetTimeDuration(r.StartedAt, r.FinishedAt)
	r.Finished = !r.FinishedAt.IsZero()

	name := r.Type.Name()
	if !r.FinishedAt.IsZero() {
		name = r.Type.PastTenseName()
	}
	r.Name = name

	return nil
}
