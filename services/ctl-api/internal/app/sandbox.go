package app

import (
	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"gorm.io/gorm"
)

type Sandbox struct {
	Model

	Name        string           `gorm:"unique" json:"name"`
	Description string           `json:"description"`
	Releases    []SandboxRelease `json:"releases"`
}

func (s *Sandbox) BeforeCreate(tx *gorm.DB) error {
	s.ID = domains.NewSandboxID()
	return nil
}
