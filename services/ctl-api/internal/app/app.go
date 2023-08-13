package app

import (
	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"gorm.io/gorm"
)

type App struct {
	Model

	CreatedByID string
	Name        string
	OrgID       string
	Org         Org         `faker:"-"`
	Components  []Component `faker:"-"`
	Installs    []Install   `faker:"-"`

	SandboxReleaseID string
	SandboxRelease   SandboxRelease
}

func (a *App) BeforeCreate(tx *gorm.DB) error {
	a.ID = domains.NewAppID()
	return nil
}
