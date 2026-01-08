package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/eventloop/bulk"
)

type AppBranch struct {
	ID          string                `gorm:"primarykey;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgID string `json:"org_id,omitzero" gorm:"notnull;default null" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `faker:"-" json:"-" temporaljson:"org,omitzero,omitempty"`

	AppID string `json:"app_id,omitzero" gorm:"not null;index:idx_app_app_branch;uniqueIndex:idx_app_branch_name_per_app" temporaljson:"app_id,omitzero,omitempty"`
	App   App    `faker:"-" json:"-" temporaljson:"app,omitzero,omitempty"`

	Name                       string                   `gorm:"uniqueIndex:idx_app_branch_name_per_app;not null" json:"name" temporaljson:"name"`
	ConnectedGithubVCSConfigID string                   `json:"connected_github_vcs_config_id,omitzero" temporaljson:"connected_github_vcs_config_id,omitzero,omitempty"`
	ConnectedGithubVCSConfig   ConnectedGithubVCSConfig `json:"-" temporaljson:"connected_github_vcs_config,omitzero,omitempty"`

	Queue Queue `json:"queue,omitzero" gorm:"polymorphic:Owner;" temporaljson:"queue,omitzero,omitempty"`

	// LastSyncedCommit string `json:"last_synced_commit,omitzero" temporaljson:"last_synced_commit,omitzero,omitempty"`

	Workflows []Workflow `json:"workflows,omitzero" gorm:"polymorphic:Owner;constraint:OnDelete:CASCADE;" temporaljson:"workflows,omitzero,omitempty"`
}

func (a *AppBranch) EventLoops() []bulk.EventLoop {
	return []bulk.EventLoop{
		{
			Namespace: "apps",
			ID:        a.ID,
		},
	}
}

func (a *AppBranch) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &AppBranch{}, "org_id"),
			Columns: []string{
				"org_id",
			},
		},
	}
}

func (a *AppBranch) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = domains.NewAppBranchID()
	}

	if a.CreatedByID == "" {
		a.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	if a.OrgID == "" {
		a.OrgID = orgIDFromContext(tx.Statement.Context)
	}

	return nil
}
