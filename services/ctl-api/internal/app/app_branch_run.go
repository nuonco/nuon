package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
)

// AppBranchRun represents a single execution of an app branch workflow.
// Each run is triggered manually or automatically and processes the branch's
// configuration through the install groups.
type AppBranchRun struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"created_by,omitempty" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgID string `json:"org_id,omitzero" gorm:"notnull" swaggerignore:"true" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `json:"-" faker:"-" temporaljson:"org,omitzero,omitempty"`

	AppBranchID string    `json:"app_branch_id,omitzero" gorm:"notnull" swaggerignore:"true" temporaljson:"app_branch_id,omitzero,omitempty"`
	AppBranch   AppBranch `json:"app_branch,omitempty" temporaljson:"app_branch,omitzero,omitempty"`

	AppBranchConfigID string          `json:"app_branch_config_id,omitzero" gorm:"notnull" swaggerignore:"true" temporaljson:"app_branch_config_id,omitzero,omitempty"`
	AppBranchConfig   AppBranchConfig `json:"app_branch_config,omitempty" temporaljson:"app_branch_config,omitzero,omitempty"`

	WorkflowID *string   `json:"workflow_id,omitempty" swaggerignore:"true" temporaljson:"workflow_id,omitzero,omitempty"`
	Workflow   *Workflow `json:"workflow,omitempty" temporaljson:"workflow,omitzero,omitempty"`

	// Status tracks the current state of the run
	// Values: pending, running, success, failed, cancelled
	Status string `json:"status,omitzero" gorm:"notnull;default:'pending'" temporaljson:"status,omitzero,omitempty"`

	// Force indicates if this run was forced (bypassing change detection)
	Force bool `json:"force,omitzero" temporaljson:"force,omitzero,omitempty"`

	// StartedAt tracks when execution actually began
	StartedAt *time.Time `json:"started_at,omitempty" temporaljson:"started_at,omitzero,omitempty"`

	// CompletedAt tracks when execution finished
	CompletedAt *time.Time `json:"completed_at,omitempty" temporaljson:"completed_at,omitzero,omitempty"`

	// ErrorMessage stores any error that occurred during execution
	ErrorMessage string `json:"error_message,omitempty" temporaljson:"error_message,omitzero,omitempty"`

	// AppConfigID is the app config that was created/synced during this run
	AppConfigID string `json:"app_config_id,omitempty" temporaljson:"app_config_id,omitzero,omitempty"`

	// LogStreamID is the log stream created during this run for event tracking
	LogStreamID *string    `json:"log_stream_id,omitempty" temporaljson:"log_stream_id,omitzero,omitempty"`
	LogStream   *LogStream `json:"log_stream,omitempty" temporaljson:"log_stream,omitzero,omitempty"`

	// CommitSHA is the VCS commit that triggered or is associated with this run
	// DEPRECATED: Use VCSConnectionCommit relationship instead
	CommitSHA string `json:"commit_sha,omitzero" temporaljson:"commit_sha,omitzero,omitempty"`

	// VCSConnectionCommit is the full commit record associated with this run
	VCSConnectionCommitID *string              `json:"vcs_connection_commit_id,omitempty" swaggerignore:"true" temporaljson:"vcs_connection_commit_id,omitzero,omitempty"`
	VCSConnectionCommit   *VCSConnectionCommit `json:"vcs_connection_commit,omitempty" temporaljson:"vcs_connection_commit,omitzero,omitempty"`

	// QueueSignal is the signal that was enqueued to trigger this run
	QueueSignal *QueueSignal `json:"queue_signal,omitempty" gorm:"polymorphic:Owner;" temporaljson:"queue_signal,omitzero,omitempty"`
}

func (a *AppBranchRun) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &AppBranchRun{}, "org_id"),
			Columns: []string{
				"org_id",
			},
		},
		{
			Name: indexes.Name(db, &AppBranchRun{}, "app_branch_id"),
			Columns: []string{
				"app_branch_id",
			},
		},
		{
			Name: indexes.Name(db, &AppBranchRun{}, "workflow_id"),
			Columns: []string{
				"workflow_id",
			},
		},
		{
			Name: indexes.Name(db, &AppBranchRun{}, "status"),
			Columns: []string{
				"status",
			},
		},
		{
			Name: indexes.Name(db, &AppBranchRun{}, "vcs_connection_commit_id"),
			Columns: []string{
				"vcs_connection_commit_id",
			},
		},
	}
}

func (a *AppBranchRun) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = domains.NewAppBranchRunID()
	}

	if a.CreatedByID == "" {
		a.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}

	if a.OrgID == "" {
		a.OrgID = orgIDFromContext(tx.Statement.Context)
	}

	return nil
}
