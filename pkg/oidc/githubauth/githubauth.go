// Package githubauth implements the GitHub App user-authorization web flow:
// building the authorize URL, exchanging the returned code for a user access
// token, and checking which app installations that user can see.
//
// GitHub documents that the installation_id on a setup-URL redirect can be
// spoofed and must be confirmed against the installing user's own token. This
// package is the shared piece of that check so the dashboard server (which
// starts the flow) and ctl-api (which finishes it) agree on the protocol.
package githubauth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-github/v50/github"
	"golang.org/x/oauth2"
)

const (
	DefaultAuthorizeURL = "https://github.com/login/oauth/authorize"
	DefaultTokenURL     = "https://github.com/login/oauth/access_token"
	DefaultAPIBaseURL   = "https://api.github.com/"

	defaultTimeout = 15 * time.Second
)

// ErrInstallationNotOwned is returned when the authorizing user cannot see the
// requested installation.
var ErrInstallationNotOwned = errors.New("github installation is not accessible to the authorizing user")

// ErrNotConfigured is returned when the client has no client ID or secret.
var ErrNotConfigured = errors.New("github oauth client is not configured")

// AuthorizeURL builds the web-flow authorize URL for a GitHub App. No scopes
// are requested: a GitHub App user token is limited by the app's own
// permissions, and listing the user's installations needs none.
func AuthorizeURL(clientID, redirectURI, state string) string {
	q := url.Values{}
	q.Set("client_id", clientID)
	q.Set("redirect_uri", redirectURI)
	q.Set("state", state)
	return DefaultAuthorizeURL + "?" + q.Encode()
}

// Client exchanges codes and queries installations on behalf of a user.
type Client struct {
	ClientID     string
	ClientSecret string

	// HTTPClient, TokenURL and APIBaseURL default to production GitHub when
	// unset. Tests point them at an httptest server.
	HTTPClient *http.Client
	TokenURL   string
	APIBaseURL string
}

type tokenResponse struct {
	AccessToken      string `json:"access_token"`
	TokenType        string `json:"token_type"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

// ExchangeCode trades a web-flow code for a user access token.
func (c *Client) ExchangeCode(ctx context.Context, code string) (string, error) {
	if c.ClientID == "" || c.ClientSecret == "" {
		return "", ErrNotConfigured
	}
	if code == "" {
		return "", errors.New("github oauth code is empty")
	}

	form := url.Values{}
	form.Set("client_id", c.ClientID)
	form.Set("client_secret", c.ClientSecret)
	form.Set("code", code)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.tokenURL(), strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("unable to build token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient().Do(req)
	if err != nil {
		return "", fmt.Errorf("unable to exchange github oauth code: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<16))
	if err != nil {
		return "", fmt.Errorf("unable to read token response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("github token endpoint returned %d", resp.StatusCode)
	}

	var tok tokenResponse
	if err := json.Unmarshal(body, &tok); err != nil {
		return "", fmt.Errorf("unable to decode token response: %w", err)
	}
	// GitHub answers 200 with an error object for a bad or reused code.
	if tok.Error != "" {
		return "", fmt.Errorf("github rejected oauth code: %s: %s", tok.Error, tok.ErrorDescription)
	}
	if tok.AccessToken == "" {
		return "", errors.New("github token response had no access_token")
	}
	return tok.AccessToken, nil
}

// UserInstallationIDs lists the app installations the token's user can see.
func (c *Client) UserInstallationIDs(ctx context.Context, token string) ([]int64, error) {
	gh := github.NewClient(oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})))
	if base := c.apiBaseURL(); base != DefaultAPIBaseURL {
		u, err := url.Parse(base)
		if err != nil {
			return nil, fmt.Errorf("invalid github api base url: %w", err)
		}
		gh.BaseURL = u
	}

	var ids []int64
	opts := &github.ListOptions{PerPage: 100}
	for {
		installs, resp, err := gh.Apps.ListUserInstallations(ctx, opts)
		if err != nil {
			return nil, fmt.Errorf("unable to list user installations: %w", err)
		}
		for _, inst := range installs {
			ids = append(ids, inst.GetID())
		}
		if resp.NextPage == 0 {
			return ids, nil
		}
		opts.Page = resp.NextPage
	}
}

// VerifyInstallation exchanges code and confirms installID is among the
// installations the resulting user can see. It returns
// ErrInstallationNotOwned when it is not.
func (c *Client) VerifyInstallation(ctx context.Context, code, installID string) error {
	want, err := strconv.ParseInt(installID, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid github installation id %q: %w", installID, err)
	}

	token, err := c.ExchangeCode(ctx, code)
	if err != nil {
		return err
	}

	ids, err := c.UserInstallationIDs(ctx, token)
	if err != nil {
		return err
	}
	for _, id := range ids {
		if id == want {
			return nil
		}
	}
	return ErrInstallationNotOwned
}

func (c *Client) httpClient() *http.Client {
	if c.HTTPClient != nil {
		return c.HTTPClient
	}
	return &http.Client{Timeout: defaultTimeout}
}

func (c *Client) tokenURL() string {
	if c.TokenURL != "" {
		return c.TokenURL
	}
	return DefaultTokenURL
}

func (c *Client) apiBaseURL() string {
	if c.APIBaseURL != "" {
		return c.APIBaseURL
	}
	return DefaultAPIBaseURL
}
