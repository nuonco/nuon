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
}

var errInvalidURL error = fmt.Errorf("unspecified or invalid webhook URL")

// NewSlackSender instantiates a new sender that sends to Slack using maybeURL
func NewSlackSender(maybeURL string) (*slackNotifier, error) {
	u, err := url.Parse(maybeURL)
	if err != nil {
		return nil, fmt.Errorf("unable to parse URL: %q: %w", maybeURL, errInvalidURL)
	}

	if u.Scheme != "https" || u.Host != "hooks.slack.com" {
		return nil, fmt.Errorf("invalid scheme or host: %q: %w", maybeURL, errInvalidURL)
	}

	return &slackNotifier{
		webhookURL: u.String(),
		iconURL:    defaultIconURL,
	}, nil
}

// Send a message via Slack
func (s *slackNotifier) Send(ctx context.Context, msg string) error {
	bs, err := json.Marshal(struct {
		Text    string `json:"text"`
		IconURL string `json:"icon_url"`
	}{
		Text:    msg,
		IconURL: s.iconURL,
	})
	if err != nil {
		return err
	}

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
