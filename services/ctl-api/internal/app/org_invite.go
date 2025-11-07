package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/migrations"
)

type OrgInviteStatus string

const (
	OrgInviteStatusPending  OrgInviteStatus = "pending"
	OrgInviteStatusAccepted OrgInviteStatus = "accepted"
)

type OrgInvite struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" gorm:"index:idx_invite_org_email,unique" temporaljson:"deleted_at,omitzero,omitempty"`

	// parent relationship
	OrgID string `json:"org_id,omitzero" gorm:"notnull;index:idx_invite_org_email,unique" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `gorm:"constraint:OnDelete:CASCADE;" json:"-" temporaljson:"org,omitzero,omitempty"`

	Email    string          `gorm:"notnull;default null;index:idx_invite_org_email,unique" json:"email,omitzero" temporaljson:"email,omitzero,omitempty"`
	Status   OrgInviteStatus `json:"status,omitzero" gorm:"notnull;default null" temporaljson:"status,omitzero,omitempty"`
	RoleType RoleType        `json:"role_type,omitzero" temporaljson:"role_type,omitzero,omitempty"`
}

func (o *OrgInvite) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &OrgInvite{}, "org_id"),
			Columns: []string{
				"org_id",
			},
		},
	}
}

func (o *OrgInvite) BeforeCreate(tx *gorm.DB) error {
	o.ID = domains.NewOrgID()
	if o.CreatedByID == "" {
		o.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	return nil
}
