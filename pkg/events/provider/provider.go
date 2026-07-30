// Package provider defines the interface an inbound event provider
// implements: creation-time defaults, envelope decoding, request
// verification, protocol handshakes, and rejection behavior. Base supplies
// the complete default behavior so plain providers are declarative data and
// protocol-heavy providers override only what differs.
package provider

import (
	"errors"
	"net/http"
	"reflect"

	"github.com/nuonco/nuon/pkg/events/envelope"
	"github.com/nuonco/nuon/pkg/events/signature"
	"github.com/nuonco/nuon/pkg/events/sns"
)

type AuthType string

const (
	AuthNone         AuthType = "none"
	AuthHMAC         AuthType = "hmac"
	AuthAPIKey       AuthType = "api_key"
	AuthBasic        AuthType = "basic"
	AuthBearerJWT    AuthType = "bearer_jwt"
	AuthSNSSignature AuthType = "sns_signature"
)

type EnvelopeType string

const (
	EnvelopeNone        EnvelopeType = "none"
	EnvelopePubSubPush  EnvelopeType = "pubsub_push"
	EnvelopeCloudEvents EnvelopeType = "cloudevents"
	EnvelopeSNS         EnvelopeType = "sns"
)

// ErrUnsupportedEnvelope is returned by Decoder for envelope types the
// provider does not implement.
var ErrUnsupportedEnvelope = errors.New("trigger envelope is not implemented")

// AuthConfig configures how inbound requests authenticate.
type AuthConfig struct {
	Header          string   `json:"header,omitempty"`
	Prefix          string   `json:"prefix,omitempty"`
	Encoding        string   `json:"encoding,omitempty"`
	Algorithm       string   `json:"algorithm,omitempty"`
	Username        string   `json:"username,omitempty"`
	Issuer          string   `json:"issuer,omitempty"`
	Audience        []string `json:"audience,omitempty"`
	TopicARN        string   `json:"topic_arn,omitempty"`
	ExpectedEmail   string   `json:"expected_email,omitempty"`
	ExpectedSubject string   `json:"expected_subject,omitempty"`
}

// CallerField names an AuthConfig field the caller supplies at creation time
// instead of the provider defaults.
type CallerField string

const (
	CallerFieldAudience        CallerField = "audience"
	CallerFieldExpectedEmail   CallerField = "expected_email"
	CallerFieldExpectedSubject CallerField = "expected_subject"
	CallerFieldTopicARN        CallerField = "topic_arn"
)

// Defaults is the declarative creation-time configuration a provider preset
// applies to a trigger.
type Defaults struct {
	Auth         AuthType
	Envelope     EnvelopeType
	AuthConfig   AuthConfig
	TypeFrom     envelope.FieldSelector
	IDFrom       envelope.FieldSelector
	CallerFields []CallerField
	// NativeProtocol marks providers that speak their own webhook protocol:
	// the envelope decoder is fixed and field selectors do not apply.
	NativeProtocol bool
	// CallerSecret marks providers whose signing secret is issued by the
	// external provider and supplied by the caller at creation time. Such
	// secrets are write-only: they cannot be revealed, rotated, or revoked
	// independently of the trigger.
	CallerSecret bool
}

// Handshake is a provider protocol response that ends request processing
// without persisting an event, such as Slack URL verification or Azure Event
// Grid subscription validation.
type Handshake struct {
	Status int
	Body   map[string]string
}

// Provider implements one inbound event source. All methods are complete on
// Base; providers embed Base and override only protocol-specific behavior.
type Provider interface {
	Name() string
	Defaults() Defaults
	// Decoder returns the envelope decoder for the trigger's configured
	// envelope type. Native-protocol providers ignore the envelope type.
	Decoder(envelopeType EnvelopeType) (envelope.Decoder, error)
	// Verifier returns the shared-secret verifier for the trigger's auth
	// mechanism, or nil for mechanisms that do not verify against managed
	// secrets (none, bearer_jwt, sns_signature).
	Verifier(auth AuthType, cfg AuthConfig) signature.Verifier
	// Handshake inspects a decoded event and returns a non-nil response when
	// the request is a protocol handshake rather than an event delivery.
	Handshake(event *envelope.Event) (*Handshake, error)
	// RejectStatus returns the HTTP status for an envelope decode failure.
	// http.StatusAccepted means record the rejection and acknowledge the
	// delivery; other statuses surface the error to the sender.
	RejectStatus(err error) int
}

// Base is the default Provider implementation: config-driven decoding and
// verification, no handshake, and accept-and-record rejection.
type Base struct {
	ProviderName     string
	ProviderDefaults Defaults
}

var _ Provider = Base{}

func (b Base) Name() string { return b.ProviderName }

func (b Base) Defaults() Defaults { return b.ProviderDefaults }

func (b Base) Decoder(envelopeType EnvelopeType) (envelope.Decoder, error) {
	switch envelopeType {
	case EnvelopeNone:
		return envelope.Raw{}, nil
	case EnvelopeCloudEvents:
		return envelope.CloudEvents{}, nil
	case EnvelopePubSubPush:
		return envelope.PubSubPush{}, nil
	case EnvelopeSNS:
		return sns.Decoder{}, nil
	default:
		return nil, ErrUnsupportedEnvelope
	}
}

func (b Base) Verifier(auth AuthType, cfg AuthConfig) signature.Verifier {
	switch auth {
	case AuthHMAC:
		if cfg.Header == "" {
			cfg = AuthConfig{Header: "X-Nuon-Signature", Prefix: "v1=", Algorithm: "sha256", Encoding: "hex"}
		}
		return signature.HMAC{Header: cfg.Header, Prefix: cfg.Prefix, Algorithm: cfg.Algorithm, Encoding: cfg.Encoding}
	case AuthAPIKey:
		return signature.APIKey{Header: cfg.Header, Prefix: cfg.Prefix}
	case AuthBasic:
		return signature.Basic{Username: cfg.Username}
	default:
		return nil
	}
}

func (b Base) Handshake(*envelope.Event) (*Handshake, error) { return nil, nil }

func (b Base) RejectStatus(error) int { return http.StatusAccepted }

// ErrConflictingAuthConfig is returned by ApplyDefaults when the caller
// supplies auth config that contradicts the provider defaults.
var ErrConflictingAuthConfig = errors.New("conflicting auth_config")

// ApplyDefaults resolves a caller-supplied auth config against provider
// defaults: caller fields are adopted, unset basic fields are defaulted, and
// any remaining difference is a conflict.
func ApplyDefaults(defaults Defaults, provided AuthConfig) (AuthConfig, error) {
	desired := defaults.AuthConfig
	for _, field := range defaults.CallerFields {
		switch field {
		case CallerFieldAudience:
			desired.Audience = provided.Audience
		case CallerFieldExpectedEmail:
			desired.ExpectedEmail = provided.ExpectedEmail
		case CallerFieldExpectedSubject:
			desired.ExpectedSubject = provided.ExpectedSubject
		case CallerFieldTopicARN:
			desired.TopicARN = provided.TopicARN
		}
	}
	if provided.Header == "" {
		provided.Header = desired.Header
	}
	if provided.Prefix == "" {
		provided.Prefix = desired.Prefix
	}
	if provided.Encoding == "" {
		provided.Encoding = desired.Encoding
	}
	if provided.Algorithm == "" {
		provided.Algorithm = desired.Algorithm
	}
	if provided.Username == "" {
		provided.Username = desired.Username
	}
	if provided.Issuer == "" {
		provided.Issuer = desired.Issuer
	}
	if !reflect.DeepEqual(provided, desired) {
		return AuthConfig{}, ErrConflictingAuthConfig
	}
	return desired, nil
}
