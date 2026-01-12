package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strings"

	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

const (
	GitHubProviderName = "github"
	GitHubUserInfoURL  = "https://api.github.com/user"
	GitHubUserOrgURL   = "https://api.github.com/orgs/:org_id/members/:username"
	GitHubUserTeamURL  = "https://api.github.com/orgs/:org_id/teams/:team_slug/memberships/:username"
)

// GitHubProvider implements the Provider interface for GitHub OAuth.
type GitHubProvider struct {
	BaseProvider
	teamWhitelist []string // org or org/team format
	userOrgURL    string
	userTeamURL   string
}

// GitHubUserInfo represents user information from GitHub's API.
type GitHubUserInfo struct {
	ID        int    `json:"id"`
	Login     string `json:"login"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
	HTMLURL   string `json:"html_url"`
	Type      string `json:"type"`
	Company   string `json:"company"`
	Blog      string `json:"blog"`
	Location  string `json:"location"`
	Bio       string `json:"bio"`
}

// GitHubTeamMembershipState represents the team membership response from GitHub.
type GitHubTeamMembershipState struct {
	State string `json:"state"` // "active" or "pending"
	Role  string `json:"role"`  // "member" or "maintainer"
}

// GitHubProviderConfig extends ProviderConfig with GitHub-specific options.
type GitHubProviderConfig struct {
	*ProviderConfig
	TeamWhitelist []string // List of org or org/team to check membership
	UserOrgURL    string   // URL template for checking org membership
	UserTeamURL   string   // URL template for checking team membership
}

// NewGitHubProvider creates a new GitHub OAuth provider instance.
func NewGitHubProvider() *GitHubProvider {
	return &GitHubProvider{
		BaseProvider: BaseProvider{
			name: GitHubProviderName,
		},
		userOrgURL:  GitHubUserOrgURL,
		userTeamURL: GitHubUserTeamURL,
	}
}

// Configure initializes the GitHub provider with the given configuration.
func (p *GitHubProvider) Configure(cfg *ProviderConfig) error {
	if cfg.Logger != nil {
		p.log = cfg.Logger
	} else {
		p.log = zap.NewNop()
	}

	// Validate required configuration
	if cfg.ClientID == "" {
		return fmt.Errorf("github: client_id is required")
	}
	if cfg.ClientSecret == "" {
		return fmt.Errorf("github: client_secret is required")
	}

	// Set GitHub-specific defaults
	if cfg.AuthURL == "" {
		cfg.AuthURL = github.Endpoint.AuthURL
	}
	if cfg.TokenURL == "" {
		cfg.TokenURL = github.Endpoint.TokenURL
	}
	if cfg.UserInfoURL == "" {
		cfg.UserInfoURL = GitHubUserInfoURL
	}

	// Set default scopes if not provided
	if len(cfg.Scopes) == 0 {
		cfg.Scopes = []string{
			"user:email",
			"read:user",
		}
	}

	p.SetupOAuth2Config(cfg)
	p.name = GitHubProviderName

	p.log.Info("GitHub provider configured",
		zap.String("userinfo_url", cfg.UserInfoURL),
		zap.Strings("scopes", cfg.Scopes))

	return nil
}

// ConfigureWithTeams initializes the GitHub provider with team/org membership checking.
func (p *GitHubProvider) ConfigureWithTeams(cfg *GitHubProviderConfig) error {
	if err := p.Configure(cfg.ProviderConfig); err != nil {
		return err
	}

	p.teamWhitelist = cfg.TeamWhitelist
	if cfg.UserOrgURL != "" {
		p.userOrgURL = cfg.UserOrgURL
	}
	if cfg.UserTeamURL != "" {
		p.userTeamURL = cfg.UserTeamURL
	}

	// Add org scope if checking teams
	if len(p.teamWhitelist) > 0 && !slices.Contains(p.oauth2Cfg.Scopes, "read:org") {
		p.oauth2Cfg.Scopes = append(p.oauth2Cfg.Scopes, "read:org")
	}

	return nil
}

// GetUserInfo exchanges the authorization code for tokens and retrieves user information.
func (p *GitHubProvider) GetUserInfo(ctx context.Context, r *http.Request, opts ...oauth2.AuthCodeOption) (*UserInfo, *ProviderTokens, error) {
	code := r.URL.Query().Get("code")
	if code == "" {
		return nil, nil, fmt.Errorf("github: authorization code not found in request")
	}

	// Exchange the code for tokens
	client, _, ptokens, err := p.ExchangeCode(ctx, code, opts...)
	if err != nil {
		return nil, nil, fmt.Errorf("github: %w", err)
	}

	p.log.Debug("token exchange successful",
		zap.Int("access_token_len", len(ptokens.AccessToken)))

	// Fetch user info from GitHub's API
	data, err := p.FetchUserInfo(ctx, client)
	if err != nil {
		return nil, ptokens, fmt.Errorf("github: %w", err)
	}

	p.log.Debug("userinfo response", zap.String("body", string(data)))

	// Parse GitHub-specific response
	var ghUser GitHubUserInfo
	if err := json.Unmarshal(data, &ghUser); err != nil {
		return nil, ptokens, fmt.Errorf("github: failed to parse userinfo: %w", err)
	}

	// Convert to standard UserInfo
	user := &UserInfo{
		Subject:        fmt.Sprintf("%d", ghUser.ID),
		Email:          ghUser.Email,
		EmailVerified:  ghUser.Email != "", // GitHub doesn't explicitly tell us, assume verified if present
		Name:           ghUser.Name,
		Username:       ghUser.Login,
		Picture:        ghUser.AvatarURL,
		ProviderUserID: fmt.Sprintf("%d", ghUser.ID),
	}

	// Store raw claims
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err == nil {
		user.RawClaims = raw
	}

	// Check team/org memberships if configured
	if len(p.teamWhitelist) > 0 {
		memberships, err := p.checkTeamMemberships(ctx, client, user.Username)
		if err != nil {
			p.log.Warn("failed to check team memberships", zap.Error(err))
		} else {
			user.RawClaims["team_memberships"] = memberships
		}
	}

	return user, ptokens, nil
}

// checkTeamMemberships verifies the user's membership in configured orgs/teams.
func (p *GitHubProvider) checkTeamMemberships(ctx context.Context, client *http.Client, username string) ([]string, error) {
	var memberships []string

	for _, orgAndTeam := range p.teamWhitelist {
		org, team := p.parseOrgAndTeam(orgAndTeam)
		if org == "" {
			p.log.Warn("invalid org/team format", zap.String("value", orgAndTeam))
			continue
		}

		var isMember bool
		var err error

		if team != "" {
			isMember, err = p.checkTeamMembership(ctx, client, username, org, team)
		} else {
			isMember, err = p.checkOrgMembership(ctx, client, username, org)
		}

		if err != nil {
			p.log.Warn("membership check failed",
				zap.String("org", org),
				zap.String("team", team),
				zap.Error(err))
			continue
		}

		if isMember {
			memberships = append(memberships, orgAndTeam)
		}
	}

	return memberships, nil
}

// parseOrgAndTeam splits "org" or "org/team" format.
func (p *GitHubProvider) parseOrgAndTeam(orgAndTeam string) (string, string) {
	parts := strings.Split(orgAndTeam, "/")
	switch len(parts) {
	case 1:
		return parts[0], ""
	case 2:
		return parts[0], parts[1]
	default:
		return "", ""
	}
}

// checkOrgMembership checks if the user is a member of the organization.
func (p *GitHubProvider) checkOrgMembership(ctx context.Context, client *http.Client, username, org string) (bool, error) {
	url := strings.NewReplacer(":org_id", org, ":username", username).Replace(p.userOrgURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return false, fmt.Errorf("failed to create org membership request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to check org membership: %w", err)
	}
	defer resp.Body.Close()

	// Handle redirect for public membership check
	if resp.StatusCode == http.StatusFound {
		location := resp.Header.Get("Location")
		if location != "" {
			req, err = http.NewRequestWithContext(ctx, http.MethodGet, location, nil)
			if err != nil {
				return false, err
			}
			resp, err = client.Do(req)
			if err != nil {
				return false, err
			}
			defer resp.Body.Close()
		}
	}

	switch resp.StatusCode {
	case http.StatusNoContent:
		return true, nil
	case http.StatusNotFound:
		return false, nil
	default:
		return false, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}

// checkTeamMembership checks if the user is a member of the team.
func (p *GitHubProvider) checkTeamMembership(ctx context.Context, client *http.Client, username, org, team string) (bool, error) {
	url := strings.NewReplacer(":org_id", org, ":team_slug", team, ":username", username).Replace(p.userTeamURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return false, fmt.Errorf("failed to create team membership request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to check team membership: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var state GitHubTeamMembershipState
		if err := json.NewDecoder(resp.Body).Decode(&state); err != nil {
			return false, fmt.Errorf("failed to decode team membership: %w", err)
		}
		return state.State == "active", nil
	case http.StatusNotFound:
		return false, nil
	default:
		return false, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}

// AuthCodeURL returns the URL to redirect the user to for authentication.
func (p *GitHubProvider) AuthCodeURL(state string, opts ...oauth2.AuthCodeOption) string {
	return p.oauth2Cfg.AuthCodeURL(state, opts...)
}
