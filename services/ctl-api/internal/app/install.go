package app

import (
	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"gorm.io/gorm"
)

type Install struct {
	Model
	CreatedByID string

	Name  string
	AppID string
	App   App
}

func (i *Install) BeforeCreate(tx *gorm.DB) error {
	i.ID = domains.NewInstallID()
	return nil
}
