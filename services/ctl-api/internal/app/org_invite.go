package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
)

type OrgInviteStatus string

const (
	OrgInviteStatusPending  OrgInviteStatus = "pending"
	OrgInviteStatusAccepted OrgInviteStatus = "accepted"
)

type OrgInvite struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" gorm:"index:idx_invite_org_email,unique" temporaljson:"deleted_at,omitzero,omitempty"`

	// parent relationship
	OrgID string `gorm:"notnull;index:idx_invite_org_email,unique" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `gorm:"constraint:OnDelete:CASCADE;" json:"-" temporaljson:"org,omitzero,omitempty"`

	Email    string          `gorm:"notnull;default null;index:idx_invite_org_email,unique" json:"email" temporaljson:"email,omitzero,omitempty"`
	Status   OrgInviteStatus `json:"status" gorm:"notnull;default null" temporaljson:"status,omitzero,omitempty"`
	RoleType RoleType        `json:"role_type" temporaljson:"role_type,omitzero,omitempty"`
}

func (o *OrgInvite) BeforeCreate(tx *gorm.DB) error {
	o.ID = domains.NewOrgID()
	if o.CreatedByID == "" {
		o.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	return nil
}
