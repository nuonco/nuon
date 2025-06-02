package app

type FlowStepExecutionType = InstallWorkflowStepExecutionType

const (
	FlowStepExecutionTypeSystem   FlowStepExecutionType = "system"
	FlowStepExecutionTypeUser     FlowStepExecutionType = "user"
	FlowStepExecutionTypeApproval FlowStepExecutionType = "approval"
	FlowStepExecutionTypeSkipped  FlowStepExecutionType = "skipped"
)

// TODO(sdboyer) actually convert to FlowStep later
type FlowStep = InstallWorkflowStep

// type FlowStep struct {
// 	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
// 	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
// 	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
// 	CreatedAt   time.Time             `json:"created_at,omitzero" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
// 	UpdatedAt   time.Time             `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
// 	DeletedAt   soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

// 	// used for RLS
// 	OrgID string `json:"org_id,omitzero" gorm:"notnull" swaggerignore:"true" temporaljson:"org_id,omitzero,omitempty"`
// 	Org   Org    `json:"-" faker:"-" temporaljson:"org,omitzero,omitempty"`

// 	OwnerID   string `json:"owner_id,omitzero" gorm:"type:text;check:owner_id_checker,char_length(id)=26;index:idx_install_workflows_owner_id,priority:1" temporaljson:"owner_id,omitzero,omitempty"`
// 	OwnerType string `json:"owner_type,omitzero" gorm:"type:text;" temporaljson:"owner_type,omitzero,omitempty"`

// 	FlowID string `json:"flow_id,omitzero" temporaljson:"flow_id,omitzero,omitempty"`

// 	// status
// 	Status CompositeStatus `json:"status,omitzero" temporaljson:"status,omitzero,omitempty"`
// 	Name   string          `json:"name,omitzero" temporaljson:"name,omitzero,omitempty"`

// 	// the signal that needs to be called
// 	Signal Signal `json:"-" temporaljson:"signal,omitzero,omitempty"`

// 	Idx int `json:"idx,omitzero" temporaljson:"idx,omitzero,omitempty"`

// 	ExecutionType FlowStepExecutionType `json:"execution_type,omitzero" temporaljson:"execution_type"`

// 	// TODO(sdboyer) eeek, i think all of this has to change - can't be polymorphic on an unknown, unbounded set of types?
// 	// the following fields are set _once_ a step is in flight, and are orchestrated via the step's signal.
// 	//
// 	// this is a polymorphic gorm relationship to one of the following objects:
// 	//
// 	// install_cloudformation_stack
// 	// install_sandbox_run
// 	// install_runner_update
// 	// install_deploy
// 	// install_action_workflow_run (can be many of these)
// 	StepTargetID   string `json:"step_target_id,omitzero" gorm:"type:text;check:owner_id_checker,char_length(id)=26" temporaljson:"step_target_id,omitzero,omitempty"`
// 	StepTargetType string `json:"step_target_type,omitzero" gorm:"type:text;" temporaljson:"step_target_type,omitzero,omitempty"`

// 	Metadata pgtype.Hstore `json:"metadata,omitzero" gorm:"type:hstore" swaggertype:"object,string" temporaljson:"metadata,omitzero,omitempty"`

// 	StartedAt  time.Time `json:"started_at,omitzero" gorm:"default:null" temporaljson:"started_at,omitzero,omitempty"`
// 	FinishedAt time.Time `json:"finished_at,omitzero" gorm:"default:null" temporaljson:"finished_at,omitzero,omitempty"`
// 	Finished   bool      `json:"finished,omitzero" gorm:"-" temporaljson:"finished,omitzero,omitempty"`

// 	// the step approval is built into each step at the runner level.
// 	ApprovalID         string                               `json:"approval_id,omitzero" gorm:"type:text;check:owner_id_checker,char_length(id)=26" temporaljson:"approval_id,omitzero,omitempty"`
// 	Approval           *InstallWorkflowStepApproval         `json:"approval,omitzero" temporaljson:"approval,omitzero,omitempty"` // TODO(sdboyer) abstract approvals away from installs
// 	PolicyValidationID string                               `json:"policy_validation_id,omitzero" gorm:"type:text;check:owner_id_checker,char_length(id)=26" temporaljson:"policy_validation_id,omitzero,omitempty"`
// 	PolicyValidation   *InstallWorkflowStepPolicyValidation `json:"policy_validation,omitzero" temporaljson:"policy_validation,omitzero,omitempty"`

// 	ExecutionTime time.Duration `json:"execution_time,omitzero" gorm:"-" swaggertype:"primitive,integer" temporaljson:"execution_time,omitzero,omitempty"`

// 	Links map[string]any `json:"links,omitzero,omitempty" temporaljson:"-" gorm:"-"`
// }

// func (a *FlowStep) BeforeCreate(tx *gorm.DB) error {
// 	if a.ID == "" {
// 		a.ID = domains.NewFlowStepID()
// 	}

// 	if a.CreatedByID == "" {
// 		a.CreatedByID = createdByIDFromContext(tx.Statement.Context)
// 	}

// 	if a.OrgID == "" {
// 		a.OrgID = orgIDFromContext(tx.Statement.Context)
// 	}
// 	return nil
// }

// func (r *FlowStep) AfterQuery(tx *gorm.DB) error {
// 	// TODO(sdboyer) link generation is bound to installs...ugh?
// 	r.Links = links.InstallWorkflowStepLinks(tx.Statement.Context, r.ID)

// 	r.ExecutionTime = generics.GetTimeDuration(r.StartedAt, r.FinishedAt)
// 	r.Finished = !r.FinishedAt.IsZero()
// 	return nil
// }

// func (r *FlowStep) TransformToInstallWorkflowStep() *InstallWorkflowStep {
// 	return &InstallWorkflowStep{
// 		ID:                strings.Replace(r.ID, "flw", "inw", 1),
// 		CreatedByID:       r.CreatedByID,
// 		CreatedBy:         r.CreatedBy,
// 		CreatedAt:         r.CreatedAt,
// 		UpdatedAt:         r.UpdatedAt,
// 		DeletedAt:         r.DeletedAt,
// 		OrgID:             r.OrgID,
// 		Org:               r.Org,
// 		InstallID:         r.OwnerID,
// 		OwnerID:           r.OwnerID,
// 		OwnerType:         r.OwnerType,
// 		InstallWorkflowID: strings.Replace(r.FlowID, "flw", "inw", 1),
// 		Status:            r.Status,
// 		Name:              r.Name,
// 		Signal:            r.Signal,
// 		Idx:               r.Idx,
// 		ExecutionType:     InstallWorkflowStepExecutionType(r.ExecutionType),
// 		StepTargetID:      r.StepTargetID,
// 		StepTargetType:    r.StepTargetType,
// 		Metadata:          r.Metadata,
// 		StartedAt:         r.StartedAt,
// 		FinishedAt:        r.FinishedAt,
// 		Finished:          r.Finished,
// 		Approval:          r.Approval,
// 		PolicyValidation:  r.PolicyValidation,
// 		ExecutionTime:     r.ExecutionTime,
// 		Links:             r.Links,
// 	}
// }
