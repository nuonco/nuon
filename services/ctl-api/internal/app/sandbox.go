package app

import (
	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"gorm.io/gorm"
)

type Sandbox struct {
	Model

	Name        string `gorm:"unique"`
	Description string
	Releases    []SandboxRelease
}

func (o *Sandbox) BeforeCreate(tx *gorm.DB) error {
	o.ID = domains.NewSandboxID()
	return nil
}
