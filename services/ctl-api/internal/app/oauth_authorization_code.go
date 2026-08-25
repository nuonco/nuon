package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/nuonco/nuon/pkg/shortid/domains"
)

// OAuthAuthorizationCode backs the OAuth 2.0 authorization-code + PKCE flow
// (RFC 6749 §4.1, RFC 7636). A row is created when a client hits /oauth/authorize
// (before the user has logged in), then completed with an AccountID and a Code once
// the user authenticates via the existing identity-provider login flow.
type OAuthAuthorizationCode struct {
	ID        string                `gorm:"primarykey" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedAt time.Time             `json:"created_at,omitzero" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt time.Time             `json:"updated_at,omitzero" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	// RequestID is the opaque handle threaded through the login redirect so the
	// flow can be resumed at /oauth/finish after the user authenticates.
	RequestID string `gorm:"unique;not null" json:"request_id,omitzero" temporaljson:"request_id,omitzero,omitempty"`

	// Code is the authorization code returned to the client. Empty until the user
	// completes login at /oauth/finish.
	Code string `gorm:"index" json:"-" temporaljson:"code,omitzero,omitempty"`

	ClientID    string `gorm:"not null" json:"client_id,omitzero" temporaljson:"client_id,omitzero,omitempty"`
	RedirectURI string `gorm:"not null" json:"redirect_uri,omitzero" temporaljson:"redirect_uri,omitzero,omitempty"`

	CodeChallenge       string `gorm:"not null" json:"-" temporaljson:"code_challenge,omitzero,omitempty"`
	CodeChallengeMethod string `gorm:"not null" json:"code_challenge_method,omitzero" temporaljson:"code_challenge_method,omitzero,omitempty"`

	Scope string `json:"scope,omitzero" temporaljson:"scope,omitzero,omitempty"`

	// ClientState is the client's `state` param, echoed back on the redirect.
	ClientState string `json:"client_state,omitzero" temporaljson:"client_state,omitzero,omitempty"`

	// Resource is the RFC 8707 resource indicator (the MCP server URL).
	Resource string `json:"resource,omitzero" temporaljson:"resource,omitzero,omitempty"`

	// AccountID is set once the user authenticates. No foreign key: the row is
	// created before login, when account_id is still empty, which would violate
	// an accounts FK.
	AccountID string `gorm:"index" json:"account_id,omitzero" temporaljson:"account_id,omitzero,omitempty"`

	ExpiresAt time.Time `gorm:"not null" json:"expires_at,omitzero" temporaljson:"expires_at,omitzero,omitempty"`
	Consumed  bool      `gorm:"default:false" json:"consumed,omitempty" temporaljson:"consumed,omitempty"`
}

func (a *OAuthAuthorizationCode) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = domains.NewOAuthAuthorizationCodeID()
	}
	return nil
}
