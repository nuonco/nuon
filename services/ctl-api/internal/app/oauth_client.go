package app

import (
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/nuonco/nuon/pkg/nuonauth"
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

func (a *OAuthClient) IsNuonCLI() bool {
	return a.ID == nuonauth.CLIClientID
}

// AllowsRedirectURI reports whether uri matches one of the client's registered
// redirect URIs. The Nuon CLI client ignores the port on loopback redirects so
// it can listen on 127.0.0.1:0 (RFC 8252). Every other client is an exact match.
func (a *OAuthClient) AllowsRedirectURI(uri string) bool {
	for _, registered := range a.RedirectURIs {
		if registered == uri {
			return true
		}
		if a.IsNuonCLI() && loopbackRedirectMatch(registered, uri) {
			return true
		}
	}
	return false
}

func loopbackRedirectMatch(registered, requested string) bool {
	reg, err := url.Parse(registered)
	if err != nil {
		return false
	}
	req, err := url.Parse(requested)
	if err != nil {
		return false
	}
	if reg.Scheme != "http" || req.Scheme != "http" {
		return false
	}
	if reg.Path != req.Path || reg.RawQuery != "" || req.RawQuery != "" || reg.Fragment != "" || req.Fragment != "" {
		return false
	}
	return isLoopbackHostname(reg.Hostname()) && isLoopbackHostname(req.Hostname())
}

func isLoopbackHostname(host string) bool {
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}
