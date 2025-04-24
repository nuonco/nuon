package app

import (
	"time"

	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
)

type TerraformWorkspaceLock struct {
	ID          string  `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string  `json:"created_by_id" gorm:"not null;default:null"`
	CreatedBy   Account `json:"-"`

	CreatedAt time.Time `json:"created_at" gorm:"notnull"`
	UpdatedAt time.Time `json:"updated_at" gorm:"notnull"`

	// Foreign key to TerraformWorkspace with unique constraint to prevent multiple active locks
	WorkspaceID string             `json:"workspace_id" gorm:"type:text;not null;foreignKey:WorkspaceID;references:ID;uniqueIndex:idx_workspace_active_lock"`
	Workspace   TerraformWorkspace `json:"-"`

	Lock *TerraformLock `json:"lock"`
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
