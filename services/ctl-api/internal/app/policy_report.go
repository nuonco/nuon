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
	PolicyID      string `json:"policy_id" temporaljson:"policy_id,omitempty"`
	PolicyName    string `json:"policy_name,omitempty" temporaljson:"policy_name,omitempty"`
	InputIndex    int    `json:"input_index" temporaljson:"input_index,omitempty"`
	InputIdentity string `json:"input_identity,omitempty" temporaljson:"input_identity,omitempty"` // Human-readable input reference (e.g., "Deployment/default/nginx")
	Message       string `json:"message" temporaljson:"message,omitempty"`
	Severity      string `json:"severity" temporaljson:"severity,omitempty"` // "deny" or "warn"
}

// PolicyResult represents the evaluation result for a single policy.
type PolicyResult struct {
	PolicyID   string `json:"policy_id"`
	PolicyName string `json:"policy_name,omitempty"`
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
	Name string `json:"name,omitempty"`
}

// PolicyReport stores canonical policy evaluation results with format-agnostic storage.
type PolicyReport struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `gorm:"foreignKey:CreatedByID;references:ID" json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `gorm:"index" json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	// used for RLS
	OrgID string `json:"org_id,omitzero" gorm:"notnull" swaggerignore:"true" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `gorm:"foreignKey:OrgID;references:ID" json:"-" faker:"-" temporaljson:"org,omitzero,omitempty"`

	// Denormalized context for filtering
	AppID       string  `json:"app_id,omitzero" gorm:"notnull" temporaljson:"app_id,omitzero,omitempty"`
	InstallID   *string `json:"install_id,omitzero" gorm:"default:null" temporaljson:"install_id,omitzero,omitempty"`
	ComponentID *string `json:"component_id,omitzero" gorm:"default:null" temporaljson:"component_id,omitzero,omitempty"`

	// Denormalized display names for human-readable reports
	OrgName       string  `json:"org_name,omitzero" gorm:"default:null" temporaljson:"org_name,omitzero,omitempty"`
	AppName       string  `json:"app_name,omitzero" gorm:"default:null" temporaljson:"app_name,omitzero,omitempty"`
	InstallName   *string `json:"install_name,omitzero" gorm:"default:null" temporaljson:"install_name,omitzero,omitempty"`
	ComponentName *string `json:"component_name,omitzero" gorm:"default:null" temporaljson:"component_name,omitzero,omitempty"`

	// Optional context references (indexes defined in Indexes())
	WorkflowStepPolicyValidationID *string `json:"workflow_step_policy_validation_id,omitzero" temporaljson:"workflow_step_policy_validation_id,omitzero,omitempty"`
	RunnerJobID                    *string `json:"runner_job_id,omitzero" temporaljson:"runner_job_id,omitzero,omitempty"`

	// Polymorphic relationship to the impacted Nuon resource (indexes defined in Indexes())
	OwnerID   string                `json:"owner_id,omitzero" gorm:"type:varchar(26);notnull;check:owner_id_checker,char_length(owner_id)=26" temporaljson:"owner_id,omitzero,omitempty"`
	OwnerType PolicyReportOwnerType `json:"owner_type,omitzero" gorm:"type:text;notnull;check:owner_type_checker,owner_type IN ('install_deploys','install_sandbox_runs','component_builds')" temporaljson:"owner_type,omitzero,omitempty"`

	// Canonical policy evaluation data
	EvaluatedAt time.Time         `json:"evaluated_at,omitzero" gorm:"notnull" temporaljson:"evaluated_at,omitzero,omitempty"`
	Violations  []PolicyViolation `json:"violations,omitzero" gorm:"type:jsonb;serializer:json;default:'[]'" temporaljson:"violations,omitzero,omitempty"`
	PolicyIDs   []string          `json:"policy_ids,omitzero" gorm:"type:jsonb;serializer:json;default:'[]'" temporaljson:"policy_ids,omitzero,omitempty"`
	Policies    []PolicyResult    `json:"policies,omitzero" gorm:"type:jsonb;serializer:json;default:'[]'" temporaljson:"policies,omitzero,omitempty"`
	Inputs      []PolicyInputRef  `json:"inputs,omitzero" gorm:"type:jsonb;serializer:json;default:'[]'" temporaljson:"inputs,omitzero,omitempty"`

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
		{
			Name: indexes.Name(db, &PolicyReport{}, "owner_latest"),
			Columns: []string{
				"org_id",
				"owner_type",
				"owner_id",
				"evaluated_at",
			},
		},
		{
			Name: indexes.Name(db, &PolicyReport{}, "workflow_step_policy_validation_id"),
			Columns: []string{
				"org_id",
				"workflow_step_policy_validation_id",
			},
		},
		{
			Name: indexes.Name(db, &PolicyReport{}, "runner_job_id"),
			Columns: []string{
				"org_id",
				"runner_job_id",
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
