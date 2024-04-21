package app

import (
	"time"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

type OrgInviteStatus string

const (
	OrgInviteStatusPending  OrgInviteStatus = "pending"
	OrgInviteStatusAccepted OrgInviteStatus = "accepted"
)

type OrgInvite struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string                `json:"created_by_id" gorm:"not null;default:null"`
	CreatedBy   UserToken             `json:"created_by" gorm:"references:Subject"`
	CreatedAt   time.Time             `json:"created_at" gorm:"notnull"`
	UpdatedAt   time.Time             `json:"updated_at" gorm:"notnull"`
	DeletedAt   soft_delete.DeletedAt `json:"-" gorm:"index:idx_invite_org_email,unique"`

	// parent relationship
	OrgID string `gorm:"notnull;index:idx_invite_org_email,unique"`
	Org   Org    `gorm:"constraint:OnDelete:CASCADE;" json:"-"`

	Email  string          `gorm:"notnull;default null;index:idx_invite_org_email,unique" json:"email"`
	Status OrgInviteStatus `json:"status" gorm:"notnull;default null"`
}

func (o *OrgInvite) BeforeCreate(tx *gorm.DB) error {
	o.ID = domains.NewOrgID()
	if o.CreatedByID == "" {
		o.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	return nil
}
