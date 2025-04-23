package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
)

type TerraformWorkspace struct {
	ID          string  `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string  `json:"created_by_id" gorm:"not null;default:null"`
	CreatedBy   Account `json:"-"`

	CreatedAt time.Time             `json:"created_at" gorm:"notnull"`
	UpdatedAt time.Time             `json:"updated_at" gorm:"notnull"`
	DeletedAt soft_delete.DeletedAt `json:"-"`

	OrgID string `json:"org_id"`
	Org   Org    `json:"-"`

	OwnerID   string `json:"owner_id" gorm:"type:text;check:owner_id_checker,char_length(id)=26;uniqueIndex:idx_owner"`
	OwnerType string `json:"owner_type" gorm:"type:text;uniqueIndex:idx_owner"`

	States      []TerraformState              `faker:"-" json:"-" swaggerignore:"true" gorm:"constraint:OnDelete:CASCADE;"`
	LockHistory []TerraformWorkspaceLockState `faker:"-" json:"-" swaggerignore:"true" gorm:"foreignKey:WorkspaceID;references:ID;constraint:OnDelete:CASCADE;"`
}

func (r *TerraformWorkspace) BeforeCreate(tx *gorm.DB) (err error) {
	if r.ID == "" {
		r.ID = domains.NewTerraformWorkspaceID()
	}

	if r.CreatedByID == "" {
		r.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}

	if r.OrgID == "" {
		r.OrgID = orgIDFromContext(tx.Statement.Context)
	}
	return nil
}
