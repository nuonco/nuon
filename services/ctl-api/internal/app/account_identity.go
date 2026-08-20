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
	AccountID string   `gorm:"not null;index:idx_account_identity_account_idp,unique" json:"account_id,omitzero" temporaljson:"account_id,omitzero,omitempty"`
	Account   *Account `gorm:"constraint:OnDelete:CASCADE" faker:"-" json:"-" temporaljson:"account,omitzero,omitempty"`

	// IdentityProviderID is either an identity_providers row ID or, for a provider configured
	// through env vars, the sentinel returned by EnvIdentityProviderID. There is deliberately no
	// FK: the env provider has no row, and ON DELETE SET NULL would collapse a deleted provider's
	// identities onto each other. NOT NULL is applied by migration rather than by tag, because
	// AutoMigrate runs before the backfill that fills in the legacy NULLs.
	IdentityProviderID string `gorm:"index:idx_account_identity_account_idp,unique;index:idx_account_identity_idp_sub,unique" json:"identity_provider_id,omitempty" temporaljson:"identity_provider_id,omitzero,omitempty"`

	// ProviderType is retained for display and for grouping identities by protocol. It no longer
	// identifies a provider on its own - several providers can share a type.
	ProviderType ProviderType `gorm:"not null" json:"provider_type,omitzero" temporaljson:"provider_type,omitzero,omitempty"`

	// Subject identifier from the IdP - the canonical, stable user identifier
	Sub string `gorm:"not null;index:idx_account_identity_idp_sub,unique" json:"sub,omitzero" temporaljson:"sub,omitzero,omitempty"`

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
			Name: indexes.Name(db, &AccountIdentity{}, "identity_provider_id"),
			Columns: []string{
				"identity_provider_id",
			},
		},
	}
}
