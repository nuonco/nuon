package providers

// UserInfo represents the authenticated user's information.
type UserInfo struct {
	// Standard claims
	Subject       string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	Username      string `json:"username,omitempty"`
	Picture       string `json:"picture,omitempty"`

	// Provider-specific identifier
	ProviderUserID string `json:"-"`

	// Raw claims from the provider (for custom claim extraction)
	RawClaims map[string]any `json:"-"`
}

// PrepareUserData ensures required fields are populated with fallbacks.
func (u *UserInfo) PrepareUserData() {
	if u.Username == "" {
		u.Username = u.Email
	}
	if u.ProviderUserID == "" {
		u.ProviderUserID = u.Subject
	}
}

// ProviderTokens holds the tokens received from the OAuth provider.
type ProviderTokens struct {
	AccessToken  string
	RefreshToken string
	IDToken      string
	TokenType    string
	Expiry       int64 // Unix timestamp
}

// CustomClaims holds additional claims extracted from the provider response.
type CustomClaims struct {
	Claims map[string]any
}
