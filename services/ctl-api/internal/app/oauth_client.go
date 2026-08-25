package app

import (
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/nuonco/nuon/pkg/shortid/domains"
)

// OAuthClient is a client registered via OAuth 2.0 Dynamic Client Registration
// (RFC 7591). MCP clients (e.g. Claude Code) self-register as public clients
// that authenticate with PKCE and hold no client secret.
type OAuthClient struct {
	ID        string                `gorm:"primarykey" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedAt time.Time             `json:"created_at,omitzero" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt time.Time             `json:"updated_at,omitzero" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	ClientName   string         `gorm:"not null" json:"client_name,omitzero" temporaljson:"client_name,omitzero,omitempty"`
	RedirectURIs pq.StringArray `gorm:"type:text[];not null" json:"redirect_uris,omitzero" temporaljson:"redirect_uris,omitzero,omitempty" swaggertype:"array,string"`

	// TokenEndpointAuthMethod is always "none" for the public clients MCP uses.
	TokenEndpointAuthMethod string `gorm:"not null;default:none" json:"token_endpoint_auth_method,omitzero" temporaljson:"token_endpoint_auth_method,omitzero,omitempty"`
}

func (a *OAuthClient) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = domains.NewOAuthClientID()
	}
	return nil
}

// AllowsRedirectURI reports whether uri exactly matches one of the client's
// registered redirect URIs.
func (a *OAuthClient) AllowsRedirectURI(uri string) bool {
	for _, r := range a.RedirectURIs {
		if r == uri {
			return true
		}
	}
	return false
}
