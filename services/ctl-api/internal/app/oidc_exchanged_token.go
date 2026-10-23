package app

import (
	"database/sql"
	"time"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
)

// OIDCExchangedToken records the SHA-256 of an OIDC token that has already
// been exchanged for a Nuon API token, so the same token cannot be exchanged
// twice.
type OIDCExchangedToken struct {
	ID        string    `gorm:"primarykey" json:"id,omitzero"`
	CreatedAt time.Time `json:"created_at,omitzero"`

	OrgID     string `gorm:"not null" json:"org_id,omitzero"`
	IssuerURL string `gorm:"not null" json:"issuer_url,omitzero"`
	TokenHash string `gorm:"not null" json:"token_hash,omitzero"`

	// ExpiresAt is the source OIDC token's own exp claim: once past, the
	// token can no longer verify, so the row no longer needs to be checked.
	ExpiresAt time.Time `gorm:"not null" json:"expires_at,omitzero"`
}

func (t *OIDCExchangedToken) BeforeCreate(tx *gorm.DB) error {
	if t.ID == "" {
		t.ID = domains.NewOIDCExchangedTokenID()
	}

	return nil
}

func (t *OIDCExchangedToken) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name:        indexes.Name(db, t, "org_issuer_token_hash"),
			Columns:     []string{"org_id", "issuer_url", "token_hash"},
			UniqueValue: sql.NullBool{Bool: true, Valid: true},
		},
		{
			Name:    indexes.Name(db, t, "expires_at"),
			Columns: []string{"expires_at"},
		},
	}
}
