package providers

import (
	"net/http"

	"github.com/nuonco/nuon/pkg/events/envelope"
	"github.com/nuonco/nuon/pkg/events/provider"
)

// azureEventGrid speaks the Azure Event Grid webhook protocol: a fixed
// single-event envelope, the subscription validation handshake, and HTTP 400
// rejections as the protocol expects.
type azureEventGrid struct {
	provider.Base
}

func (azureEventGrid) Decoder(provider.EnvelopeType) (envelope.Decoder, error) {
	return envelope.AzureEventGrid{}, nil
}

func (azureEventGrid) Handshake(event *envelope.Event) (*provider.Handshake, error) {
	validationCode, err := envelope.AzureEventGridValidationCode(event)
	if err != nil {
		return nil, err
	}
	if validationCode == "" {
		return nil, nil
	}
	return &provider.Handshake{Status: http.StatusOK, Body: map[string]string{"validationResponse": validationCode}}, nil
}

func (azureEventGrid) RejectStatus(error) int { return http.StatusBadRequest }
