package app

import (
	"time"

	"gorm.io/plugin/soft_delete"
)

type VCSConnectionPullRequest struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `gorm:"index" json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	VCSConnectionID string        `json:"vcs_connection_id,omitzero" gorm:"notnull;index:idx_app_component_name,unique" temporaljson:"app_id,omitzero,omitempty"`
	VCSConnection   VCSConnection `faker:"-" json:"-" temporaljson:"vcs_connection,omitzero,omitempty"`
}
