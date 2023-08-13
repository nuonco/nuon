package app

import (
	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"gorm.io/gorm"
)

type Org struct {
	Model

	CreatedByID     string
	Name            string `gorm:"uniqueIndex"`
	Apps            []App  `faker:"-" swaggerignore:"true"`
	IsNew           bool   `gorm:"-:all"`
	GithubInstallID string
}

func (o *Org) BeforeCreate(tx *gorm.DB) error {
	o.ID = domains.NewOrgID()
	return nil
}
