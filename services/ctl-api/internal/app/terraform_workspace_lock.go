package app

import (
	"time"

	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
)

type TerraformWorkspaceLock struct {
	ID          string  `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string  `json:"created_by_id" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account `json:"-" temporaljson:"created_by,omitzero,omitempty"`

	CreatedAt time.Time `json:"created_at" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt time.Time `json:"updated_at" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`

	// Foreign key to TerraformWorkspace with unique constraint to prevent multiple active locks
	WorkspaceID string             `json:"workspace_id" gorm:"type:text;not null;foreignKey:WorkspaceID;references:ID;uniqueIndex:idx_workspace_active_lock" temporaljson:"workspace_id,omitzero,omitempty"`
	Workspace   TerraformWorkspace `json:"-" temporaljson:"workspace,omitzero,omitempty"`

	Lock *TerraformLock `json:"lock" temporaljson:"lock,omitzero,omitempty"`
}

func (r *TerraformWorkspaceLock) BeforeCreate(tx *gorm.DB) (err error) {
	if r.ID == "" {
		r.ID = domains.NewTerraformWorkspaceLockID()
	}
	if r.CreatedByID == "" {
		r.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}

	return nil
}
