package app

type Flow = InstallWorkflow

// type FlowType string
type FlowType = InstallWorkflowType

// type FlowStepErrorBehavior string
type FlowStepErrorBehavior = StepErrorBehavior

const (
	// abort on error
	FlowStepErrorBehaviorAbort FlowStepErrorBehavior = "abort"

	// continue on error
	FlowStepErrorBehaviorContinue FlowStepErrorBehavior = "continue"
)

// We start with this to make it easier to iterate on them, for now.
// type Flow struct {
// 	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
// 	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
// 	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
// 	CreatedAt   time.Time             `json:"created_at,omitzero" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
// 	UpdatedAt   time.Time             `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
// 	DeletedAt   soft_delete.DeletedAt `gorm:"index" json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

// 	// used for RLS
// 	OrgID string `json:"org_id,omitzero" gorm:"notnull" swaggerignore:"true" temporaljson:"org_id,omitzero,omitempty"`
// 	Org   Org    `json:"-" faker:"-" temporaljson:"org,omitzero,omitempty"`

// 	Install   Install `swaggerignore:"true" json:"-" temporaljson:"install,omitzero,omitempty"`
// 	InstallID string  `json:"install_id,omitzero" gorm:"notnull;default null" temporaljson:"install_id,omitzero,omitempty"`

// 	OwnerID   string `json:"owner_id,omitzero" gorm:"type:text;check:owner_id_checker,char_length(id)=26" temporaljson:"owner_id,omitzero,omitempty"`
// 	OwnerType string `json:"owner_type,omitzero" gorm:"type:text;" temporaljson:"owner_type,omitzero,omitempty"`

// 	Type              FlowType          `json:"type,omitzero" gorm:"not null;default null" temporaljson:"type,omitzero,omitempty"`
// 	Metadata          pgtype.Hstore     `json:"metadata,omitzero" gorm:"type:hstore" swaggertype:"object,string" temporaljson:"metadata,omitzero,omitempty"`
// 	Status            CompositeStatus   `json:"status,omitzero" temporaljson:"status,omitzero,omitempty"`
// 	StepErrorBehavior FlowStepErrorBehavior `json:"step_error_behavior,omitzero" temporaljson:"step_error_behavior,omitzero,omitempty"`

// 	StartedAt  time.Time `json:"started_at,omitzero" gorm:"default:null" temporaljson:"started_at,omitzero,omitempty"`
// 	FinishedAt time.Time `json:"finished_at,omitzero" gorm:"default:null" temporaljson:"finished_at,omitzero,omitempty"`
// 	Finished   bool      `json:"finished,omitzero" gorm:"-" temporaljson:"finished,omitzero,omitempty"`

// 	// steps represent each piece of the workflow
// 	Steps []FlowStep `json:"steps,omitzero" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"steps,omitzero,omitempty"`
// 	Name  string     `json:"name,omitzero" gorm:"-" temporaljson:"name,omitzero,omitempty"`

// 	ExecutionTime time.Duration `json:"execution_time,omitzero" gorm:"-" swaggertype:"primitive,integer" temporaljson:"execution_time,omitzero,omitempty"`

// 	// InstallSandboxRuns        []InstallSandboxRun        `json:"install_sandbox_runs,omitzero" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"install_sandbox_runs,omitzero,omitempty"`
// 	// InstallDeploys            []InstallDeploy            `json:"install_deploys,omitzero" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"install_deploys,omitzero,omitempty"`
// 	// InstallActionWorkflowRuns []InstallActionWorkflowRun `json:"install_action_workflow_runs,omitzero" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"install_action_runs,omitzero,omitempty"`

// 	Links map[string]any `json:"links,omitzero,omitempty" temporaljson:"-" gorm:"-"`
// }

// func (i *Flow) TableName() string {
// 	return "install_workflows"
// }

// func (i *Flow) BeforeCreate(tx *gorm.DB) error {
// 	if i.ID == "" {
// 		i.ID = domains.NewFlowID()
// 	}

// 	i.CreatedByID = createdByIDFromContext(tx.Statement.Context)
// 	i.OrgID = orgIDFromContext(tx.Statement.Context)

// 	return nil
// }

// func (r *Flow) AfterQuery(tx *gorm.DB) error {
// 	// TODO(sdboyer) link generation is bound to installs...ugh?
// 	r.Links = links.WorkflowLinks(tx.Statement.Context, r.ID)

// 	r.ExecutionTime = generics.GetTimeDuration(r.StartedAt, r.FinishedAt)
// 	r.Finished = !r.FinishedAt.IsZero()

// 	// TODO(sdboyer) can't do this either, needs to be moved up abstraction levels
// 	// name := r.Type.Name()
// 	// if !r.FinishedAt.IsZero() {
// 	// 	name = r.Type.PastTenseName()
// 	// }
// 	// r.Name = name

// 	return nil
// }

// func (i *Flow) TransformToInstallWorkflow() *InstallWorkflow {
// 	osteps := make([]InstallWorkflowStep, 0, len(i.Steps))
// 	for _, step := range i.Steps {
// 		osteps = append(osteps, *step.TransformToInstallWorkflowStep())
// 	}

// 	return &InstallWorkflow{
// 		ID:          i.ID,
// 		CreatedByID: i.CreatedByID,
// 		CreatedBy:   i.CreatedBy,
// 		CreatedAt:   i.CreatedAt,
// 		UpdatedAt:   i.UpdatedAt,
// 		DeletedAt:   i.DeletedAt,

// 		OwnerID:   i.OwnerID,
// 		OwnerType: i.OwnerType,
// 		OrgID:     i.OrgID,
// 		Org:       i.Org,
// 		InstallID: i.OwnerID,

// 		Type:              InstallWorkflowType(i.Type),
// 		Metadata:          i.Metadata,
// 		Status:            i.Status,
// 		StepErrorBehavior: i.StepErrorBehavior,

// 		StartedAt:  i.StartedAt,
// 		FinishedAt: i.FinishedAt,
// 		Finished:   i.Finished,

// 		Steps: osteps,
// 		Name:  i.Name,

// 		ExecutionTime: i.ExecutionTime,

// 		// TODO(sdboyer) fix these up
// 		Links: i.Links,
// 	}
// }
