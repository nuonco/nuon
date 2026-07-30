package providers

import (
	"net/http"

	"github.com/nuonco/nuon/pkg/events/envelope"
	"github.com/nuonco/nuon/pkg/events/provider"
	"github.com/nuonco/nuon/pkg/events/signature"
)

// slackEvents speaks the Slack Events API: a fixed envelope, timestamp-bound
// v0 signatures, the url_verification handshake, and HTTP 200 rejections so
// Slack does not retry or disable the subscription.
type slackEvents struct {
	provider.Base
}

func (slackEvents) Decoder(provider.EnvelopeType) (envelope.Decoder, error) {
	return envelope.SlackEvents{}, nil
}

func (slackEvents) Verifier(provider.AuthType, provider.AuthConfig) signature.Verifier {
	return signature.Slack{}
}

func (slackEvents) Handshake(event *envelope.Event) (*provider.Handshake, error) {
	challenge, err := envelope.SlackChallenge(event)
	if err != nil {
		return nil, err
	}
	if challenge == "" {
		return nil, nil
	}
	return &provider.Handshake{Status: http.StatusOK, Body: map[string]string{"challenge": challenge}}, nil
}

func (slackEvents) RejectStatus(error) int { return http.StatusOK }
