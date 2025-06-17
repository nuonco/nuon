package sender

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

const (
	defaultIconURL string = "https://github.com/nuonco.png?size=57"
)

type slackNotifier struct {
	webhookURL string
	iconURL    string
	iconEmoji  string
	channel    string
}

var errInvalidURL error = fmt.Errorf("unspecified or invalid webhook URL")

type Option func(*slackNotifier)

// NewSlackSender instantiates a new sender that sends to Slack using maybeURL
func NewSlackSender(maybeURL string, opts ...Option) (*slackNotifier, error) {
	u, err := url.Parse(maybeURL)
	if err != nil {
		return nil, fmt.Errorf("unable to parse URL: %q: %w", maybeURL, errInvalidURL)
	}

	if u.Scheme != "https" || u.Host != "hooks.slack.com" {
		return nil, fmt.Errorf("invalid scheme or host: %q: %w", maybeURL, errInvalidURL)
	}

	obj := &slackNotifier{
		webhookURL: u.String(),
		iconURL:    defaultIconURL,
	}
	for _, opt := range opts {
		opt(obj)
	}

	return obj, nil
}

func WithChannel(channel string) Option {
	return func(ec *slackNotifier) {
		ec.channel = channel
	}
}

func WithIconEmoji(iconEmoji string) Option {
	return func(ec *slackNotifier) {
		ec.iconEmoji = iconEmoji
	}
}

type slackMessageJSON struct {
	Text      string `json:"text"`
	IconURL   string `json:"icon_url"`
	IconEmoji string `json:"icon_emoji,omitempty"`
	Channel   string `json:"channel,omitempty"`
}

// Send a message via Slack
//
// to setup a slack notification
// https://nuonco.slack.com/marketplace/new/A0F7XDUAZ-incoming-webhooks
func (s *slackNotifier) Send(ctx context.Context, msg string) error {
	bs, err := json.Marshal(slackMessageJSON{
		Text:      msg,
		IconURL:   s.iconURL,
		Channel:   s.channel,
		IconEmoji: s.iconEmoji,
	})
	if err != nil {
		return err
	}

	fmt.Println(string(bs))

	req, err := http.NewRequestWithContext(ctx, "POST", s.webhookURL, bytes.NewBuffer(bs))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unsuccessful return status: %d", resp.StatusCode)
	}
	return nil
}
