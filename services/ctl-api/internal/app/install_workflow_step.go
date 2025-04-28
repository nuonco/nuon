package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/pkg/shortid/domains"
)

type InstallWorkflowStep struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	// used for RLS
	OrgID string `json:"org_id" gorm:"notnull" swaggerignore:"true" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `json:"-" faker:"-" temporaljson:"org,omitzero,omitempty"`

	Install   Install `swaggerignore:"true" json:"-" temporaljson:"install,omitzero,omitempty"`
	InstallID string  `json:"install_id" gorm:"notnull;default null" temporaljson:"install_id,omitzero,omitempty"`

	InstallWorkflowID string `json:"install_workflow_id" temporaljson:"install_workflow_id,omitzero,omitempty"`

	// status
	Status CompositeStatus `json:"status" temporaljson:"status,omitzero,omitempty"`
	Name   string          `json:"name" temporaljson:"name,omitzero,omitempty"`

	// the signal that needs to be called
	Signal Signal `json:"-" temporaljson:"signal,omitzero,omitempty"`

	Idx int `json:"idx" temporaljson:"idx,omitzero,omitempty"`

	// the following fields are set _once_ a step is in flight, and are orchestrated via the step's signal.
	//
	// this is a polymorphic gorm relationship to one of the following objects:
	//
	// install_cloudformation_stack
	// install_sandbox_run
	// install_runner_update
	// install_deploy
	// install_action_workflow_run (can be many of these)
	StepTargetID   string `json:"step_target_id" gorm:"type:text;check:owner_id_checker,char_length(id)=26" temporaljson:"step_target_id,omitzero,omitempty"`
	StepTargetType string `json:"step_target_type" gorm:"type:text;" temporaljson:"step_target_type,omitzero,omitempty"`

	StartedAt  time.Time `json:"started_at" gorm:"default:null" temporaljson:"started_at,omitzero,omitempty"`
	FinishedAt time.Time `json:"finished_at" gorm:"default:null" temporaljson:"finished_at,omitzero,omitempty"`
	Finished   bool      `json:"finished" gorm:"-" temporaljson:"finished,omitzero,omitempty"`

	// the step approval is built into each step at the runner level.
	Approval         *InstallWorkflowStepApproval         `json:"approval" temporaljson:"approval,omitzero,omitempty"`
	PolicyValidation *InstallWorkflowStepPolicyValidation `json:"policy_validation" temporaljson:"policy_validation,omitzero,omitempty"`

	ExecutionTime time.Duration `json:"execution_time" gorm:"-" swaggertype:"primitive,integer" temporaljson:"execution_time,omitzero,omitempty"`
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
