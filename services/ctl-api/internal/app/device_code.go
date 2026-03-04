package app

import (
	"database/sql"
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
)

// AccountIdentity links an account to an identity provider using the IdP's subject identifier.
// This enables secure authentication where users are identified by their stable `sub` claim
// rather than by email (which can change or be reassigned).
type DeviceCode struct {
	ID        string                `gorm:"primarykey" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedAt time.Time             `json:"created_at,omitzero" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt time.Time             `json:"updated_at,omitzero" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	// Account relationship
	AccountID string   `gorm:"not null" json:"account_id,omitzero" temporaljson:"account_id,omitzero,omitempty"`
	Account   *Account `gorm:"constraint:OnDelete:CASCADE" faker:"-" json:"-" temporaljson:"account,omitzero,omitempty"`
	Code      string   `gorm:"not null"`

	ExpiresAt time.Time `gorm:"not null"`      // 2 min from approval
	Consumed  bool      `gorm:"default:false"` // Token issued?
}

func (a *DeviceCode) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = domains.NewDeviceCodeID()
	}
	return nil
}

func (a *DeviceCode) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &DeviceCode{}, "account_id"),
			Columns: []string{
				"account_id",
			},
		},
		{
			Name:        "idx_device_code_account",
			Columns:     []string{"deleted_at", "account_id", "code"},
			UniqueValue: sql.NullBool{Bool: true, Valid: true},
		},
		{
			Name:        indexes.Name(db, &DeviceCode{}, "code"),
			Columns:     []string{"code"},
			UniqueValue: sql.NullBool{Bool: true, Valid: true},
		},
	}
}
