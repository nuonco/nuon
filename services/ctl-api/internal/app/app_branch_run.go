package app

import (
	"database/sql"
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
)

type AppBranchRunType string

const (
	AppBranchRunTypeManual     AppBranchRunType = "manual-run"
	AppBranchRunTypeGit        AppBranchRunType = "git-run"
	AppBranchRunTypeGitPreview AppBranchRunType = "git-preview-run"
)

// AppBranchRunLabelBuildsCompleted is set on AppBranchRun.labels when the builds
// step finishes. Used to select baseline runs for AppBranchRunComparison.
const AppBranchRunLabelBuildsCompleted = "builds_completed"

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

	AppBranchConfigID      string          `json:"app_branch_config_id,omitzero" gorm:"notnull" swaggerignore:"true" temporaljson:"app_branch_config_id,omitzero,omitempty"`
	AppBranchConfig        AppBranchConfig `json:"app_branch_config,omitempty" temporaljson:"app_branch_config,omitzero,omitempty"`
	TriggerEventDispatchID *string         `json:"trigger_event_dispatch_id,omitempty" gorm:"<-:create" temporaljson:"trigger_event_dispatch_id,omitzero,omitempty"`
	TriggerEventDispatch   *EventDispatch  `json:"-" gorm:"constraint:OnDelete:SET NULL" temporaljson:"-"`

	WorkflowID *string   `json:"workflow_id,omitempty" temporaljson:"workflow_id,omitzero,omitempty"`
	Workflow   *Workflow `json:"workflow,omitempty" temporaljson:"workflow,omitzero,omitempty"`

	Status string `json:"status,omitzero" gorm:"notnull;default:'pending'" temporaljson:"status,omitzero,omitempty"`

	RunType AppBranchRunType `json:"run_type,omitzero" gorm:"notnull;default:'manual-run'" temporaljson:"run_type,omitzero,omitempty"`

	Force bool `json:"force,omitzero" temporaljson:"force,omitzero,omitempty"`

	PlanOnly bool `json:"plan_only,omitzero" temporaljson:"plan_only,omitzero,omitempty"`

	EventType string `json:"event_type,omitempty" temporaljson:"event_type,omitzero,omitempty"`

	PRNumber   *int   `json:"pr_number,omitempty" temporaljson:"pr_number,omitzero,omitempty"`
	HeadSHA    string `json:"head_sha,omitempty" temporaljson:"head_sha,omitzero,omitempty"`
	BaseBranch string `json:"base_branch,omitempty" temporaljson:"base_branch,omitzero,omitempty"`

	GithubCommentID *int64 `json:"github_comment_id,omitempty" temporaljson:"github_comment_id,omitzero,omitempty"`
	NoConfigChanges bool   `json:"no_config_changes,omitempty" temporaljson:"no_config_changes,omitzero,omitempty"`

	StartedAt *time.Time `json:"started_at,omitempty" temporaljson:"started_at,omitzero,omitempty"`

	CompletedAt *time.Time `json:"completed_at,omitempty" temporaljson:"completed_at,omitzero,omitempty"`

	ErrorMessage string `json:"error_message,omitempty" temporaljson:"error_message,omitzero,omitempty"`

	CompositeError *compositeerrors.CompositeErrorData `json:"composite_error,omitempty" gorm:"type:jsonb" temporaljson:"composite_error,omitzero,omitempty"`

	AppConfigID string `json:"app_config_id,omitempty" temporaljson:"app_config_id,omitzero,omitempty"`

	LogStreamID *string    `json:"log_stream_id,omitempty" temporaljson:"log_stream_id,omitzero,omitempty"`
	LogStream   *LogStream `json:"log_stream,omitempty" temporaljson:"log_stream,omitzero,omitempty"`

	Comparison *AppBranchRunComparison `json:"comparison,omitempty" gorm:"foreignKey:HeadRunID" temporaljson:"comparison,omitzero,omitempty"`

	Preview *AppBranchRunPreview `json:"preview,omitempty" gorm:"foreignKey:AppBranchRunID;references:ID" temporaljson:"preview,omitzero,omitempty"`

	VCSConnectionCommitID *string              `json:"vcs_connection_commit_id,omitempty" swaggerignore:"true" temporaljson:"vcs_connection_commit_id,omitzero,omitempty"`
	VCSConnectionCommit   *VCSConnectionCommit `json:"vcs_connection_commit,omitempty" temporaljson:"vcs_connection_commit,omitzero,omitempty"`

	QueueSignal *QueueSignal `json:"queue_signal,omitempty" gorm:"polymorphic:Owner;" temporaljson:"queue_signal,omitzero,omitempty"`

	AwaitingApproval bool `json:"awaiting_approval,omitzero" gorm:"-" temporaljson:"awaiting_approval,omitzero,omitempty"`

	labels.Labeled
}

// IsPreview reports whether the run is a preview run (has AppBranchRunPreview child).
// PlanOnly and git-preview-run are legacy signals mapped to preview at create time.
func (a *AppBranchRun) IsPreview() bool {
	if a.Preview != nil {
		return true
	}
	return a.RunType == AppBranchRunTypeGitPreview || a.PlanOnly
}

func (a *AppBranchRun) PreviewGitHubSetStatuses() bool {
	if a.Preview != nil {
		return a.Preview.GitHubSetStatuses()
	}
	return a.IsPreview()
}

func (a *AppBranchRun) PreviewGitHubComment() bool {
	if a.Preview != nil {
		return a.Preview.GitHubComment()
	}
	return a.IsPreview()
}

func (a *AppBranchRun) PreviewMode() AppBranchRunPreviewMode {
	if a.Preview != nil && a.Preview.Mode != "" {
		return a.Preview.Mode
	}
	if a.IsPreview() {
		return AppBranchRunPreviewModePlanOnly
	}
	return ""
}

func (a *AppBranchRun) PreviewInstallPlanOnly() bool {
	if a.Preview != nil {
		return a.Preview.Mode == AppBranchRunPreviewModePlanOnly || a.Preview.Mode == AppBranchRunPreviewModePlanInfra
	}
	return a.PlanOnly
}

func (a *AppBranchRun) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{Name: indexes.Name(db, &AppBranchRun{}, "trigger_event_dispatch_id"), Columns: []string{"trigger_event_dispatch_id"}, UniqueValue: sql.NullBool{Bool: true, Valid: true}, Option: "WHERE deleted_at = 0 AND trigger_event_dispatch_id IS NOT NULL"},
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
