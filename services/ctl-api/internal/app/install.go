package app

import (
	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"gorm.io/gorm"
)

type Install struct {
	Model
	Name string

	App   App `swaggerignore:"true" json:"-"`
	AppID string

	AWSAccountID string
	AWSAccount   AWSAccount

	SandboxReleaseID string
	SandboxRelease   SandboxRelease
}

func (i *Install) BeforeCreate(tx *gorm.DB) error {
	i.ID = domains.NewInstallID()
	return nil
}
