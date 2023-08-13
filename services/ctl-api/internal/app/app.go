package app

import (
	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"gorm.io/gorm"
)

type App struct {
	Model

	CreatedByID string      `json:"created_by_id"`
	Name        string      `json:"name"`
	OrgID       string      `json:"-"`
	Org         Org         `faker:"-" json:"-"`
	Components  []Component `faker:"-" json:"components,omitempty"`
	Installs    []Install   `faker:"-" json:"installs,omitempty"`

	SandboxReleaseID string         `json:"-"`
	SandboxRelease   SandboxRelease `json:"sandbox_release,omitempty"`
}

func (a *App) BeforeCreate(tx *gorm.DB) error {
	a.ID = domains.NewAppID()
	return nil
}
