package providers

import "errors"

// Config validation errors
var (
	ErrMissingClientID     = errors.New("client_id is required")
	ErrMissingClientSecret = errors.New("client_secret is required")
	ErrMissingIssuerURL    = errors.New("issuer_url is required")
	ErrMissingRedirectURL  = errors.New("redirect_url is required")
)

// BaseConfig holds common OAuth configuration fields shared by all providers.
type BaseConfig struct {
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"`
	RedirectURL  string   `json:"redirect_url"`
	Scopes       []string `json:"scopes,omitempty"`
}

// Validate checks that required base fields are present.
func (c *BaseConfig) Validate() error {
	if c.ClientID == "" {
		return ErrMissingClientID
	}
	if c.ClientSecret == "" {
		return ErrMissingClientSecret
	}
	if c.RedirectURL == "" {
		return ErrMissingRedirectURL
	}
	return nil
}

// OpenIDConfig holds configuration for generic OpenID Connect providers.
type OpenIDConfig struct {
	BaseConfig

	// IssuerURL is used for OIDC discovery (/.well-known/openid-configuration)
	IssuerURL string `json:"issuer_url"`

	// Optional: explicit URLs if discovery is not available
	AuthURL     string `json:"auth_url,omitempty"`
	TokenURL    string `json:"token_url,omitempty"`
	UserInfoURL string `json:"userinfo_url,omitempty"`

	// Optional: claims to extract from the ID token or userinfo response
	ClaimsToExtract []string `json:"claims_to_extract,omitempty"`
}

// Validate checks that required OpenID fields are present.
func (c *OpenIDConfig) Validate() error {
	if err := c.BaseConfig.Validate(); err != nil {
		return err
	}
	// Either issuer_url or explicit URLs must be provided
	if c.IssuerURL == "" && (c.AuthURL == "" || c.TokenURL == "") {
		return ErrMissingIssuerURL
	}
	return nil
}

// GoogleConfig holds configuration for Google OAuth.
type GoogleConfig struct {
	BaseConfig

	// HostedDomain restricts login to a specific G Suite domain (optional)
	HostedDomain string `json:"hosted_domain,omitempty"`
}

// Validate checks that required Google fields are present.
func (c *GoogleConfig) Validate() error {
	return c.BaseConfig.Validate()
}

// GitHubConfig holds configuration for GitHub OAuth.
type GitHubConfig struct {
	BaseConfig

	// Organization restricts login to members of specific GitHub orgs (optional)
	AllowedOrgs []string `json:"allowed_orgs,omitempty"`

	// Teams restricts login to members of specific teams (format: "org/team")
	AllowedTeams []string `json:"allowed_teams,omitempty"`

	// GitHub Enterprise support (optional)
	EnterpriseURL string `json:"enterprise_url,omitempty"`
}

// Validate checks that required GitHub fields are present.
func (c *GitHubConfig) Validate() error {
	return c.BaseConfig.Validate()
}
