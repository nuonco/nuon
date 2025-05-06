package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

// ComponentDependency is a many2many table used by gorm under the hood
type ComponentDependency struct {
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgID string `json:"org_id,omitzero" gorm:"notnull" swaggerignore:"true" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `json:"-" faker:"-" temporaljson:"org,omitzero,omitempty"`

	ComponentID  string `json:"component_id,omitzero" gorm:"primary_key" temporaljson:"component_id,omitzero,omitempty"`
	DependencyID string `json:"dependency_id,omitzero" gorm:"primary_key" temporaljson:"dependency_id,omitzero,omitempty"`
}

func (c *ComponentDependency) BeforeSave(tx *gorm.DB) error {
	c.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	c.OrgID = orgIDFromContext(tx.Statement.Context)

	return nil
}
