package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
)

type VCSConnection struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string                `json:"created_by_id" gorm:"not null;default:null"`
	CreatedBy   Account               `json:"created_by"`
	CreatedAt   time.Time             `json:"created_at" gorm:"notnull"`
	UpdatedAt   time.Time             `json:"updated_at" gorm:"notnull"`
	DeletedAt   soft_delete.DeletedAt `gorm:"index" json:"-"`

	OrgID string `json:"org_id" swaggerignore:"true" gorm:"index:idx_github_install_id,unique"`
	Org   Org    `swaggerignore:"true" json:"-"`

	GithubInstallID           string                     `json:"github_install_id" gorm:"index:idx_github_install_id,unique"`
	Commits                   []VCSConnectionCommit      `json:"vcs_connection_commit" gorm:"constraint:OnDelete:CASCADE;"`
	ConnectedGithubVCSConfigs []ConnectedGithubVCSConfig `json:"-" gorm:"constraint:OnDelete:CASCADE;"`
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
