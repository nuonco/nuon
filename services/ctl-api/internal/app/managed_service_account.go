package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
)

type ManagedServiceAccountPurpose string

const ManagedServiceAccountPurposeStack ManagedServiceAccountPurpose = "stack"

type ManagedServiceAccount struct {
	AccountID string  `gorm:"primarykey" json:"account_id,omitzero" temporaljson:"account_id,omitzero,omitempty"`
	Account   Account `gorm:"constraint:OnDelete:CASCADE" json:"-" temporaljson:"-"`

	CreatedByID string                `gorm:"notnull" json:"created_by_id,omitzero" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"-"`
	CreatedAt   time.Time             `gorm:"notnull" json:"created_at,omitzero" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `gorm:"notnull" json:"updated_at,omitzero" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgID string `gorm:"notnull" json:"org_id,omitzero" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `json:"-" temporaljson:"-"`

	OwnerType   string                       `gorm:"notnull" json:"owner_type,omitzero" temporaljson:"owner_type,omitzero,omitempty"`
	OwnerID     string                       `gorm:"notnull" json:"owner_id,omitzero" temporaljson:"owner_id,omitzero,omitempty"`
	Purpose     ManagedServiceAccountPurpose `gorm:"notnull" json:"purpose,omitzero" temporaljson:"purpose,omitzero,omitempty"`
	InstanceKey string                       `gorm:"notnull" json:"instance_key,omitzero" temporaljson:"instance_key,omitzero,omitempty"`

	PrivateRoleID *string `json:"-" temporaljson:"private_role_id,omitempty"`
	PrivateRole   *Role   `json:"-" temporaljson:"-"`
}

func (a *ManagedServiceAccount) BeforeCreate(tx *gorm.DB) error {
	if a.CreatedByID == "" {
		a.CreatedByID = createdByIDFromContext(tx.Statement.Context)
		if a.CreatedByID == "" {
			a.CreatedByID = a.AccountID
		}
	}
	return nil
}

func (a *ManagedServiceAccount) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, a, "org_id"), Columns: []string{"org_id"},
		},
		{
			Name: indexes.Name(db, a, "owner"), Columns: []string{"owner_type", "owner_id"},
		},
		{
			Name:        indexes.Name(db, a, "identity_slot"),
			Columns:     []string{"org_id", "owner_type", "owner_id", "purpose", "instance_key"},
			UniqueValue: generics.NewNullBool(true),
			Option:      "WHERE deleted_at = 0",
		},
		{
			Name:        indexes.Name(db, a, "private_role_id"),
			Columns:     []string{"private_role_id"},
			UniqueValue: generics.NewNullBool(true),
		},
	}
}
