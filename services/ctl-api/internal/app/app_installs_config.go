package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
)

type AppInstallsConfig struct {
	ID          string                `gorm:"primarykey;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgID string `json:"org_id,omitzero" gorm:"notnull;default null" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `faker:"-" json:"-" temporaljson:"org,omitzero,omitempty"`

	AppID string `json:"app_id,omitzero" gorm:"not null" temporaljson:"app_id,omitzero,omitempty"`
	App   App    `faker:"-" json:"-" temporaljson:"app,omitzero,omitempty"`

	VCSType         string  `json:"vcs_type,omitzero" gorm:"not null" temporaljson:"vcs_type,omitzero,omitempty"`
	VCSConnectionID *string `json:"vcs_connection_id,omitempty" temporaljson:"vcs_connection_id,omitzero,omitempty"`

	Repo      string `json:"repo,omitzero" gorm:"not null" temporaljson:"repo,omitzero,omitempty"`
	Branch    string `json:"branch,omitzero" gorm:"not null" temporaljson:"branch,omitzero,omitempty"`
	Directory string `json:"directory,omitzero" gorm:"not null;default:'.'" temporaljson:"directory,omitzero,omitempty"`

	Source string `json:"source,omitzero" gorm:"not null;default:'config'" temporaljson:"source,omitzero,omitempty"`

	ConnectedGithubVCSConfig *ConnectedGithubVCSConfig `json:"connected_github_vcs_config,omitempty" gorm:"polymorphic:ComponentConfig;polymorphicValue:app_installs_configs" temporaljson:"connected_github_vcs_config,omitzero,omitempty"`
	PublicGitVCSConfig       *PublicGitVCSConfig       `json:"public_git_vcs_config,omitempty" gorm:"polymorphic:ComponentConfig;polymorphicValue:app_installs_configs" temporaljson:"public_git_vcs_config,omitzero,omitempty"`
}

func (a *AppInstallsConfig) TableName() string {
	return "app_installs_configs"
}

func (a *AppInstallsConfig) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name:    indexes.Name(db, &AppInstallsConfig{}, "org_id"),
			Columns: []string{"org_id"},
		},
		{
			Name:    indexes.Name(db, &AppInstallsConfig{}, "app_id"),
			Columns: []string{"app_id"},
		},
		{
			Name:    indexes.Name(db, &AppInstallsConfig{}, "repo_branch"),
			Columns: []string{"org_id", "repo", "branch"},
		},
	}
}

func (a *AppInstallsConfig) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = domains.NewAppInstallsConfigID()
	}

	if a.CreatedByID == "" {
		a.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	if a.OrgID == "" {
		a.OrgID = orgIDFromContext(tx.Statement.Context)
	}

	return nil
}
