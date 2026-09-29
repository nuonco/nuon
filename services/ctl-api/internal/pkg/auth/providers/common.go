package providers

type UserInfo struct {
	Subject       string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	Username      string `json:"username,omitempty"`
	Picture       string `json:"picture,omitempty"`

	ProviderUserID string `json:"-"`

	RawClaims map[string]any `json:"-"`
}

func (u *UserInfo) PrepareUserData() {
	if u.Username == "" {
		u.Username = u.Email
	}
	if u.ProviderUserID == "" {
		u.ProviderUserID = u.Subject
	}
}

type ProviderTokens struct {
	AccessToken  string
	RefreshToken string
	IDToken      string
	TokenType    string
	Expiry       int64
}

type CustomClaims struct {
	Claims map[string]any
}
