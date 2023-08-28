package app

import (
	"time"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"gorm.io/gorm"
)

type VCSConnection struct {
	ID          string         `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string         `json:"created_by_id" gorm:"notnull"`
	CreatedAt   time.Time      `json:"created_at" gorm:"notnull"`
	UpdatedAt   time.Time      `json:"updated_at" gorm:"notnull"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	OrgID string `json:"org_id" gorm:"notnull" swaggerignore:"true"`
	Org   Org    `swaggerignore:"true" json:"-"`

	GithubInstallID           string                     `json:"github_install_id" gorm:"notnull"`
	Commits                   []VCSConnectionCommit      `json:"vcs_connection_commit" `
	ConnectedGithubVCSConfigs []ConnectedGithubVCSConfig `json:"-" gorm:"constraint:OnDelete:CASCADE;"`
}

func (v *VCSConnection) BeforeCreate(tx *gorm.DB) error {
	v.ID = domains.NewVCSConnectionID()
	v.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	return nil
}
