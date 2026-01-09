package app

import (
	"time"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
)

// AccountIdentity links an account to an identity provider using the IdP's subject identifier.
// This enables secure authentication where users are identified by their stable `sub` claim
// rather than by email (which can change or be reassigned).
type AccountIdentity struct {
	ID        string    `gorm:"primarykey" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedAt time.Time `json:"created_at,omitzero" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitzero" temporaljson:"updated_at,omitzero,omitempty"`

	// Account relationship
	AccountID string   `gorm:"not null;index:idx_account_identity_account_provider,unique" json:"account_id,omitzero" temporaljson:"account_id,omitzero,omitempty"`
	Account   *Account `gorm:"constraint:OnDelete:CASCADE" faker:"-" json:"-" temporaljson:"account,omitzero,omitempty"`

	// Identity Provider relationship (nullable for default env-var provider)
	IdentityProviderID *string           `gorm:"index" json:"identity_provider_id,omitempty" temporaljson:"identity_provider_id,omitzero,omitempty"`
	IdentityProvider   *IdentityProvider `gorm:"constraint:OnDelete:SET NULL" faker:"-" json:"-" temporaljson:"identity_provider,omitzero,omitempty"`

	// Provider type - required, enables lookup when using the default env-var provider
	ProviderType ProviderType `gorm:"not null;index:idx_account_identity_account_provider,unique;index:idx_account_identity_provider_sub,unique" json:"provider_type,omitzero" temporaljson:"provider_type,omitzero,omitempty"`

	// Subject identifier from the IdP - the canonical, stable user identifier
	Sub string `gorm:"not null;index:idx_account_identity_provider_sub,unique" json:"sub,omitzero" temporaljson:"sub,omitzero,omitempty"`

	// User profile information from the identity provider
	Name    string `json:"name,omitempty" temporaljson:"name,omitempty"`
	Picture string `json:"picture,omitempty" temporaljson:"picture,omitempty"`
}

func (a AccountIdentity) TableName() string {
	return "account_identities"
}

func (a *AccountIdentity) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = domains.NewAccountIdentityID()
	}
	return nil
}

func (a *AccountIdentity) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &AccountIdentity{}, "account_id"),
			Columns: []string{
				"account_id",
			},
		},
		{
			Name: indexes.Name(db, &AccountIdentity{}, "provider_type_sub"),
			Columns: []string{
				"provider_type",
				"sub",
			},
		},
	}
}
