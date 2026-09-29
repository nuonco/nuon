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

type GitHubProvider struct {
	BaseProvider
	teamWhitelist []string
	userOrgURL    string
	userTeamURL   string
}

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

type GitHubTeamMembershipState struct {
	State string `json:"state"`
	Role  string `json:"role"`
}

type GitHubProviderConfig struct {
	*ProviderConfig
	TeamWhitelist []string
	UserOrgURL    string
	UserTeamURL   string
}

func NewGitHubProvider() *GitHubProvider {
	return &GitHubProvider{
		BaseProvider: BaseProvider{
			name: GitHubProviderName,
		},
		userOrgURL:  GitHubUserOrgURL,
		userTeamURL: GitHubUserTeamURL,
	}
}

func (p *GitHubProvider) Configure(cfg *ProviderConfig) error {
	if cfg.Logger != nil {
		p.log = cfg.Logger
	} else {
		p.log = zap.NewNop()
	}

	if cfg.ClientID == "" {
		return fmt.Errorf("github: client_id is required")
	}
	if cfg.ClientSecret == "" {
		return fmt.Errorf("github: client_secret is required")
	}

	if cfg.AuthURL == "" {
		cfg.AuthURL = github.Endpoint.AuthURL
	}
	if cfg.TokenURL == "" {
		cfg.TokenURL = github.Endpoint.TokenURL
	}
	if cfg.UserInfoURL == "" {
		cfg.UserInfoURL = GitHubUserInfoURL
	}

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

	if len(p.teamWhitelist) > 0 && !slices.Contains(p.oauth2Cfg.Scopes, "read:org") {
		p.oauth2Cfg.Scopes = append(p.oauth2Cfg.Scopes, "read:org")
	}

	return nil
}

func (p *GitHubProvider) GetUserInfo(ctx context.Context, r *http.Request, opts ...oauth2.AuthCodeOption) (*UserInfo, *ProviderTokens, error) {
	code := r.URL.Query().Get("code")
	if code == "" {
		return nil, nil, fmt.Errorf("github: authorization code not found in request")
	}

	client, _, ptokens, err := p.ExchangeCode(ctx, code, opts...)
	if err != nil {
		return nil, nil, fmt.Errorf("github: %w", err)
	}

	p.log.Debug("token exchange successful",
		zap.Int("access_token_len", len(ptokens.AccessToken)))

	data, err := p.FetchUserInfo(ctx, client)
	if err != nil {
		return nil, ptokens, fmt.Errorf("github: %w", err)
	}

	p.log.Debug("userinfo response", zap.String("body", string(data)))

	var ghUser GitHubUserInfo
	if err := json.Unmarshal(data, &ghUser); err != nil {
		return nil, ptokens, fmt.Errorf("github: failed to parse userinfo: %w", err)
	}

	user := &UserInfo{
		Subject:        fmt.Sprintf("%d", ghUser.ID),
		Email:          ghUser.Email,
		EmailVerified:  ghUser.Email != "",
		Name:           ghUser.Name,
		Username:       ghUser.Login,
		Picture:        ghUser.AvatarURL,
		ProviderUserID: fmt.Sprintf("%d", ghUser.ID),
	}

	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err == nil {
		user.RawClaims = raw
	}

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

func (p *GitHubProvider) AuthCodeURL(state string, opts ...oauth2.AuthCodeOption) string {
	return p.oauth2Cfg.AuthCodeURL(state, opts...)
}
