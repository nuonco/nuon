package app

import (
	"time"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"gorm.io/gorm"
)

type App struct {
	ID          string         `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string         `json:"created_by_id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	Name       string      `json:"name"`
	OrgID      string      `json:"org_id"`
	Org        Org         `faker:"-" json:"-"`
	Components []Component `faker:"-" json:"-" swaggerignore:"true"`
	Installs   []Install   `faker:"-" json:"-" swaggerignore:"true"`
	Status     string      `json:"status"`

	SandboxReleaseID string         `json:"-"`
	SandboxRelease   SandboxRelease `json:"sandbox_release,omitempty"`
}

func (a *App) BeforeCreate(tx *gorm.DB) error {
	a.ID = domains.NewAppID()
	return nil
}
