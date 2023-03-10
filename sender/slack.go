package sender

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"go.uber.org/zap"
)

type slackNotifier struct {
	webhookURL string
	l          *zap.Logger
}

var (
	errInvalidURL    error = fmt.Errorf("unspecified or invalid webhook URL")
	errMissingLogger error = fmt.Errorf("missing logger")
)

// NewSlackSender instantiates a new sender that sends to Slack using maybeURL
func NewSlackSender(maybeURL string, l *zap.Logger) (*slackNotifier, error) {
	u, err := url.Parse(maybeURL)
	if err != nil {
		return nil, fmt.Errorf("unable to parse URL: %q: %w", maybeURL, errInvalidURL)
	}

	if u.Scheme != "https" || u.Host != "hooks.slack.com" {
		return nil, fmt.Errorf("invalid scheme or host: %q: %w", maybeURL, errInvalidURL)
	}

	if l == nil {
		return nil, errMissingLogger
	}

	return &slackNotifier{webhookURL: u.String(), l: l}, nil
}

// Send a message via Slack
func (s *slackNotifier) Send(ctx context.Context, msg string) error {
	s.l.Debug("starting to send slack notification", zap.String("msg", msg))
	bs, err := json.Marshal(struct {
		Text string `json:"text"`
	}{
		Text: msg,
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
