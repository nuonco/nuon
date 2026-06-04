// Package client is a thin Slack Web API client scoped to the surface
// area the Nuon slackbot integration needs:
//
//   - oauth.v2.access  — exchange OAuth code for a workspace bot token
//   - chat.postMessage — post lifecycle / approval messages
//   - chat.update      — edit a previously posted message (e.g. mark approved)
//   - conversations.list — enumerate channels for the /nuon subscribe flow
//   - auth.test        — probe a token (used after install to capture
//     bot_user_id, app_id, etc.)
//
// We avoid pulling in a third-party Slack SDK because the surface area is
// small, easy to review, and we want full control over JSON shape, retry
// behavior, and observability.
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/slack-go/slack"
)

// BaseURL is the public Slack Web API endpoint. Tests override this.
const BaseURL = "https://slack.com/api"

// defaultTimeout caps individual Slack API requests. Slack typically responds
// in under a second; cap conservatively to avoid blocking webhook handlers.
const defaultTimeout = 10 * time.Second

// Client is a Slack Web API client. Construct via New.
type Client struct {
	httpClient *http.Client
	baseURL    string
}

// Option mutates Client during construction.
type Option func(*Client)

// WithBaseURL overrides the default Slack API base URL (used in tests).
func WithBaseURL(u string) Option {
	return func(c *Client) { c.baseURL = u }
}

// WithHTTPClient overrides the default HTTP client.
func WithHTTPClient(h *http.Client) Option {
	return func(c *Client) { c.httpClient = h }
}

// New constructs a Slack Web API client with sensible defaults.
func New(opts ...Option) *Client {
	c := &Client{
		httpClient: &http.Client{Timeout: defaultTimeout},
		baseURL:    BaseURL,
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

// baseResponse is embedded in every Slack API response.
type baseResponse struct {
	OK    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
}

// OAuthV2AccessRequest contains the inputs for an OAuth code exchange.
type OAuthV2AccessRequest struct {
	ClientID     string
	ClientSecret string
	Code         string
	RedirectURI  string
}

// OAuthV2AccessResponse is the subset of oauth.v2.access we care about.
//
// Slack returns far more fields (incoming_webhook, authed_user, etc.) but we
// only persist what's needed to drive the bot: the bot token, the workspace
// identity, and the install-time scope grant.
type OAuthV2AccessResponse struct {
	baseResponse

	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
	BotUserID   string `json:"bot_user_id"`
	AppID       string `json:"app_id"`

	Team struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"team"`

	Enterprise *struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"enterprise"`

	// IsEnterpriseInstall is true for org-wide (Enterprise Grid) installs.
	// We reject these at the OAuth callback in v1 — see Phase 4.
	IsEnterpriseInstall bool `json:"is_enterprise_install"`

	AuthedUser struct {
		ID string `json:"id"`
	} `json:"authed_user"`
}

// OAuthV2Access exchanges an OAuth code for a workspace access token.
func (c *Client) OAuthV2Access(ctx context.Context, req OAuthV2AccessRequest) (*OAuthV2AccessResponse, error) {
	form := url.Values{}
	form.Set("client_id", req.ClientID)
	form.Set("client_secret", req.ClientSecret)
	form.Set("code", req.Code)
	if req.RedirectURI != "" {
		form.Set("redirect_uri", req.RedirectURI)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/oauth.v2.access", strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("slack: build oauth.v2.access: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	var resp OAuthV2AccessResponse
	if err := c.do(httpReq, &resp); err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, fmt.Errorf("slack oauth.v2.access: %s", resp.Error)
	}
	return &resp, nil
}

// PostMessageRequest is the input for chat.postMessage.
type PostMessageRequest struct {
	Channel  string         `json:"channel"`
	Text     string         `json:"text,omitempty"`
	Blocks   []slack.Block  `json:"blocks,omitempty"`
	ThreadTS string         `json:"thread_ts,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

// PostMessageResponse mirrors chat.postMessage's response.
type PostMessageResponse struct {
	baseResponse

	Channel string `json:"channel"`
	TS      string `json:"ts"`
}

// PostMessage posts a message to a channel as the supplied bot token.
func (c *Client) PostMessage(ctx context.Context, botToken string, req PostMessageRequest) (*PostMessageResponse, error) {
	var resp PostMessageResponse
	if err := c.callJSON(ctx, "chat.postMessage", botToken, req, &resp); err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, fmt.Errorf("slack chat.postMessage: %s", resp.Error)
	}
	return &resp, nil
}

// UpdateMessageRequest mutates a previously posted message.
type UpdateMessageRequest struct {
	Channel string        `json:"channel"`
	TS      string        `json:"ts"`
	Text    string        `json:"text,omitempty"`
	Blocks  []slack.Block `json:"blocks,omitempty"`
}

// UpdateMessageResponse mirrors chat.update's response.
type UpdateMessageResponse struct {
	baseResponse

	Channel string `json:"channel"`
	TS      string `json:"ts"`
}

// UpdateMessage edits a previously posted message.
func (c *Client) UpdateMessage(ctx context.Context, botToken string, req UpdateMessageRequest) (*UpdateMessageResponse, error) {
	var resp UpdateMessageResponse
	if err := c.callJSON(ctx, "chat.update", botToken, req, &resp); err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, fmt.Errorf("slack chat.update: %s", resp.Error)
	}
	return &resp, nil
}

// ConversationsListRequest paginates through channel listings.
type ConversationsListRequest struct {
	Cursor          string
	Limit           int
	Types           string // e.g. "public_channel,private_channel"
	ExcludeArchived bool
}

// Conversation is a single channel entry from conversations.list.
type Conversation struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	IsChannel  bool   `json:"is_channel"`
	IsPrivate  bool   `json:"is_private"`
	IsArchived bool   `json:"is_archived"`
	IsMember   bool   `json:"is_member"`
}

// ConversationsListResponse is the response from conversations.list.
type ConversationsListResponse struct {
	baseResponse

	Channels         []Conversation `json:"channels"`
	ResponseMetadata struct {
		NextCursor string `json:"next_cursor"`
	} `json:"response_metadata"`
}

// ConversationsList enumerates channels visible to the bot.
func (c *Client) ConversationsList(ctx context.Context, botToken string, req ConversationsListRequest) (*ConversationsListResponse, error) {
	q := url.Values{}
	if req.Cursor != "" {
		q.Set("cursor", req.Cursor)
	}
	if req.Limit > 0 {
		q.Set("limit", fmt.Sprintf("%d", req.Limit))
	}
	if req.Types != "" {
		q.Set("types", req.Types)
	}
	if req.ExcludeArchived {
		q.Set("exclude_archived", "true")
	}

	endpoint := c.baseURL + "/conversations.list"
	if encoded := q.Encode(); encoded != "" {
		endpoint += "?" + encoded
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("slack: build conversations.list: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+botToken)

	var resp ConversationsListResponse
	if err := c.do(httpReq, &resp); err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, fmt.Errorf("slack conversations.list: %s", resp.Error)
	}
	return &resp, nil
}

type ConversationsInfoResponse struct {
	baseResponse

	Channel Conversation `json:"channel"`
}

// ConversationsInfo fetches metadata for a single channel by ID.
func (c *Client) ConversationsInfo(ctx context.Context, botToken, channelID string) (*ConversationsInfoResponse, error) {
	q := url.Values{}
	q.Set("channel", channelID)

	endpoint := c.baseURL + "/conversations.info?" + q.Encode()
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("slack: build conversations.info: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+botToken)

	var resp ConversationsInfoResponse
	if err := c.do(httpReq, &resp); err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, fmt.Errorf("slack conversations.info: %s", resp.Error)
	}
	return &resp, nil
}

// ViewsOpenRequest opens a modal in response to a trigger_id from an
// interactive surface (slash command, button click, etc.). View is the
// Block-Kit modal definition; we accept it as a generic map so callers can
// build whatever shape they need without us schema-locking the modal here.
type ViewsOpenRequest struct {
	TriggerID string         `json:"trigger_id"`
	View      map[string]any `json:"view"`
}

// ViewsResponse is the shared response shape for views.open / views.update /
// views.push. Slack returns the persisted view (with its allocated id) on
// success.
type ViewsResponse struct {
	baseResponse

	View struct {
		ID         string `json:"id"`
		CallbackID string `json:"callback_id"`
		Hash       string `json:"hash"`
	} `json:"view"`

	// ResponseMetadata.Messages carries Block-Kit validation errors when
	// Slack rejects a view payload (e.g. an unknown block type). Surfaces
	// in error messages so misshapen modals are debuggable.
	ResponseMetadata struct {
		Messages []string `json:"messages,omitempty"`
	} `json:"response_metadata,omitempty"`
}

// ViewsOpen opens a new modal anchored to the given trigger_id. Trigger ids
// are short-lived (~3s) so callers must not block before invoking this.
func (c *Client) ViewsOpen(ctx context.Context, botToken string, req ViewsOpenRequest) (*ViewsResponse, error) {
	var resp ViewsResponse
	if err := c.callJSON(ctx, "views.open", botToken, req, &resp); err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, fmt.Errorf("slack views.open: %s%s", resp.Error, formatViewMessages(resp.ResponseMetadata.Messages))
	}
	return &resp, nil
}

// ViewsUpdateRequest replaces the contents of an already-open modal. Exactly
// one of ViewID / ExternalID identifies the view; ViewID is the standard
// case (Slack provides it on every view payload).
type ViewsUpdateRequest struct {
	ViewID     string         `json:"view_id,omitempty"`
	ExternalID string         `json:"external_id,omitempty"`
	Hash       string         `json:"hash,omitempty"`
	View       map[string]any `json:"view"`
}

// ViewsUpdate edits a previously opened modal in place. Used by the
// unsubscribe modal's Remove buttons (re-render after each delete) and by
// the subscribe modal's scope/install pivots.
func (c *Client) ViewsUpdate(ctx context.Context, botToken string, req ViewsUpdateRequest) (*ViewsResponse, error) {
	var resp ViewsResponse
	if err := c.callJSON(ctx, "views.update", botToken, req, &resp); err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, fmt.Errorf("slack views.update: %s%s", resp.Error, formatViewMessages(resp.ResponseMetadata.Messages))
	}
	return &resp, nil
}

// ViewsPushRequest pushes a new modal onto an already-open modal stack. We
// don't currently use push (everything fits in a single root modal) but
// expose it for parity with views.open / views.update.
type ViewsPushRequest struct {
	TriggerID string         `json:"trigger_id"`
	View      map[string]any `json:"view"`
}

// ViewsPush stacks a new modal on top of the current one.
func (c *Client) ViewsPush(ctx context.Context, botToken string, req ViewsPushRequest) (*ViewsResponse, error) {
	var resp ViewsResponse
	if err := c.callJSON(ctx, "views.push", botToken, req, &resp); err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, fmt.Errorf("slack views.push: %s%s", resp.Error, formatViewMessages(resp.ResponseMetadata.Messages))
	}
	return &resp, nil
}

// formatViewMessages flattens Slack's response_metadata.messages array into a
// single trailing " (msg1; msg2)" suffix for log/error-friendly inclusion.
func formatViewMessages(msgs []string) string {
	if len(msgs) == 0 {
		return ""
	}
	return " (" + strings.Join(msgs, "; ") + ")"
}

// AuthTestResponse is the response from auth.test.
type AuthTestResponse struct {
	baseResponse

	URL          string `json:"url"`
	Team         string `json:"team"`
	User         string `json:"user"`
	TeamID       string `json:"team_id"`
	UserID       string `json:"user_id"`
	BotID        string `json:"bot_id"`
	EnterpriseID string `json:"enterprise_id,omitempty"`
}

// AuthTest probes a bot token. Useful immediately after install to verify
// scope grants and capture the bot user id / team id from the token itself.
func (c *Client) AuthTest(ctx context.Context, botToken string) (*AuthTestResponse, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/auth.test", nil)
	if err != nil {
		return nil, fmt.Errorf("slack: build auth.test: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+botToken)

	var resp AuthTestResponse
	if err := c.do(httpReq, &resp); err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, fmt.Errorf("slack auth.test: %s", resp.Error)
	}
	return &resp, nil
}

// callJSON POSTs a JSON body to the given Slack method using the bot token.
func (c *Client) callJSON(ctx context.Context, method, botToken string, body, out any) error {
	buf, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("slack: marshal %s: %w", method, err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/"+method, bytes.NewReader(buf))
	if err != nil {
		return fmt.Errorf("slack: build %s: %w", method, err)
	}
	httpReq.Header.Set("Content-Type", "application/json; charset=utf-8")
	httpReq.Header.Set("Authorization", "Bearer "+botToken)
	return c.do(httpReq, out)
}

// do issues an HTTP request and decodes the JSON response into out.
func (c *Client) do(req *http.Request, out any) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("slack: request %s: %w", req.URL.Path, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return fmt.Errorf("slack: read response %s: %w", req.URL.Path, err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("slack %s: http %d: %s", req.URL.Path, resp.StatusCode, strings.TrimSpace(string(body)))
	}

	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("slack %s: decode: %w", req.URL.Path, err)
	}
	return nil
}
