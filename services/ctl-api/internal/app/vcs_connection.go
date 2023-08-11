package app

import (
	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"gorm.io/gorm"
)

type VCSConnection struct {
	Model

	OrgID string
	Org   Org `swaggerignore:"true"`

	GithubInstallID string
}

func (v *VCSConnection) BeforeCreate(tx *gorm.DB) error {
	v.ID = domains.NewVCSConnectionID()
	return nil
}
