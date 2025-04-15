package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
)

type InstallWorkflowType string

const (
	InstallWorkflowTypeProvision   InstallWorkflowType = "provision"
	InstallWorkflowTypeDeprovision InstallWorkflowType = "deprovision"

	// day-2 triggers
	InstallWorkflowTypeManualDeploy       InstallWorkflowType = "manual_deploy"
	InstallWorkflowTypeInputUpdate        InstallWorkflowType = "input_update"
	InstallWorkflowTypeDeployComponents   InstallWorkflowType = "deploy_components"
	InstallWorkflowTypeTeardownComponents InstallWorkflowType = "teardown_components"
	InstallWorkflowTypeReprovision        InstallWorkflowType = "reprovision"
)

func (i InstallWorkflowType) Description() string {
	switch i {
	case InstallWorkflowTypeProvision:
		return "Creates a runner stack, waits for it to be applied and then provisions the sandbox and deploys all components."
	case InstallWorkflowTypeReprovision:
		return "Creates a new runner stack, waits for it to be applied and then reprovisions the sandbox and deploys all components."
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
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string                `json:"created_by_id" gorm:"not null;default:null"`
	CreatedBy   Account               `json:"-"`
	CreatedAt   time.Time             `json:"created_at" gorm:"notnull"`
	UpdatedAt   time.Time             `json:"updated_at" gorm:"notnull"`
	DeletedAt   soft_delete.DeletedAt `gorm:"index:idx_app_install_name,unique" json:"-"`

	// used for RLS
	OrgID string `json:"org_id" gorm:"notnull" swaggerignore:"true"`
	Org   Org    `json:"-" faker:"-"`

	Install   Install `swaggerignore:"true" json:"-"`
	InstallID string  `json:"install_id" gorm:"notnull;default null"`

	Type              InstallWorkflowType `json:"type" gorm:"not null;default null"`
	Metadata          pgtype.Hstore       `json:"metadata" gorm:"type:hstore" swaggertype:"object,string"`
	Status            CompositeStatus     `json:"status"`
	StepErrorBehavior StepErrorBehavior   `json:"step_error_behavior"`

	// steps represent each piece of the workflow
	Steps []InstallWorkflowStep `json:"steps" gorm:"constraint:OnDelete:CASCADE;"`
}

func (i *InstallWorkflow) BeforeCreate(tx *gorm.DB) error {
	if i.ID == "" {
		i.ID = domains.NewInstallWorkflowID()
	}

	i.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	i.OrgID = orgIDFromContext(tx.Statement.Context)

	return nil
}
