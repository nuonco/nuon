package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/pkg/shortid/domains"
)

type InstallWorkflowStep struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string                `json:"created_by_id" gorm:"not null;default:null"`
	CreatedBy   Account               `json:"-" temporaljson:"-"`
	CreatedAt   time.Time             `json:"created_at" gorm:"notnull"`
	UpdatedAt   time.Time             `json:"updated_at" gorm:"notnull"`
	DeletedAt   soft_delete.DeletedAt `json:"-"`

	// used for RLS
	OrgID string `json:"org_id" gorm:"notnull" swaggerignore:"true"`
	Org   Org    `json:"-" faker:"-"`

	Install   Install `swaggerignore:"true" json:"-"`
	InstallID string  `json:"install_id" gorm:"notnull;default null"`

	InstallWorkflowID string `json:"install_workflow_id"`

	// status
	Status CompositeStatus `json:"status"`
	Name   string          `json:"name"`

	// the signal that needs to be called
	Signal Signal `json:"-" temporaljson:"signal"`

	Idx int `json:"idx"`

	// the following fields are set _once_ a step is in flight, and are orchestrated via the step's signal.
	//
	// this is a polymorphic gorm relationship to one of the following objects:
	//
	// install_cloudformation_stack
	// install_sandbox_run
	// install_runner_update
	// install_deploy
	// install_action_workflow_run (can be many of these)
	StepTargetID   string `json:"step_target_id" gorm:"type:text;check:owner_id_checker,char_length(id)=26"`
	StepTargetType string `json:"step_target_type" gorm:"type:text;"`

	StartedAt  time.Time `json:"started_at"  gorm:"default:null"`
	FinishedAt time.Time `json:"finished_at" gorm:"default:null"`
  Finished   bool      `json:"finished" gorm:"-"`

	// the step approval is built into each step at the runner level.
	Approval         *InstallWorkflowStepApproval         `json:"approval"`
	PolicyValidation *InstallWorkflowStepPolicyValidation `json:"policy_validation"`

	ExecutionTime time.Duration `json:"execution_time" gorm:"-" swaggertype:"primitive,integer"`
}

func (a *InstallWorkflowStep) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = domains.NewInstallWorkflowStepID()
	}

	if a.CreatedByID == "" {
		a.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}

	if a.OrgID == "" {
		a.OrgID = orgIDFromContext(tx.Statement.Context)
	}
	return nil
}

func (r *InstallWorkflowStep) AfterQuery(tx *gorm.DB) error {
	r.ExecutionTime = generics.GetTimeDuration(r.StartedAt, r.FinishedAt)
	r.Finished = !r.FinishedAt.IsZero()
	return nil
}
