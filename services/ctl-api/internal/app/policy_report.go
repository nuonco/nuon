package app

import (
	"time"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

// PolicyReportOwnerType represents the type of resource that was evaluated by policies.
type PolicyReportOwnerType string

const (
	PolicyReportOwnerTypeInstallDeploy     PolicyReportOwnerType = "install_deploys"
	PolicyReportOwnerTypeInstallSandboxRun PolicyReportOwnerType = "install_sandbox_runs"
	PolicyReportOwnerTypeComponentBuild    PolicyReportOwnerType = "component_builds"
)

// PolicyViolation represents a single policy violation from evaluation.
type PolicyViolation struct {
	PolicyID   string `json:"policy_id" temporaljson:"policy_id,omitempty"`
	InputIndex int    `json:"input_index" temporaljson:"input_index,omitempty"`
	Message    string `json:"message" temporaljson:"message,omitempty"`
	Severity   string `json:"severity" temporaljson:"severity,omitempty"` // "deny" or "warn"
}

// PolicyResult represents the evaluation result for a single policy.
type PolicyResult struct {
	PolicyID   string `json:"policy_id"`
	Status     string `json:"status"` // "deny", "warn", or "pass"
	DenyCount  int    `json:"deny_count"`
	WarnCount  int    `json:"warn_count"`
	PassCount  int    `json:"pass_count"`
	InputCount int    `json:"input_count"`
}

// PolicyInputRef represents a reference to an input that was evaluated.
type PolicyInputRef struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

// PolicyReport stores canonical policy evaluation results with format-agnostic storage.
type PolicyReport struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	// used for RLS
	OrgID string `json:"org_id,omitzero" gorm:"notnull" swaggerignore:"true" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `json:"-" faker:"-" temporaljson:"org,omitzero,omitempty"`

	// Denormalized context for filtering
	AppID       string  `json:"app_id,omitzero" gorm:"notnull" temporaljson:"app_id,omitzero,omitempty"`
	InstallID   *string `json:"install_id,omitzero" gorm:"default:null" temporaljson:"install_id,omitzero,omitempty"`
	ComponentID *string `json:"component_id,omitzero" gorm:"default:null" temporaljson:"component_id,omitzero,omitempty"`

	// Optional context references
	WorkflowStepPolicyValidationID *string `json:"workflow_step_policy_validation_id,omitzero" gorm:"index" temporaljson:"workflow_step_policy_validation_id,omitzero,omitempty"`
	RunnerJobID                    *string `json:"runner_job_id,omitzero" gorm:"index" temporaljson:"runner_job_id,omitzero,omitempty"`

	// Polymorphic relationship to the impacted Nuon resource
	OwnerID   string                `json:"owner_id,omitzero" gorm:"type:text;notnull;index:idx_policy_reports_owner" temporaljson:"owner_id,omitzero,omitempty"`
	OwnerType PolicyReportOwnerType `json:"owner_type,omitzero" gorm:"type:text;notnull;index:idx_policy_reports_owner" temporaljson:"owner_type,omitzero,omitempty"`

	// Canonical policy evaluation data
	EvaluatedAt time.Time         `json:"evaluated_at,omitzero" gorm:"notnull" temporaljson:"evaluated_at,omitzero,omitempty"`
	Violations  []PolicyViolation `json:"violations,omitzero" gorm:"type:jsonb;serializer:json" temporaljson:"violations,omitzero,omitempty"`
	PolicyIDs   []string          `json:"policy_ids,omitzero" gorm:"type:jsonb;serializer:json" temporaljson:"policy_ids,omitzero,omitempty"`
	Policies    []PolicyResult    `json:"policies,omitzero" gorm:"type:jsonb;serializer:json" temporaljson:"policies,omitzero,omitempty"`
	Inputs      []PolicyInputRef  `json:"inputs,omitzero" gorm:"type:jsonb;serializer:json" temporaljson:"inputs,omitzero,omitempty"`

	// Summary counts for list views
	DenyCount int `json:"deny_count" gorm:"notnull;default:0" temporaljson:"deny_count,omitzero,omitempty"`
	WarnCount int `json:"warn_count" gorm:"notnull;default:0" temporaljson:"warn_count,omitzero,omitempty"`
	PassCount int `json:"pass_count" gorm:"notnull;default:0" temporaljson:"pass_count,omitzero,omitempty"`

	Status CompositeStatus `json:"status,omitzero" gorm:"type:jsonb" temporaljson:"status,omitzero,omitempty"`
}

func (r *PolicyReport) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &PolicyReport{}, "org_id"),
			Columns: []string{
				"org_id",
			},
		},
		{
			Name: indexes.Name(db, &PolicyReport{}, "policy_reports_filter"),
			Columns: []string{
				"org_id",
				"app_id",
				"install_id",
				"owner_type",
			},
		},
	}
}

func (r *PolicyReport) TableName() string {
	return "policy_reports"
}

func (r *PolicyReport) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = domains.NewPolicyReportID()
	}

	if r.CreatedByID == "" {
		r.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}

	if r.OrgID == "" {
		r.OrgID = orgIDFromContext(tx.Statement.Context)
	}

	return nil
}
