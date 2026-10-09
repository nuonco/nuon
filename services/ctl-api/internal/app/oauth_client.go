package app

import (
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/pkg/shortid/domains"
)

// OAuthClient represents a public PKCE client registered through DCR (RFC 7591)
// or a confidential client provisioned for a service account by an org admin.
type OAuthClient struct {
	ID        string                `gorm:"primarykey" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedAt time.Time             `json:"created_at,omitzero" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt time.Time             `json:"updated_at,omitzero" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	ClientName   string         `gorm:"not null" json:"client_name,omitzero" temporaljson:"client_name,omitzero,omitempty"`
	RedirectURIs pq.StringArray `gorm:"type:text[];not null" json:"redirect_uris,omitzero" temporaljson:"redirect_uris,omitzero,omitempty" swaggertype:"array,string"`

	// Public clients use "none"; service-account clients use "client_secret_basic".
	TokenEndpointAuthMethod string `gorm:"not null;default:none" json:"token_endpoint_auth_method,omitzero" temporaljson:"token_endpoint_auth_method,omitzero,omitempty"`

	OrgID      generics.NullString `gorm:"index" json:"org_id,omitzero" swaggertype:"string" temporaljson:"org_id,omitzero,omitempty"`
	AccountID  generics.NullString `gorm:"index" json:"account_id,omitzero" swaggertype:"string" temporaljson:"account_id,omitzero,omitempty"`
	GrantTypes pq.StringArray      `gorm:"type:text[]" json:"grant_types,omitzero" temporaljson:"grant_types,omitzero,omitempty" swaggertype:"array,string"`

	Secrets       []OAuthClientSecret `gorm:"foreignKey:ClientID" json:"secrets,omitzero" temporaljson:"secrets,omitzero,omitempty"`
	TokenEndpoint string              `gorm:"-" json:"token_endpoint,omitzero" temporaljson:"token_endpoint,omitzero,omitempty"`
}

const (
	OAuthTokenEndpointAuthMethodNone              = "none"
	OAuthTokenEndpointAuthMethodClientSecretBasic = "client_secret_basic"

	OAuthGrantTypeClientCredentials = "client_credentials"
)

func (a *OAuthClient) IsConfidential() bool {
	return a.AccountID.Valid && a.AccountID.String != "" && a.OrgID.Valid && a.OrgID.String != "" && a.TokenEndpointAuthMethod == OAuthTokenEndpointAuthMethodClientSecretBasic
}

func (a *OAuthClient) AllowsGrantType(grantType string) bool {
	for _, g := range a.GrantTypes {
		if g == grantType {
			return true
		}
	}
	return false
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
