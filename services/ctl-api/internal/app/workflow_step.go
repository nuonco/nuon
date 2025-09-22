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

type WorkflowStepExecutionType string

const (
	WorkflowStepExecutionTypeSystem   WorkflowStepExecutionType = "system"
	WorkflowStepExecutionTypeUser     WorkflowStepExecutionType = "user"
	WorkflowStepExecutionTypeApproval WorkflowStepExecutionType = "approval"
	WorkflowStepExecutionTypeSkipped  WorkflowStepExecutionType = "skipped"
	WorkflowStepExecutionTypeHidden   WorkflowStepExecutionType = "hidden"
)

type WorkflowStep struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"created_by" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	// used for RLS
	OrgID string `json:"org_id,omitzero" gorm:"notnull" swaggerignore:"true" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `json:"-" faker:"-" temporaljson:"org,omitzero,omitempty"`

	OwnerID   string `json:"owner_id,omitzero" gorm:"type:text;check:owner_id_checker,char_length(id)=26;index:idx_install_workflows_owner_id,priority:1" temporaljson:"owner_id,omitzero,omitempty"`
	OwnerType string `json:"owner_type,omitzero" gorm:"type:text;" temporaljson:"owner_type,omitzero,omitempty"`

	// DEPRECATED: this is the install workflow ID, which is now the workflow ID.
	InstallWorkflowID string `json:"install_workflow_id,omitzero" temporaljson:"install_workflow_id,omitzero,omitempty"`

	Status CompositeStatus `json:"status,omitzero" temporaljson:"status,omitzero,omitempty"`
	Name   string          `json:"name,omitzero" temporaljson:"name,omitzero,omitempty"`

	// the signal that needs to be called
	Signal Signal `json:"-" temporaljson:"signal,omitzero,omitempty"`

	Idx int `json:"idx,omitzero" temporaljson:"idx,omitzero,omitempty"`

	// to group steps which belong to same logical group, eg, plan/apply
	GroupIdx int `json:"group_idx,omitzero" temporaljson:"group_idx,omitzero,omitempty"`
	// counter for every retry attempted on a group
	GroupRetryIdx int `json:"group_retry_idx" gorm:"default:0" temporaljson:"group_retry_idx,omitzero,omitempty"`

	ExecutionType WorkflowStepExecutionType `json:"execution_type,omitzero" temporaljson:"execution_type"`

	// the following fields are set _once_ a step is in flight, and are orchestrated via the step's signal.
	//
	// this is a polymorphic gorm relationship to one of the following objects:
	//
	// install_cloudformation_stack
	// install_sandbox_run
	// install_runner_update
	// install_deploy
	// install_action_workflow_run (can be many of these)
	StepTargetID   string `json:"step_target_id,omitzero" gorm:"type:text;check:owner_id_checker,char_length(id)=26" temporaljson:"step_target_id,omitzero,omitempty"`
	StepTargetType string `json:"step_target_type,omitzero" gorm:"type:text;" temporaljson:"step_target_type,omitzero,omitempty"`

	Metadata pgtype.Hstore `json:"metadata,omitzero" gorm:"type:hstore" swaggertype:"object,string" temporaljson:"metadata,omitzero,omitempty"`

	StartedAt  time.Time `json:"started_at,omitzero" gorm:"default:null" temporaljson:"started_at,omitzero,omitempty"`
	FinishedAt time.Time `json:"finished_at,omitzero" gorm:"default:null" temporaljson:"finished_at,omitzero,omitempty"`
	Finished   bool      `json:"finished,omitzero" gorm:"-" temporaljson:"finished,omitzero,omitempty"`

	// the step approval is built into each step at the runner level.

	Approval         *WorkflowStepApproval         `gorm:"foreignKey:InstallWorkflowStepID" json:"approval,omitzero" temporaljson:"approval,omitzero,omitempty"`
	PolicyValidation *WorkflowStepPolicyValidation `gorm:"foreignKey:InstallWorkflowStepID" json:"policy_validation,omitzero" temporaljson:"policy_validation,omitzero,omitempty"`

	ExecutionTime time.Duration `json:"execution_time,omitzero" gorm:"-" swaggertype:"primitive,integer" temporaljson:"execution_time,omitzero,omitempty"`

	Links map[string]any `json:"links,omitzero,omitempty" temporaljson:"-" gorm:"-"`

	Retryable bool `json:"retryable,omitzero" gorm:"default:false" temporaljson:"retryable,omitzero,omitempty"`
	Skippable bool `json:"skippable,omitzero" gorm:"default:false" temporaljson:"skippable,omitzero,omitempty"`
	Retried   bool `json:"retried,omitzero" gorm:"default:false" temporaljson:"retried,omitzero,omitempty"`

	// Fields that are de-nested at read time using AfterQuery
	WorkflowID string `json:"workflow_id,omitzero" gorm:"-" temporaljson:"workflow_id,omitzero,omitempty"`
}

func (i *WorkflowStep) TableName() string {
	// WorkflowStep used to be called InstallWorkflowStep
	return "install_workflow_steps"
}

func (i *WorkflowStep) BeforeSave(tx *gorm.DB) error {
	return nil
}

func (a *WorkflowStep) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = domains.NewWorkflowStepID()
	}

	if a.CreatedByID == "" {
		a.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}

	if a.OrgID == "" {
		a.OrgID = orgIDFromContext(tx.Statement.Context)
	}
	return nil
}

func (r *WorkflowStep) AfterQuery(tx *gorm.DB) error {
	r.Links = links.InstallWorkflowStepLinks(tx.Statement.Context, r.ID)

	r.ExecutionTime = generics.GetTimeDuration(r.StartedAt, r.FinishedAt)
	r.Finished = !r.FinishedAt.IsZero()

	r.WorkflowID = r.InstallWorkflowID
	return nil
}

// Dual-write install workflows to flows.
// func (i *InstallWorkflowStep) AfterCreate(tx *gorm.DB) error {
// 	return errors.Wrap(tx.Create(i.TransformToFlowStep()).Error, "failed dual-write to flows after create")
// }

// func (i *InstallWorkflowStep) AfterUpdate(tx *gorm.DB) error {
// 	fls := i.TransformToFlowStep()
// 	return errors.Wrap(tx.Model(&FlowStep{
// 		ID: fls.ID,
// 	}).Updates(fls).Error, "failed dual-write to flows after save")
// }

// func (i *InstallWorkflowStep) AfterDelete(tx *gorm.DB) error {
// 	return errors.Wrap(tx.Delete(i.TransformToFlowStep()).Error, "failed dual-write to flows after delete")
// }

// func (r *InstallWorkflowStep) TransformToFlowStep() *FlowStep {
// 	return &FlowStep{
// 		ID:               strings.Replace(r.ID, "iws", "fls", 1),
// 		CreatedByID:      r.CreatedByID,
// 		CreatedBy:        r.CreatedBy,
// 		CreatedAt:        r.CreatedAt,
// 		UpdatedAt:        r.UpdatedAt,
// 		DeletedAt:        r.DeletedAt,
// 		OrgID:            r.OrgID,
// 		Org:              r.Org,
// 		OwnerID:          r.OwnerID,
// 		OwnerType:        r.OwnerType,
// 		FlowID:           strings.Replace(r.InstallWorkflowID, "inw", "flw", 1),
// 		Status:           r.Status,
// 		Name:             r.Name,
// 		Signal:           r.Signal,
// 		Idx:              r.Idx,
// 		ExecutionType:    FlowStepExecutionType(r.ExecutionType),
// 		StepTargetID:     r.StepTargetID,
// 		StepTargetType:   r.StepTargetType,
// 		Metadata:         r.Metadata,
// 		StartedAt:        r.StartedAt,
// 		FinishedAt:       r.FinishedAt,
// 		Finished:         r.Finished,
// 		Approval:         r.Approval,
// 		PolicyValidation: r.PolicyValidation,
// 		ExecutionTime:    r.ExecutionTime,
// 		Links:            r.Links,
// 	}
// }
