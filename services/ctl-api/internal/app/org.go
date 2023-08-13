package app

import (
	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"gorm.io/gorm"
)

type Org struct {
	Model

	CreatedByID string `json:"created_by_id"`
	Name        string `gorm:"uniqueIndex" json:"name"`
	Apps        []App  `faker:"-" swaggerignore:"true" json:"apps"`
}

func (o *Org) BeforeCreate(tx *gorm.DB) error {
	o.ID = domains.NewOrgID()
	return nil
}
