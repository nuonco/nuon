package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

// ComponentConfigDependency is a many2many table used by gorm under the hood
type ComponentConfigDependency struct {
	CreatedByID string                `json:"created_by_id" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgID string `json:"org_id" gorm:"notnull" swaggerignore:"true" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `json:"-" faker:"-" temporaljson:"org,omitzero,omitempty"`

	ComponentConfigConnectionID string                    `json:"component_config_connection_id" gorm:"notnull" swaggerignore:"true" temporaljson:"component_config_connection_id,omitzero,omitempty"`
	ComponentConfigConnection   ComponentConfigConnection `json:"-" faker:"-"`

	ComponentID  string `gorm:"primary_key" temporaljson:"component_id,omitzero,omitempty"`
	DependencyID string `gorm:"primary_key" temporaljson:"dependency_id,omitzero,omitempty"`
}

func (c *ComponentConfigDependency) BeforeSave(tx *gorm.DB) error {
	c.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	c.OrgID = orgIDFromContext(tx.Statement.Context)

	return nil
}
