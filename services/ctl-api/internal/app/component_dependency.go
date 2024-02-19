package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

type ComponentDependency struct {
	CreatedByID string                `json:"created_by_id" gorm:"notnull"`
	CreatedAt   time.Time             `json:"created_at" gorm:"notnull"`
	UpdatedAt   time.Time             `json:"updated_at" gorm:"notnull"`
	DeletedAt   soft_delete.DeletedAt `json:"-"`

	OrgID string `json:"org_id" gorm:"notnull" swaggerignore:"true"`
	Org   Org    `json:"-" faker:"-"`

	ComponentID  string `gorm:"primary_key"`
	DependencyID string `gorm:"primary_key"`
}

func (c *ComponentDependency) BeforeSave(tx *gorm.DB) error {
	c.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	c.OrgID = orgIDFromContext(tx.Statement.Context)

	return nil
}
