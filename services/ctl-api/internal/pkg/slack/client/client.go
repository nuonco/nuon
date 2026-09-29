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

const BaseURL = "https://slack.com/api"

const defaultTimeout = 10 * time.Second

type Client struct {
	httpClient *http.Client
	baseURL    string
}

type Option func(*Client)

func WithBaseURL(u string) Option {
	return func(c *Client) { c.baseURL = u }
}

func WithHTTPClient(h *http.Client) Option {
	return func(c *Client) { c.httpClient = h }
}

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

type baseResponse struct {
	OK    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
}

type OAuthV2AccessRequest struct {
	ClientID     string
	ClientSecret string
	Code         string
	RedirectURI  string
}

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

	IsEnterpriseInstall bool `json:"is_enterprise_install"`

	AuthedUser struct {
		ID string `json:"id"`
	} `json:"authed_user"`
}

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

type PostMessageRequest struct {
	Channel  string         `json:"channel"`
	Text     string         `json:"text,omitempty"`
	Blocks   []slack.Block  `json:"blocks,omitempty"`
	ThreadTS string         `json:"thread_ts,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

type PostMessageResponse struct {
	baseResponse

	Channel string `json:"channel"`
	TS      string `json:"ts"`
}

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

type UpdateMessageRequest struct {
	Channel string        `json:"channel"`
	TS      string        `json:"ts"`
	Text    string        `json:"text,omitempty"`
	Blocks  []slack.Block `json:"blocks,omitempty"`
}

type UpdateMessageResponse struct {
	baseResponse

	Channel string `json:"channel"`
	TS      string `json:"ts"`
}

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

type ConversationsListRequest struct {
	Cursor          string
	Limit           int
	Types           string
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

type ConversationsListResponse struct {
	baseResponse

	Channels         []Conversation `json:"channels"`
	ResponseMetadata struct {
		NextCursor string `json:"next_cursor"`
	} `json:"response_metadata"`
}

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

type ViewsOpenRequest struct {
	TriggerID string         `json:"trigger_id"`
	View      map[string]any `json:"view"`
}

type ViewsResponse struct {
	baseResponse

	View struct {
		ID         string `json:"id"`
		CallbackID string `json:"callback_id"`
		Hash       string `json:"hash"`
	} `json:"view"`

	ResponseMetadata struct {
		Messages []string `json:"messages,omitempty"`
	} `json:"response_metadata,omitempty"`
}

// why: ViewsOpen opens a new modal anchored to the given trigger_id. Trigger ids
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

type ViewsUpdateRequest struct {
	ViewID     string         `json:"view_id,omitempty"`
	ExternalID string         `json:"external_id,omitempty"`
	Hash       string         `json:"hash,omitempty"`
	View       map[string]any `json:"view"`
}

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

type ViewsPushRequest struct {
	TriggerID string         `json:"trigger_id"`
	View      map[string]any `json:"view"`
}

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

func formatViewMessages(msgs []string) string {
	if len(msgs) == 0 {
		return ""
	}
	return " (" + strings.Join(msgs, "; ") + ")"
}

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
