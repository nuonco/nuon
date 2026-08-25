package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/nuonco/nuon/pkg/shortid/domains"
)

// OAuthRefreshToken backs the OAuth 2.0 refresh-token grant (RFC 6749 §6). Refresh
// tokens are rotated on use: the old row is marked consumed and a new one issued.
type OAuthRefreshToken struct {
	ID        string                `gorm:"primarykey" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedAt time.Time             `json:"created_at,omitzero" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt time.Time             `json:"updated_at,omitzero" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	Token    string `gorm:"unique;not null" json:"-" temporaljson:"token,omitzero,omitempty"`
	ClientID string `gorm:"not null" json:"client_id,omitzero" temporaljson:"client_id,omitzero,omitempty"`
	Scope    string `json:"scope,omitzero" temporaljson:"scope,omitzero,omitempty"`

	AccountID string   `gorm:"not null;index" json:"account_id,omitzero" temporaljson:"account_id,omitzero,omitempty"`
	Account   *Account `gorm:"constraint:OnDelete:CASCADE" faker:"-" json:"-" temporaljson:"account,omitzero,omitempty"`

	ExpiresAt time.Time `gorm:"not null" json:"expires_at,omitzero" temporaljson:"expires_at,omitzero,omitempty"`
	Consumed  bool      `gorm:"default:false" json:"consumed,omitempty" temporaljson:"consumed,omitempty"`
}

func (a *OAuthRefreshToken) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = domains.NewOAuthRefreshTokenID()
	}
	return nil
}
