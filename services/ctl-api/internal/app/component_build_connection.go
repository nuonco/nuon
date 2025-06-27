package app

import (
	"time"

	"gorm.io/plugin/soft_delete"
)

// ComponentBuildConnection represents the connection between a component-build and ComponentConfigConnection.
type ComponentBuildConnection struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `gorm:"index" json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	// used for RLS
	OrgID string `json:"org_id,omitzero" gorm:"notnull" swaggerignore:"true" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `json:"-" faker:"-" temporaljson:"org,omitzero,omitempty"`

	AppConfigID string `json:"app_config_id,omitzero" temporaljson:"app_config_id,omitzero,omitempty"`

	ComponentID   string    `json:"component_id,omitzero" gorm:"notnull" temporaljson:"component_id,omitzero,omitempty"`
	ComponentName string    `json:"component_name,omitzero" gorm:"-" temporaljson:"component_name,omitzero,omitempty"`
	Component     Component `json:"-" temporaljson:"component,omitzero,omitempty"`

	ComponentBuildID string         `json:"component_build_id,omitzero" gorm:"not null" temporaljson:"component_build_id,omitzero,omitempty"`
	ComponentBuild   ComponentBuild `json:"-" temporaljson:"component_build,omitzero,omitempty"`

	ComponentConfigConnectionID string                    `json:"component_config_connection_id,omitzero" gorm:"not null" temporaljson:"component_config_connection_id,omitzero,omitempty"`
	ComponentConfigConnection   ComponentConfigConnection `json:"-" temporaljson:"component_config_connection,omitzero,omitempty"`
}
