package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
)

type VCSConnection struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `gorm:"index" json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgID string `json:"org_id,omitzero" swaggerignore:"true" gorm:"index:idx_github_install_id,unique" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `swaggerignore:"true" json:"-" temporaljson:"org,omitzero,omitempty"`

	GithubInstallID string `json:"github_install_id,omitzero" gorm:"index:idx_github_install_id,unique" temporaljson:"github_install_id,omitzero,omitempty"`

	GithubAccountID   string `json:"github_account_id,omitempty" gorm:"not null;default:''" temporaljson:"github_account_id,omitzero,omitempty"`
	GithubAccountName string `json:"github_account_name,omitempty" gorm:"not null;default:''" temporaljson:"github_account_name,omitzero,omitempty"`

	Commits                   []VCSConnectionCommit      `json:"vcs_connection_commit,omitzero" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"commits,omitzero,omitempty"`
	ConnectedGithubVCSConfigs []ConnectedGithubVCSConfig `json:"-" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"connected_github_vcs_configs,omitzero,omitempty"`
}

func (v *VCSConnection) BeforeCreate(tx *gorm.DB) error {
	v.ID = domains.NewVCSConnectionID()
	if v.OrgID == "" {
		v.OrgID = orgIDFromContext(tx.Statement.Context)
	}
	if v.CreatedByID == "" {
		v.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	return nil
}
