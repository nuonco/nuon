package app

import (
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/views"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/viewsql"
)

type AppBranchConfig struct {
	ID          string                `gorm:"primarykey;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgID string `json:"org_id,omitzero" gorm:"notnull;default null" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `faker:"-" json:"-" temporaljson:"org,omitzero,omitempty"`

	AppBranchID string    `json:"app_branch_id,omitzero" gorm:"not null;index:idx_app_branch_configs" temporaljson:"app_branch_id,omitzero,omitempty"`
	AppBranch   AppBranch `faker:"-" json:"-" temporaljson:"app_branch,omitzero,omitempty"`

	PublicGitVCSConfig       *PublicGitVCSConfig       `gorm:"polymorphic:ComponentConfig;constraint:OnDelete:CASCADE;" json:"public_git_vcs_config,omitzero,omitempty" temporaljson:"public_git_vcs_config,omitzero,omitempty"`
	ConnectedGithubVCSConfig *ConnectedGithubVCSConfig `gorm:"polymorphic:ComponentConfig;constraint:OnDelete:CASCADE;" json:"connected_github_vcs_config,omitzero,omitempty" temporaljson:"connected_github_vcs_config,omitzero,omitempty"`

	InstallGroups []AppBranchInstallGroup `json:"install_groups,omitzero" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"install_groups,omitzero,omitempty"`

	ComponentIDs pq.StringArray `gorm:"type:text[]" json:"component_ids,omitzero" temporaljson:"component_ids,omitzero,omitempty" swaggertype:"array,string"`
	ActionIDs    pq.StringArray `gorm:"type:text[]" json:"action_ids,omitzero" temporaljson:"action_ids,omitzero,omitempty" swaggertype:"array,string"`

	Workflows []Workflow `json:"workflows,omitzero" gorm:"polymorphic:Owner;constraint:OnDelete:CASCADE;" temporaljson:"workflows,omitzero,omitempty"`

	// generated view field
	ConfigNumber int `json:"config_number,omitzero" gorm:"->;-:migration" temporaljson:"config_number,omitzero,omitempty"`
}

func (a *AppBranchConfig) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &AppBranchConfig{}, "org_id"),
			Columns: []string{
				"org_id",
			},
		},
		{
			Name: indexes.Name(db, &AppBranchConfig{}, "app_branch_id"),
			Columns: []string{
				"app_branch_id",
			},
		},
	}
}

func (a *AppBranchConfig) UseView() bool {
	return true
}

func (a *AppBranchConfig) ViewVersion() string {
	return "v1"
}

func (a *AppBranchConfig) Views(db *gorm.DB) []migrations.View {
	return []migrations.View{
		{
			Name:          views.DefaultViewName(db, &AppBranchConfig{}, 1),
			SQL:           viewsql.AppBranchConfigsViewV1,
			AlwaysReapply: true,
		},
	}
}

func (a *AppBranchConfig) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = domains.NewAppBranchConfigID()
	}

	if a.CreatedByID == "" {
		a.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	if a.OrgID == "" {
		a.OrgID = orgIDFromContext(tx.Statement.Context)
	}

	return nil
}
