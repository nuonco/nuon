package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
)

type AppSandboxBuildStatus string

const (
	AppSandboxBuildStatusQueued   AppSandboxBuildStatus = "queued"
	AppSandboxBuildStatusPlanning AppSandboxBuildStatus = "planning"
	AppSandboxBuildStatusBuilding AppSandboxBuildStatus = "building"
	AppSandboxBuildStatusActive   AppSandboxBuildStatus = "active"
	AppSandboxBuildStatusError    AppSandboxBuildStatus = "error"
)

// AppSandboxBuild represents a build run against an app's sandbox terraform config.
// It is triggered by the app branches workflow and runs terraform plan against the sandbox config.
type AppSandboxBuild struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"created_by,omitzero" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `gorm:"index" json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	// used for RLS
	OrgID string `json:"org_id,omitzero" gorm:"notnull" swaggerignore:"true" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `json:"-" faker:"-" temporaljson:"org,omitzero,omitempty"`

	AppID              string           `json:"app_id,omitzero" gorm:"notnull" temporaljson:"app_id,omitzero,omitempty"`
	App                App              `json:"-" temporaljson:"app,omitzero,omitempty"`
	AppConfigID        string           `json:"app_config_id,omitzero" gorm:"notnull" temporaljson:"app_config_id,omitzero,omitempty"`
	AppConfig          AppConfig        `json:"-" temporaljson:"app_config,omitzero,omitempty"`
	AppSandboxConfigID string           `json:"app_sandbox_config_id,omitzero" gorm:"notnull" temporaljson:"app_sandbox_config_id,omitzero,omitempty"`
	AppSandboxConfig   AppSandboxConfig `json:"-" temporaljson:"app_sandbox_config,omitzero,omitempty"`

	VCSConnectionCommitID *string              `json:"vcs_connection_commit_id,omitempty" temporaljson:"vcs_connection_commit_id,omitzero,omitempty"`
	VCSConnectionCommit   *VCSConnectionCommit `json:"vcs_connection_commit,omitzero" temporaljson:"vcs_connection_commit,omitzero,omitempty"`

	RunnerJob RunnerJob `json:"runner_job,omitzero" gorm:"polymorphic:Owner;" temporaljson:"runner_job,omitzero,omitempty"`
	LogStream LogStream `json:"log_stream,omitzero" gorm:"polymorphic:Owner;" temporaljson:"log_stream,omitzero,omitempty"`

	Status            AppSandboxBuildStatus `json:"status,omitzero" gorm:"notnull" swaggertype:"string" temporaljson:"status,omitzero,omitempty"`
	StatusDescription string                `json:"status_description,omitzero" gorm:"notnull" temporaljson:"status_description,omitzero,omitempty"`
	StatusV2          CompositeStatus       `json:"status_v2,omitzero" gorm:"type:jsonb" temporaljson:"status_v2,omitzero,omitempty"`
}

func (a *AppSandboxBuild) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &AppSandboxBuild{}, "org_id"),
			Columns: []string{
				"org_id",
			},
		},
		{
			Name: indexes.Name(db, &AppSandboxBuild{}, "app_id"),
			Columns: []string{
				"app_id",
			},
		},
		{
			Name: indexes.Name(db, &AppSandboxBuild{}, "app_config_id"),
			Columns: []string{
				"app_config_id",
			},
		},
	}
}

func (a *AppSandboxBuild) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = domains.NewBuildID()
	}
	if a.CreatedByID == "" {
		a.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	if a.OrgID == "" {
		a.OrgID = orgIDFromContext(tx.Statement.Context)
	}
	return nil
}

func (a *AppSandboxBuild) AfterQuery(tx *gorm.DB) error {
	if a.StatusV2.Status != "" {
		a.Status = AppSandboxBuildStatus(a.StatusV2.Status)
		a.StatusDescription = a.StatusV2.StatusHumanDescription
	}
	return nil
}
